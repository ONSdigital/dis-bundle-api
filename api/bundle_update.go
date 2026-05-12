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

//nolint:gocognit,gocyclo // cognitive complexity 21 (> 20) is acceptable for now
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

	ifMatchHeader := r.Header.Get("If-Match")
	if ifMatchHeader == "" {
		log.Error(ctx, "putBundle endpoint: missing If-Match header", nil, logData)
		code := models.CodeMissingParameters
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionMissingIfMatchHeader,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
		return
	}

	bundleUpdate, validationErrors, err := api.CreateAndValidateBundleUpdate(r, bundleID, authEntityData.GetUserID())
	if err != nil {
		api.handleBadRequestError(ctx, w, r, "bundle creation or validation failed", err, logData)
		return
	}

	var allValidationErrors []*models.Error
	allValidationErrors = append(allValidationErrors, validationErrors...)

	if len(allValidationErrors) > 0 {
		logData["validation_errors"] = allValidationErrors
		log.Error(ctx, "putBundle endpoint: bundle validation failed", nil, logData)
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, allValidationErrors...)
		return
	}

	updatedBundle, err := api.stateMachineBundleAPI.PutBundle(ctx, bundleID, bundleUpdate, authEntityData, ifMatchHeader)
	if err != nil {
		log.Error(ctx, "putBundle endpoint: bundle update failed", err, logData)

		if err == apierrors.ErrBundleTitleAlreadyExists {
			code := models.CodeInvalidParameters
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionMalformedRequest,
				Source:      &models.Source{Field: "/title"},
			}
			utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
			return
		}
		if err == apierrors.ErrInvalidTransition {
			code := models.CodeInvalidParameters
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionInvalidStateTransition,
				Source:      &models.Source{Field: "/state"},
			}
			utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
			return
		}
		handleErr(ctx, w, r, err, logData, RouteNamePutBundle)
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

// Helper function to create and validate bundle update
func (api BundleAPI) CreateAndValidateBundleUpdate(r *http.Request, bundleID, email string) (*models.Bundle, []*models.Error, error) {
	bundleUpdate, err := models.CreateBundle(r.Body, email)
	if err != nil {
		return nil, nil, err
	}

	bundleUpdate.ID = bundleID

	models.CleanBundle(bundleUpdate)

	validationErrors := models.ValidateBundle(bundleUpdate)
	if len(validationErrors) > 0 {
		return bundleUpdate, validationErrors, nil
	}

	return bundleUpdate, nil, nil
}
