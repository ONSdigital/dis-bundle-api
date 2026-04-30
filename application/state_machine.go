package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/slack"
	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/ONSdigital/log.go/v2/log"
)

type StateMachine struct {
	states      map[string]State
	transitions map[string][]string
	datastore   store.Datastore
	ctx         context.Context
}

type Transition struct {
	Label               string
	TargetState         State
	AllowedSourceStates []string
}

type State struct {
	Name string
}

func (s State) String() string {
	return s.Name
}

func getStateByName(stateName string) (*State, bool) {
	switch stateName {
	case "DRAFT":
		return &Draft, true
	case "IN_REVIEW":
		return &InReview, true
	case "APPROVED":
		return &Approved, true
	case "PUBLISHED":
		return &Published, true
	default:
		return nil, false
	}
}

func NewStateMachine(ctx context.Context, states []State, transitions []Transition, datastore store.Datastore) *StateMachine {
	statesMap := make(map[string]State)
	for _, state := range states {
		statesMap[state.String()] = state
	}

	transitionsMap := make(map[string][]string)
	for _, transition := range transitions {
		transitionsMap[transition.TargetState.String()] = transition.AllowedSourceStates
	}

	StateMachine := &StateMachine{
		states:      statesMap,
		transitions: transitionsMap,
		datastore:   datastore,
		ctx:         ctx,
	}

	return StateMachine
}

func (sm *StateMachine) Transition(ctx context.Context, stateMachineBundleAPI *StateMachineBundleAPI, currentBundle, bundleUpdate *models.Bundle) error {
	var valid bool

	match := false

	if currentBundle == nil {
		if bundleUpdate.State.String() == models.BundleStateDraft.String() {
			return nil
		} else {
			return errors.New("bundle state must be DRAFT when creating a new bundle")
		}
	}

	if bundleUpdate == nil {
		if currentBundle.State.String() == models.BundleStatePublished.String() {
			return errors.New("cannot update a published bundle")
		}
		return nil
	}

	for state, transitions := range sm.transitions {
		if state == bundleUpdate.State.String() {
			for i := range transitions {
				if currentBundle.State.String() != transitions[i] {
					continue
				}
				match = true

				_, valid = getStateByName(state)
				if !valid {
					return errors.New("incorrect state value")
				}

				if currentBundle.State.String() == InReview.String() && bundleUpdate.State.String() == Approved.String() {
					allBundleContentsApproved, err := stateMachineBundleAPI.CheckAllBundleContentsAreApproved(ctx, currentBundle.ID)
					if err != nil {
						log.Error(ctx, "error checking if all bundle contents are approved", err, log.Data{"bundle_id": currentBundle.ID})
						return err
					}

					if !allBundleContentsApproved {
						return errors.New("not all bundle contents are approved")
					}
				}
				break
			}
		}
	}

	if !match {
		return apierrors.ErrInvalidTransition
	}

	return nil
}

// IsValidTransition validates whether the sourceState can transition to the targetState. If not, an error is returned
func (sm *StateMachine) IsValidTransition(ctx context.Context, sourceState, targetState *models.BundleState) error {
	allowedSourceStates, exists := sm.transitions[targetState.String()]

	if !exists {
		return apierrors.ErrInvalidTransition
	}

	if !slices.Contains(allowedSourceStates, sourceState.String()) {
		return apierrors.ErrInvalidTransition
	}

	return nil
}

