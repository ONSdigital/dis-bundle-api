package api

import (
	"net/http"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	"github.com/ONSdigital/log.go/v2/log"
)

func (api *BundleAPI) getBundles(w http.ResponseWriter, r *http.Request, limit, offset int) (bundles any, errCode int, errBundles *models.Error) {
	ctx := r.Context()

	bundles, totalCount, err := api.stateMachineBundleAPI.ListBundles(ctx, offset, limit)
	if err != nil {
		code := models.CodeInternalServerError
		log.Error(ctx, "failed to get bundles", err)
		return nil, http.StatusInternalServerError, &models.Error{Code: &code, Description: "Failed to process the request due to an internal error"}
	}

	return bundles, totalCount, nil
}

func (api *BundleAPI) createBundle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.Info(ctx, "createBundle: creating a new bundle")

	bundle, err := models.CreateBundle(r.Body)
	if err != nil {
		code := models.CodeBadRequest
		log.Error(ctx, "failed to create bundle from request body", err)
		utils.HandleBundleAPIErr(w, r, &models.Error{
			Code:        &code,
			Description: "Invalid request body",
		}, http.StatusBadRequest)
		return
	}

	log.Info(ctx, "createBundle: created bundle from request body", log.Data{"bundle": bundle})

	err = models.ValidateBundle(bundle)
	if err != nil {
		code := models.CodeBadRequest
		log.Error(ctx, "failed to validate bundle", err)
		utils.HandleBundleAPIErr(w, r, &models.Error{
			Code:        &code,
			Description: "Invalid bundle data",
		}, http.StatusBadRequest)
		return
	}

	err = api.stateMachineBundleAPI.StateMachine.Transition(ctx, api.stateMachineBundleAPI, nil, bundle)
	if err != nil {
		code := models.CodeBadRequest
		log.Error(ctx, "failed to transition bundle state", err)
		utils.HandleBundleAPIErr(w, r, &models.Error{
			Code:        &code,
			Description: "Failed to transition bundle state",
		}, http.StatusBadRequest)
		return
	}

	_, err = api.stateMachineBundleAPI.GetBundleByTitle(ctx, bundle.Title)
	if err != nil {
		if err != apierrors.ErrBundleNotFound {
			code := models.CodeInternalServerError
			log.Error(ctx, "failed to check existing bundle by title", err)
			utils.HandleBundleAPIErr(w, r, &models.Error{
				Code:        &code,
				Description: "Failed to check existing bundle by title",
			}, http.StatusInternalServerError)
			return
		}
	} else {
		code := models.CodeConflict
		log.Error(ctx, "bundle with the same title already exists", nil)
		utils.HandleBundleAPIErr(w, r, &models.Error{
			Code:        &code,
			Description: "A bundle with the same title already exists",
		}, http.StatusConflict)
		return
	}

	if bundle.BundleType == models.BundleTypeScheduled && bundle.ScheduledAt == nil {
		code := models.CodeBadRequest
		log.Error(ctx, "scheduled_at is required for scheduled bundles", nil)
		utils.HandleBundleAPIErr(w, r, &models.Error{
			Code:        &code,
			Description: "scheduled_at is required for scheduled bundles",
		}, http.StatusBadRequest)
		return
	}

	if bundle.BundleType == models.BundleTypeManual && bundle.ScheduledAt != nil {
		code := models.CodeBadRequest
		log.Error(ctx, "scheduled_at should not be set for manual bundles", nil)
		utils.HandleBundleAPIErr(w, r, &models.Error{
			Code:        &code,
			Description: "scheduled_at should not be set for manual bundles",
		}, http.StatusBadRequest)
		return
	}

	if bundle.ScheduledAt != nil && bundle.ScheduledAt.Before(time.Now()) {
		code := models.CodeBadRequest
		log.Error(ctx, "scheduled_at cannot be in the past", nil)
		utils.HandleBundleAPIErr(w, r, &models.Error{
			Code:        &code,
			Description: "scheduled_at cannot be in the past",
		}, http.StatusBadRequest)
		return
	}
}
