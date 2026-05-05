package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
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

	logData := log.Data{"bundle_id": bundleID}

	authEntityData, err := api.GetAuthEntityData(r)
	if err != nil {
		handleErr(ctx, w, r, err, logData, RouteNamePutBundle)
		return
	}

	etag, err := utils.GetETag(r)
	if err != nil {
		handleErr(ctx, w, r, err, logData, RouteNamePutBundle)
		return
	}

	bundleUpdate, validationErrors, err := api.CreateAndValidateBundleUpdate(r, bundleID, authEntityData.GetUserID())
	if err != nil {
		api.handleBadRequestError(ctx, w, r, "bundle creation or validation failed", err, logData)
		return
	}

	var allValidationErrors []*models.Error
	allValidationErrors = append(allValidationErrors, validationErrors...)

	//validationErrs := api.stateMachineBundleAPI.ValidateBundleRules(ctx, bundleUpdate, currentBundle)
	//allValidationErrors = append(allValidationErrors, validationErrs...)

	// if bundleUpdate.State != "" && bundleUpdate.State.IsValid() && currentBundle.State != "" && bundleUpdate.State != currentBundle.State {
	// 	err := api.stateMachineBundleAPI.StateMachine.Transition(context.Background(), api.stateMachineBundleAPI, currentBundle, bundleUpdate)
	// 	if err != nil {
	// 		if err == apierrors.ErrInvalidTransition {
	// 			code := models.CodeInvalidParameters
	// 			stateError := &models.Error{
	// 				Code:        &code,
	// 				Description: apierrors.ErrorDescriptionMalformedRequest,
	// 				Source:      &models.Source{Field: "/state"},
	// 			}
	// 			allValidationErrors = append(allValidationErrors, stateError)
	// 		} else {
	// 			api.handleInternalError(ctx, w, r, "state transition validation failed", err, logdata)
	// 			return
	// 		}
	// 	}
	// }

	if len(allValidationErrors) > 0 {
		logData["validation_errors"] = allValidationErrors
		log.Error(ctx, "putBundle endpoint: bundle validation failed", nil, logData)
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, allValidationErrors...)
		return
	}

	// Create policies for any preview teams added in the update.
	// NOTE: This does not currently handle the case where existing preview teams are removed.
	// If a preview team is removed from the bundle, their policies will still exist.
	if err := api.stateMachineBundleAPI.CreateBundlePolicies(ctx, authEntityData.Headers.AccessToken, bundleUpdate.PreviewTeams, models.RoleDatasetsPreviewer); err != nil {
		api.handleInternalError(ctx, w, r, "failed to create bundle policies", err, logData)
		return
	}

	updatedBundle, err := api.stateMachineBundleAPI.PutBundle(ctx, bundleID, bundleUpdate, authEntityData, *etag)
	if err != nil {
		log.Error(ctx, "putBundle endpoint: bundle update failed", err, logData)
		handleErr(ctx, w, r, err, logData, RouteNamePutBundle)
		// switch err {
		// case apierrors.ErrInvalidTransition:
		// 	code := models.CodeInvalidParameters
		// 	errInfo := &models.Error{
		// 		Code:        &code,
		// 		Description: apierrors.ErrorDescriptionMalformedRequest,
		// 		Source:      &models.Source{Field: "/state"},
		// 	}
		// 	utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
		// case apierrors.ErrNotFound:
		// 	code := models.CodeNotFound
		// 	errInfo := &models.Error{
		// 		Code:        &code,
		// 		Description: apierrors.ErrorDescriptionNotFound,
		// 		Source:      &models.Source{Field: "/dataset_id"},
		// 	}
		// 	utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
		// default:
		// 	api.handleInternalError(ctx, w, r, "bundle update failed", err, logData)
		// }
		return
	}

	bundleJSON, err := json.Marshal(updatedBundle)
	if err != nil {
		api.handleInternalError(ctx, w, r, "failed to marshal bundle to JSON", err, logData)
		return
	}

	dpresponse.SetETag(w, updatedBundle.ETag)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(bundleJSON); err != nil {
		log.Error(ctx, "putBundle endpoint: error writing response body", err, logData)
	}
}

func (api *BundleAPI) handleBadRequestError(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, logdata log.Data) {
	log.Error(ctx, "putBundle endpoint: "+message, err, logdata)
	code := models.CodeInvalidParameters
	errInfo := &models.Error{
		Code:        &code,
		Description: apierrors.ErrorDescriptionMalformedRequest,
	}
	utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
}

func (api *BundleAPI) handleInternalError(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, logdata log.Data) {
	log.Error(ctx, "putBundle endpoint: "+message, err, logdata)
	code := models.CodeInternalError
	errInfo := &models.Error{
		Code:        &code,
		Description: apierrors.ErrorDescriptionInternalError,
	}
	utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
}

func (api *BundleAPI) handleGetBundleError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error, logdata log.Data) {
	if err == apierrors.ErrBundleNotFound {
		log.Error(ctx, "putBundle endpoint: bundle not found", err, logdata)
		code := models.CodeNotFound
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionNotFound,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
		return
	}
	api.handleInternalError(ctx, w, r, "failed to get bundle", err, logdata)
}

// Helper function to create and validate bundle update
func (api BundleAPI) CreateAndValidateBundleUpdate(r *http.Request, bundleID string, email string) (*models.Bundle, []*models.Error, error) {
	bundleUpdate, err := models.CreateBundle(r.Body, email)
	if err != nil {
		return nil, nil, err
	}

	bundleUpdate.ID = bundleID
	// bundleUpdate.CreatedAt = currentBundle.CreatedAt
	// bundleUpdate.CreatedBy = currentBundle.CreatedBy

	models.CleanBundle(bundleUpdate)

	validationErrors := models.ValidateBundle(bundleUpdate)
	if len(validationErrors) > 0 {
		return bundleUpdate, validationErrors, nil
	}

	return bundleUpdate, nil, nil
}
