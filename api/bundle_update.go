package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

func (api *BundleAPI) putBundle(w http.ResponseWriter, r *http.Request) {
	defer dphttp.DrainBody(r)

	ctx := r.Context()
	vars := mux.Vars(r)
	bundleID := vars["bundle-id"]

	logdata := log.Data{"bundle_id": bundleID}

	ifMatchHeader := r.Header.Get("If-Match")
	if ifMatchHeader == "" {
		log.Error(ctx, "putBundle endpoint: missing If-Match header", nil, logdata)
		code := models.CodeConflict
		errInfo := &models.Error{
			Code:        &code,
			Description: "Change rejected due to a conflict with the current resource state. A common cause is attempted to change a bundle that is already locked pending publication or has already been published.",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusConflict, errInfo)
		return
	}

	JWTData, err := api.authMiddleware.Parse(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if err != nil {
		api.handleInternalError(ctx, w, r, "failed to parse JWT from authorization header", err, logdata)
		return
	}

	currentBundle, err := api.stateMachineBundleAPI.GetBundle(ctx, bundleID)
	if err != nil {
		api.handleGetBundleError(ctx, w, r, err, logdata)
		return
	}

	trimmedIfMatch := strings.Trim(ifMatchHeader, "\"")
	trimmedDBETag := strings.Trim(currentBundle.ETag, "\"")

	if trimmedDBETag != trimmedIfMatch {
		log.Error(ctx, "putBundle endpoint: ETag mismatch", nil, logdata)
		code := models.CodeConflict
		errInfo := &models.Error{
			Code:        &code,
			Description: "Change rejected due to a conflict with the current resource state. A common cause is attempted to change a bundle that is already locked pending publication or has already been published.",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusConflict, errInfo)
		return
	}

	bundleUpdate, validationErrors, err := api.createAndValidateBundleUpdate(r, bundleID, currentBundle, JWTData.UserID)
	if err != nil {
		api.handleBadRequestError(ctx, w, r, "bundle creation or validation failed", err, logdata)
		return
	}
	if len(validationErrors) > 0 {
		logdata["validation_errors"] = validationErrors
		log.Error(ctx, "putBundle endpoint: bundle validation failed", nil, logdata)
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, validationErrors...)
		return
	}

	validationErrs := validateBundleRules(ctx, bundleUpdate, currentBundle, api)
	if len(validationErrs) > 0 {
		logdata["validation_errors"] = validationErrs
		log.Error(ctx, "putBundle endpoint: bundle business rule validation failed", nil, logdata)
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, validationErrs...)
		return
	}

	stateChangingToPublished, err := api.handleStateTransition(ctx, bundleUpdate, currentBundle)
	if err != nil {
		log.Error(ctx, "putBundle endpoint: invalid state transition", err, logdata)
		code := models.ErrInvalidParameters
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionMalformedRequest,
			Source:      &models.Source{Field: "/state"},
		}
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
		return
	}

	now := time.Now()
	bundleUpdate.UpdatedAt = &now
	bundleUpdate.LastUpdatedBy = &models.User{Email: JWTData.UserID}

	_, err = api.stateMachineBundleAPI.UpdateBundle(ctx, bundleID, bundleUpdate)
	if err != nil {
		api.handleInternalError(ctx, w, r, "failed to update bundle in database", err, logdata)
		return
	}

	updatedBundle, err := api.stateMachineBundleAPI.UpdateBundleETag(ctx, bundleID, JWTData.UserID)
	if err != nil {
		api.handleInternalError(ctx, w, r, "failed to update bundle ETag", err, logdata)
		return
	}

	if stateChangingToPublished {
		var authHeaders datasetAPISDK.Headers
		if r.Header.Get("X-Florence-Token") != "" {
			authHeaders.ServiceToken = r.Header.Get("X-Florence-Token")
		} else {
			authHeaders.ServiceToken = r.Header.Get("Authorization")
		}

		err = updateContentItemsWithDatasetInfo(ctx, api, bundleID, authHeaders)
		if err != nil {
			log.Error(ctx, "putBundle endpoint: failed to update content items with dataset info", err, logdata)
		}
	}

	event := &models.Event{
		RequestedBy: &models.RequestedBy{
			ID:    JWTData.UserID,
			Email: JWTData.UserID,
		},
		Action:   models.ActionUpdate,
		Resource: "/bundles/" + bundleID,
		Bundle: &models.EventBundle{
			ID:            updatedBundle.ID,
			BundleType:    updatedBundle.BundleType,
			CreatedBy:     updatedBundle.CreatedBy,
			CreatedAt:     updatedBundle.CreatedAt,
			LastUpdatedBy: updatedBundle.LastUpdatedBy,
			PreviewTeams:  updatedBundle.PreviewTeams,
			ScheduledAt:   updatedBundle.ScheduledAt,
			State:         updatedBundle.State,
			Title:         updatedBundle.Title,
			UpdatedAt:     updatedBundle.UpdatedAt,
			ManagedBy:     updatedBundle.ManagedBy,
		},
	}

	err = models.ValidateEvent(event)
	if err != nil {
		api.handleInternalError(ctx, w, r, "event validation failed", err, logdata)
		return
	}

	err = api.stateMachineBundleAPI.CreateBundleEvent(ctx, event)
	if err != nil {
		api.handleInternalError(ctx, w, r, "failed to create event in database", err, logdata)
		return
	}

	bundleJSON, err := json.Marshal(updatedBundle)
	if err != nil {
		api.handleInternalError(ctx, w, r, "failed to marshal bundle to JSON", err, logdata)
		return
	}

	dpresponse.SetETag(w, updatedBundle.ETag)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(bundleJSON); err != nil {
		log.Error(ctx, "putBundle endpoint: error writing response body", err, logdata)
	}
}

