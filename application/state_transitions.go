package application

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
)

func CreateTransition(label string, targetState models.BundleState, allowedSourceStates []models.BundleState, handler TransitionHandler) Transition {
	return Transition{
		Label:               label,
		TargetState:         targetState,
		AllowedSourceStates: allowedSourceStates,
		Handler:             handler,
	}
}

func GetListTransitions() []Transition {
	return []Transition{
		CreateTransition("DRAFT", models.BundleStateDraft, []models.BundleState{"IN_REVIEW", "APPROVED"}, UpdateBundleState),
		CreateTransition("IN_REVIEW", models.BundleStateInReview, []models.BundleState{"DRAFT", "APPROVED"}, UpdateBundleState),
		CreateTransition("APPROVED", models.BundleStateApproved, []models.BundleState{"IN_REVIEW"}, UpdateBundleState),
		CreateTransition("PUBLISHED", models.BundleStatePublished, []models.BundleState{"APPROVED"}, UpdateBundleState),
	}
}

func UpdateBundleState(ctx context.Context, api *StateMachineBundleAPI, r *http.Request, bundle *models.Bundle, targetState models.BundleState) *models.Error {
	bundleContents, err := getBundleContents(ctx, api, bundle)
	if err != nil {
		return err
	}

	updateContentsErr := updateStatesForBundleContents(ctx, r, bundleContents, api, bundle, targetState)
	if updateContentsErr != nil {
		return updateContentsErr
	}

	updateBundleStateError := api.Datastore.UpdateBundleState(ctx, bundle.ID, targetState)
	if err != nil {
		log.Error(ctx, "error updating bundle state", updateBundleStateError, log.Data{models.KeyBundleId: bundle.ID, models.KeyTargetState: targetState})
		return models.CreateModelError(models.CodeInternalServerError, updateBundleStateError.Error())
	}

	updateEventErr := api.Events.InsertBundleUpdatedEvent(ctx, r, bundle)
	if updateEventErr != nil {
		log.Error(ctx, "error creating bundle update event", errors.New(updateEventErr.Description), log.Data{models.KeyBundleId: bundle.ID, models.KeyTargetState: targetState})
		return updateEventErr
	}

	return nil

}
func updateStatesForBundleContents(ctx context.Context, r *http.Request, bundleContents *[]models.ContentItem, api *StateMachineBundleAPI, bundle *models.Bundle, targetState models.BundleState) *models.Error {
	for _, content := range *bundleContents {
		contentErr := tryUpdateVersionStateForContentItem(ctx, r, api, content, bundle, targetState)

		if contentErr != nil {
			return contentErr
		}
	}

	return nil
}

func getBundleContents(ctx context.Context, api *StateMachineBundleAPI, bundle *models.Bundle) (*[]models.ContentItem, *models.Error) {
	bundleContents, err := api.Datastore.GetBundleContentItems(ctx, bundle.ID)

	if err != nil {
		log.Error(ctx, "error getting bundle content items", err, log.Data{models.KeyBundleId: bundle.ID})
		return nil, models.CreateModelError(models.CodeInternalServerError, err.Error())
	}

	if len(bundleContents) == 0 {
		log.Warn(ctx, "found no content items for bundle", log.Data{models.KeyBundleId: bundle.ID})
		return nil, models.CreateModelError(models.CodeNotFound, apierrors.ErrorDescriptionNoContentItemsFound)
	}
	return &bundleContents, nil
}

func tryUpdateVersionStateForContentItem(ctx context.Context, r *http.Request, api *StateMachineBundleAPI, content models.ContentItem, bundle *models.Bundle, targetState models.BundleState) *models.Error {
	if !strings.EqualFold(string(*content.State), string(*bundle.State)) {
		log.Warn(ctx, "skipping updating content item due to mismatched state", log.Data{models.KeyBundleId: bundle.ID, models.KeyContentItemState: *content.State, models.KeyBundleState: *bundle.State})
		return nil
	}

	version, err := api.DatasetsClient.Versions().GetForContentItem(ctx, r, content)

	if err != nil {
		log.Error(ctx, "error getting version for content item", err, log.Data{models.KeyBundleId: bundle.ID, models.KeyContentItemId: content.ID})
		return models.CreateModelError(models.CodeInternalServerError, err.Error())
	}

	if !strings.EqualFold(string(version.State), string(*bundle.State)) {
		log.Warn(ctx, "skipping updating version state due to mismatched state", log.Data{models.KeyBundleId: bundle.ID, models.KeyContentItemState: *content.State, models.KeyBundleState: *bundle.State, "version_id": version.ID})
		return nil
	}

	err = api.DatasetsClient.Versions().UpdateStateForContentItem(ctx, r, content, targetState)

	if err != nil {
		log.Error(ctx, "error updating version state for content item", err, log.Data{models.KeyBundleId: bundle.ID, models.KeyContentItemId: content.ID, models.KeyTargetState: targetState})
		return models.CreateModelError(models.CodeInternalServerError, err.Error())
	}

	err = api.Datastore.UpdateBundleContentItemState(ctx, content.ID, targetState)
	if err != nil {
		log.Error(ctx, "error updating content item state", err, log.Data{models.KeyBundleId: bundle.ID, models.KeyContentItemId: content.ID, models.KeyTargetState: targetState})
		return models.CreateModelError(models.CodeInternalServerError, err.Error())
	}

	return nil
}
