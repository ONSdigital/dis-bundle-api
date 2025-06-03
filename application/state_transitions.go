package application

import (
	"context"
	"strings"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
)

func GetTransitionHandlers() []StateMachineTransition {
	stateMachineTransitions := []StateMachineTransition{{
		sourceState:       models.BundleStateInReview,
		targetState:       models.BundleStateApproved,
		transitionHandler: updateBundleState,
	},
		{
			sourceState:       models.BundleStateApproved,
			targetState:       models.BundleStatePublished,
			transitionHandler: updateBundleState,
		},
	}

	return stateMachineTransitions
}

func updateBundleState(ctx context.Context, api *StateMachineBundleAPI, bundle *models.Bundle, targetState models.BundleState) *models.Error {
	bundleContents, err := getBundleContents(ctx, api, bundle)
	if err != nil {
		return err
	}

	updateContentsErr := updateStatesForBundleContents(ctx, bundleContents, api, bundle, targetState)
	if updateContentsErr != nil {
		return updateContentsErr
	}

	updateBundleStateError := api.Datastore.UpdateBundleState(ctx, bundle.ID, targetState)
	if err != nil {
		log.Error(ctx, "error updating bundle state", updateBundleStateError, log.Data{models.KeyBundleId: bundle.ID, models.KeyTargetState: targetState})
		return models.CreateModelError(models.CodeInternalServerError, updateBundleStateError.Error())
	}

	return nil

}
func updateStatesForBundleContents(ctx context.Context, bundleContents *[]models.ContentItem, api *StateMachineBundleAPI, bundle *models.Bundle, targetState models.BundleState) *models.Error {
	for _, content := range *bundleContents {
		contentErr := tryUpdateVersionStateForContentItem(ctx, api, content, bundle, targetState)

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

func tryUpdateVersionStateForContentItem(ctx context.Context, api *StateMachineBundleAPI, content models.ContentItem, bundle *models.Bundle, targetState models.BundleState) *models.Error {
	if !strings.EqualFold(string(*content.State), string(*bundle.State)) {
		log.Warn(ctx, "skipping updating content item due to mismatched state", log.Data{models.KeyBundleId: bundle.ID, models.KeyContentItemState: *content.State, models.KeyBundleState: *bundle.State})
		return nil
	}

	version, err := api.DatasetsClient.Versions().GetForContentItem(ctx, content)

	if err != nil {
		log.Error(ctx, "error getting version for content item", err, log.Data{models.KeyBundleId: bundle.ID, models.KeyContentItemId: content.ID})
		return models.CreateModelError(models.CodeInternalServerError, err.Error())
	}

	if !strings.EqualFold(string(version.State), string(*bundle.State)) {
		log.Warn(ctx, "skipping updating version state due to mismatched state", log.Data{models.KeyBundleId: bundle.ID, models.KeyContentItemState: *content.State, models.KeyBundleState: *bundle.State, "version_id": version.ID})
		return nil
	}

	err = api.DatasetsClient.Versions().UpdateStateForContentItem(ctx, content, targetState)

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