// Helper function to handle state transitions
func (api *BundleAPI) handleStateTransition(ctx context.Context, bundleUpdate, currentBundle *models.Bundle) (bool, error) {
	stateChangingToPublished := bundleUpdate.State != "" && currentBundle.State != "" &&
		bundleUpdate.State == models.BundleStatePublished && currentBundle.State != models.BundleStatePublished

	if bundleUpdate.State != "" && currentBundle.State != "" && bundleUpdate.State != currentBundle.State {
		err := api.stateMachineBundleAPI.StateMachine.Transition(ctx, api.stateMachineBundleAPI, currentBundle, bundleUpdate)
		if err != nil {
			return false, err
		}
	}

	return stateChangingToPublished, nil
}

// Helper function for bad request errors
func (api *BundleAPI) handleBadRequestError(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, logdata log.Data) {
	log.Error(ctx, "putBundle endpoint: "+message, err, logdata)
	code := models.ErrInvalidParameters
	errInfo := &models.Error{
		Code:        &code,
		Description: apierrors.ErrorDescriptionMalformedRequest,
	}
	utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
}

// Helper function to create and validate bundle update
func (api *BundleAPI) createAndValidateBundleUpdate(r *http.Request, bundleID string, currentBundle *models.Bundle, email string) (*models.Bundle, []*models.Error, error) {
	bundleUpdate, err := models.CreateBundle(r.Body, email)
	if err != nil {
		return nil, nil, err
	}

	bundleUpdate.ID = bundleID
	bundleUpdate.CreatedAt = currentBundle.CreatedAt
	bundleUpdate.CreatedBy = currentBundle.CreatedBy

	models.CleanBundle(bundleUpdate)

	validationErrors := models.ValidateBundle(bundleUpdate)
	if len(validationErrors) > 0 {
		return nil, validationErrors, nil
	}

	return bundleUpdate, nil, nil
}

func (api *BundleAPI) handleInternalError(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, logdata log.Data) {
	log.Error(ctx, "putBundle endpoint: "+message, err, logdata)
	code := models.InternalError
	errInfo := &models.Error{
		Code:        &code,
		Description: apierrors.ErrorDescriptionInternalError,
	}
	utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
}

