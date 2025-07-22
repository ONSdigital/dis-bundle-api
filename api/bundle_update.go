package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	permSDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

func (api *BundleAPI) putBundle(w http.ResponseWriter, r *http.Request) {
	defer dphttp.DrainBody(r)

	var entityData *permSDK.EntityData

	ctx := r.Context()
	vars := mux.Vars(r)
	bundleID := vars["bundle-id"]

	logdata := log.Data{"bundle_id": bundleID}

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

	entityData, err := api.getAuthData(r)
	if err != nil {
		handleErr(ctx, w, r, err, logdata, RouteNamePutBundle)
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

	bundleUpdate, validationErrors, err := api.CreateAndValidateBundleUpdate(r, bundleID, currentBundle, entityData.UserID)
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

	validationErrs := api.stateMachineBundleAPI.ValidateBundleRules(ctx, bundleUpdate, currentBundle)
	if len(validationErrs) > 0 {
		logdata["validation_errors"] = validationErrs
		log.Error(ctx, "putBundle endpoint: bundle business rule validation failed", nil, logdata)
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, validationErrs...)
		return
	}

	authHeaders := getAuthHeaders(r)

	updatedBundle, err := api.stateMachineBundleAPI.PutBundle(ctx, bundleID, bundleUpdate, currentBundle, entityData, authHeaders)
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
		return nil, validationErrors, nil
	}

	return bundleUpdate, nil, nil
}

func getAuthHeaders(r *http.Request) datasetAPISDK.Headers {
	var authHeaders datasetAPISDK.Headers
	if r.Header.Get("X-Florence-Token") != "" {
		authHeaders.ServiceToken = r.Header.Get("X-Florence-Token")
	} else {
		authHeaders.ServiceToken = r.Header.Get("Authorization")
	}
	return authHeaders
}
