package application

import (
	"context"
	"errors"
	"slices"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
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
	Name      string
	EnterFunc func(ctx context.Context, sm *StateMachine, smAPI *StateMachineBundleAPI, bundle *models.Bundle, auth *models.AuthEntityData) error
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

func (sm *StateMachine) Transition(ctx context.Context, smAPI *StateMachineBundleAPI, currentBundle, bundleUpdate *models.Bundle) error {
	var valid bool
	match := false

	if currentBundle == nil {
		if bundleUpdate.State.String() == models.BundleStateDraft.String() {
			return nil
		}
		return errors.New("bundle state must be DRAFT when creating a new bundle")
	}

	if bundleUpdate == nil {
		if currentBundle.State.String() == models.BundleStatePublished.String() {
			return errors.New("cannot update a published bundle")
		}
		return nil
	}

	for state, transitions := range sm.transitions {
		if state == bundleUpdate.State.String() {
			for _, allowed := range transitions {
				if currentBundle.State.String() != allowed {
					continue
				}
				match = true
				_, valid = getStateByName(state)
				if !valid {
					return errors.New("incorrect state value")
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

func (sm *StateMachine) TransitionBundle(ctx context.Context, smAPI *StateMachineBundleAPI, bundle *models.Bundle, targetState *models.BundleState, auth *models.AuthEntityData) (*models.Bundle, error) {
	// Validate transition
	if err := sm.IsValidTransition(ctx, &bundle.State, targetState); err != nil {
		return nil, err
	}

	// Update bundle state + user
	bundle.State = *targetState
	bundle.LastUpdatedBy.Email = auth.GetUserEmail()

	updatedBundle, err := smAPI.Datastore.UpdateBundle(ctx, bundle.ID, bundle)
	if err != nil {
		return nil, err
	}

	// Call the EnterFunc for the new state
	state, ok := getStateByName(targetState.String())
	if !ok {
		return nil, errors.New("target state not found")
	}
	if state.EnterFunc != nil {
		if err := state.EnterFunc(ctx, sm, smAPI, updatedBundle, auth); err != nil {
			return nil, err
		}
	}

	// Create bundle event
	event, err := models.CreateEventModel(auth.GetUserID(), auth.GetUserEmail(), models.ActionUpdate, models.CreateBundleResourceLocation(updatedBundle), nil, updatedBundle)
	if err != nil {
		return nil, err
	}
	if err := smAPI.CreateBundleEvent(ctx, event); err != nil {
		return nil, err
	}

	return updatedBundle, nil
}

func (*StateMachine) transitionContentItem(ctx context.Context, contentItem *models.ContentItem, smAPI *StateMachineBundleAPI, targetState models.BundleState, auth *models.AuthEntityData) error {
	if err := smAPI.updateVersionStateForContentItem(ctx, contentItem, &targetState, auth.Headers); err != nil {
		return err
	}

	if err := smAPI.Datastore.UpdateContentItemState(ctx, contentItem.ID, targetState.String()); err != nil {
		return err
	}

	event, err := models.CreateEventModel(auth.GetUserID(), auth.GetUserEmail(), models.ActionUpdate, models.CreateBundleContentResourceLocation(contentItem), contentItem, nil)
	if err != nil {
		return err
	}

	if err := smAPI.CreateBundleEvent(ctx, event); err != nil {
		return err
	}

	return nil
}