func (sm *StateMachine) TransitionBundle(ctx context.Context, stateMachineBundleAPI *StateMachineBundleAPI, bundle *models.Bundle, targetState *models.BundleState, authEntityData *models.AuthEntityData) (*models.Bundle, bool, error) {
	if err := sm.IsValidTransition(ctx, &bundle.State, targetState); err != nil {
		return nil, false, err
	}

	contents, err := stateMachineBundleAPI.Datastore.GetBundleContentsForBundle(ctx, bundle.ID)
	if err != nil {
		return nil, false, err
	}

	if contents == nil || len(*contents) == 0 {
		return nil, false, apierrors.ErrBundleHasNoContentItems
	}

	hadContentItemFailures := false

	if targetState.String() == models.BundleStateApproved.String() || targetState.String() == models.BundleStatePublished.String() {
		for index := range *contents {
			contentItem := &(*contents)[index]
			err = sm.transitionContentItem(ctx, contentItem, stateMachineBundleAPI, targetState, authEntityData)
			if err != nil {
				log.Warn(ctx, fmt.Sprintf("Error occurred transitioning content item for bundle: %s", err.Error()), log.Data{"bundle-id": bundle.ID, "content-item-id": contentItem.ID})

				previewURL := stateMachineBundleAPI.PreviewServiceURL + contentItem.Links.Preview

				alarmFields := []slack.Field{
					{Title: "Bundle ID", Value: bundle.ID},
					{Title: "Bundle Title", Value: bundle.Title},
					{Title: "Dataset ID", Value: contentItem.Metadata.DatasetID},
					{Title: "Edition", Value: contentItem.Metadata.EditionID},
					{Title: "Version", Value: strconv.Itoa(contentItem.Metadata.VersionID)},
					{Title: "Preview Link", Value: previewURL},
				}

				_, alarmErr := stateMachineBundleAPI.DataBundleSlackClient.SendAlarm(ctx, "Bundle content item failed to publish", err, alarmFields)
				if alarmErr != nil {
					log.Error(ctx, "failed to send slack alarm for content item failure", alarmErr, log.Data{"bundle-id": bundle.ID, "content-item-id": contentItem.ID})
				}

				log.Info(ctx, "sending slack alarm for content item failure", log.Data{
					"bundle-id":       bundle.ID,
					"content-item-id": contentItem.ID,
					"alarm_fields":    alarmFields,
				})

				hadContentItemFailures = true
				continue
			}
		}
	}

	bundle.State = *targetState
	bundle.LastUpdatedBy.Email = authEntityData.GetUserEmail()

	updatedBundle, err := stateMachineBundleAPI.Datastore.UpdateBundle(ctx, bundle.ID, bundle)
	if err != nil {
		return nil, hadContentItemFailures, err
	}

	identityType := log.USER
	if authEntityData.IsServiceAuth {
		identityType = log.SERVICE
	}
	logAuth := log.Auth(identityType, authEntityData.EntityData.UserID)

	if err = stateMachineBundleAPI.CreateEvent(ctx, authEntityData, models.ActionUpdate, updatedBundle, nil); err != nil {
		log.Error(ctx, "failed to create event", err, log.Classification(log.ProtectiveMonitoring), logAuth, log.Data{"bundle_id": updatedBundle.ID, "action": models.ActionUpdate})
		return nil, hadContentItemFailures, err
	}
	log.Info(ctx, "bundle event creation successful", log.Classification(log.ProtectiveMonitoring), logAuth, log.Data{"bundle_id": updatedBundle.ID, "action": models.ActionUpdate})

	return updatedBundle, hadContentItemFailures, nil
}

func (*StateMachine) transitionContentItem(ctx context.Context, contentItem *models.ContentItem, stateMachineBundleAPI *StateMachineBundleAPI, targetState *models.BundleState, authEntityData *models.AuthEntityData) error {
	if err := stateMachineBundleAPI.updateVersionStateForContentItem(ctx, contentItem, targetState, authEntityData.Headers); err != nil {
		return err
	}

	if err := stateMachineBundleAPI.Datastore.UpdateContentItemState(ctx, contentItem.ID, targetState.String()); err != nil {
		return err
	}

	identityType := log.USER
	if authEntityData.IsServiceAuth {
		identityType = log.SERVICE
	}
	logAuth := log.Auth(identityType, authEntityData.EntityData.UserID)

	if err := stateMachineBundleAPI.CreateEvent(ctx, authEntityData, models.ActionUpdate, nil, contentItem); err != nil {
		log.Error(ctx, "failed to create event", err, log.Classification(log.ProtectiveMonitoring), logAuth, log.Data{"bundle_id": contentItem.BundleID, "content_item_id": contentItem.ID, "action": models.ActionUpdate})
		return err
	}
	log.Info(ctx, "bundle event creation successful", log.Classification(log.ProtectiveMonitoring), logAuth, log.Data{"bundle_id": contentItem.BundleID, "content_item_id": contentItem.ID, "action": models.ActionUpdate})
	return nil
}
