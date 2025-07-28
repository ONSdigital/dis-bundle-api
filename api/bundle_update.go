package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

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

	logdata := log.Data{"bundle_id": bundleID}

	authEntityData, err := api.GetAuthEntityData(r)
	if err != nil {
		handleErr(ctx, w, r, err, logdata, RouteNamePutBundle)
		return
	}

	ifMatchHeader := r.Header.Get("If-Match")
	if ifMatchHeader == "" {
		log.Error(ctx, "putBundle endpoint: missing If-Match header", nil, logdata)
		code := models.CodeMissingParameters
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionMissingIfMatchHeader,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
		return
	}

	currentBundle, err := api.stateMachineBundleAPI.GetBundle(ctx, bundleID)
	if err != nil {
		api.handleGetBundleError(ctx, w, r, err, logdata)
		return
	}

	if currentBundle.ETag != ifMatchHeader {
		log.Error(ctx, "putBundle endpoint: ETag mismatch", nil, logdata)
		code := models.CodeConflict
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionConflict,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusConflict, errInfo)
		return
	}

	bundleUpdate, validationErrors, err := api.CreateAndValidateBundleUpdate(r, bundleID, currentBundle, authEntityData.GetUserID())
	if err != nil {
		api.handleBadRequestError(ctx, w, r, "bundle creation or validation failed", err, logdata)
		return
	}

	var allValidationErrors []*models.Error
	allValidationErrors = append(allValidationErrors, validationErrors...)

	validationErrs := api.stateMachineBundleAPI.ValidateBundleRules(ctx, bundleUpdate, currentBundle)
	allValidationErrors = append(allValidationErrors, validationErrs...)

	if bundleUpdate.State != "" && bundleUpdate.State.IsValid() && currentBundle.State != "" && bundleUpdate.State != currentBundle.State {
		err := api.stateMachineBundleAPI.StateMachine.Transition(context.Background(), api.stateMachineBundleAPI, currentBundle, bundleUpdate)
		if err != nil {
			if err == apierrors.ErrInvalidTransition || strings.Contains(err.Error(), "state not allowed to transition") {
				code := models.CodeInvalidParameters
				stateError := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionMalformedRequest,
					Source:      &models.Source{Field: "/state"},
				}
				allValidationErrors = append(allValidationErrors, stateError)
			}
		}
	}

	if len(allValidationErrors) > 0 {
		logdata["validation_errors"] = allValidationErrors
		log.Error(ctx, "putBundle endpoint: bundle validation failed", nil, logdata)
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, allValidationErrors...)
		return
	}

	updatedBundle, err := api.stateMachineBundleAPI.PutBundle(ctx, bundleID, bundleUpdate, currentBundle, authEntityData)
	if err != nil {
		log.Error(ctx, "putBundle endpoint: bundle update failed", err, logdata)
		switch err {
		case apierrors.ErrInvalidTransition:
			code := models.CodeInvalidParameters
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionMalformedRequest,
				Source:      &models.Source{Field: "/state"},
			}
			utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
		case apierrors.ErrNotFound:
			code := models.CodeNotFound
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionNotFound,
				Source:      &models.Source{Field: "/dataset_id"},
			}
			utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
		default:
			api.handleInternalError(ctx, w, r, "bundle update failed", err, logdata)
		}
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
func (api BundleAPI) CreateAndValidateBundleUpdate(r *http.Request, bundleID string, currentBundle *models.Bundle, email string) (*models.Bundle, []*models.Error, error) {
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
		return bundleUpdate, validationErrors, nil
	}

	return bundleUpdate, nil, nil
}