// Helper function for GetBundle error handling
func (api *BundleAPI) handleGetBundleError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error, logdata log.Data) {
	if err == apierrors.ErrBundleNotFound {
		log.Error(ctx, "putBundle endpoint: bundle not found", err, logdata)
		code := models.NotFound
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionNotFound,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
		return
	}
	api.handleInternalError(ctx, w, r, "failed to get bundle", err, logdata)
}

// validateBundleRules validates the rules for bundle updates
func validateBundleRules(ctx context.Context, bundleUpdate, currentBundle *models.Bundle, api *BundleAPI) []*models.Error {
	var validationErrors []*models.Error

	if bundleUpdate.Title != currentBundle.Title {
		exists, err := api.stateMachineBundleAPI.CheckBundleExistsByTitleUpdate(ctx, bundleUpdate.Title, bundleUpdate.ID)
		if err != nil {
			log.Error(ctx, "failed to check bundle title uniqueness", err)
			code := models.InternalError
			validationErrors = append(validationErrors, &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionInternalError,
			})
			return validationErrors
		}
		if exists {
			code := models.ErrInvalidParameters
			validationErrors = append(validationErrors, &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionMalformedRequest,
				Source:      &models.Source{Field: "/title"},
			})
		}
	}

	if bundleUpdate.BundleType == models.BundleTypeScheduled {
		if bundleUpdate.ScheduledAt == nil {
			code := models.ErrInvalidParameters
			validationErrors = append(validationErrors, &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionMalformedRequest,
				Source:      &models.Source{Field: "/scheduled_at"},
			})
		} else if bundleUpdate.ScheduledAt.Before(time.Now()) {
			code := models.ErrInvalidParameters
			validationErrors = append(validationErrors, &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionMalformedRequest,
				Source:      &models.Source{Field: "/scheduled_at"},
			})
		}
	}

	if bundleUpdate.BundleType == models.BundleTypeManual && bundleUpdate.ScheduledAt != nil {
		code := models.ErrInvalidParameters
		validationErrors = append(validationErrors, &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionMalformedRequest,
			Source:      &models.Source{Field: "/scheduled_at"},
		})
	}

	return validationErrors
}

func updateContentItemsWithDatasetInfo(ctx context.Context, api *BundleAPI, bundleID string, authHeaders datasetAPISDK.Headers) error {
	contentItems, err := api.stateMachineBundleAPI.GetContentItemsByBundleID(ctx, bundleID)
	if err != nil {
		log.Error(ctx, "failed to get content items", err, log.Data{"bundle_id": bundleID})
		return err
	}

	if len(contentItems) == 0 {
		log.Info(ctx, "no content items found for bundle", log.Data{"bundle_id": bundleID})
		return nil
	}

	for _, contentItem := range contentItems {
		log.Info(ctx, "calling dataset API client", log.Data{
			"dataset_id":                 contentItem.Metadata.DatasetID,
			"auth_headers_service_token": authHeaders.ServiceToken != "",
			"client_debug":               fmt.Sprintf("%+v", api.stateMachineBundleAPI.DatasetAPIClient),
		})

		dataset, err := api.stateMachineBundleAPI.DatasetAPIClient.GetDataset(ctx, authHeaders, "", contentItem.Metadata.DatasetID)
		if err != nil {
			log.Error(ctx, "dataset api client call failed", err, log.Data{
				"content_item_id": contentItem.ID,
				"dataset_id":      contentItem.Metadata.DatasetID,
			})
			continue
		}

		err = api.stateMachineBundleAPI.UpdateContentItemDatasetInfo(ctx, contentItem.ID, dataset.Title, dataset.State)
		if err != nil {
			log.Error(ctx, "update content item failed", err, log.Data{
				"content_item_id": contentItem.ID,
				"dataset_id":      contentItem.Metadata.DatasetID,
			})
			continue
		}
	}
	return nil
}
