package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/filters"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	permSDK "github.com/ONSdigital/dp-permissions-api/sdk"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

const (
	// Route variable names
	RouteVariableBundleID  = "bundle-id"
	RouteVariableContentID = "content-id"

	// Route names
	RouteNameGetBundle = "getBundle"
	RouteNamePutBundle = "putBundle"
)

func (api *BundleAPI) getBundles(w http.ResponseWriter, r *http.Request, limit, offset int) (successResult *models.PaginationSuccessResult[models.Bundle], errorResult *models.ErrorResult[models.Error]) {
	ctx := r.Context()

	filters, filtersErr := filters.CreateBundlefilters(r)
	if filtersErr != nil {
		log.Error(ctx, filtersErr.Error.Error(), errs.ErrInvalidQueryParameter)
		code := models.CodeInternalServerError
		invalidRequestError := &models.Error{Code: &code, Description: errs.ErrorDescriptionMalformedRequest, Source: filtersErr.Source}
		return nil, models.CreateInternalServerErrorResult(invalidRequestError)
	}

	bundles, totalCount, err := api.stateMachineBundleAPI.ListBundles(ctx, offset, limit, filters)
	if err != nil {
		code := models.CodeInternalServerError
		log.Error(ctx, "failed to get bundles", err)
		internalServerError := &models.Error{Code: &code, Description: errs.ErrorDescriptionInternalError}
		return nil, models.CreateInternalServerErrorResult(internalServerError)
	}

	if totalCount == 0 && filters.PublishDate != nil {
		code := models.CodeNotFound
		log.Warn(ctx, fmt.Sprintf("Request for bundles with publish_date %s produced no results", filters.PublishDate))
		notFoundError := &models.Error{Code: &code, Description: errs.ErrorDescriptionNotFound}
		return nil, models.CreateNotFoundResult(notFoundError)
	}

	return models.CreatePaginationSuccessResult(bundles, totalCount), nil
}

func (api *BundleAPI) getBundle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bundleID, logData := getBundleIDAndLogData(r)

	bundle, err := api.stateMachineBundleAPI.GetBundle(ctx, bundleID)
	if err != nil {
		handleErr(ctx, w, r, err, logData, RouteNameGetBundle)
		return
	}

	bundleBytes := setETagAndCacheControlHeaders(ctx, w, r, bundle, logData)

	_, err = w.Write(bundleBytes)
	if err != nil {
		log.Error(ctx, "failed writing bytes to response", err, logData)
		return
	}

	logSuccessfulRequest(ctx, logData, RouteNameGetBundle)
}

func (api *BundleAPI) putBundleState(w http.ResponseWriter, r *http.Request) {
	defer dphttp.DrainBody(r)

	ctx := r.Context()

	bundleID, logData := getBundleIDAndLogData(r)

	etag, err := utils.GetETag(r)
	if err != nil {
		handleErr(ctx, w, r, err, logData, RouteNamePutBundle)
		return
	}

	stateRequest, err := getUpdateStateRequestBody(r)
	if err != nil {
		handleErr(ctx, w, r, err, logData, RouteNamePutBundle)
		return
	}

	authData, err := api.GetAuthEntityData(r)

	if err != nil {
		handleErr(ctx, w, r, err, logData, RouteNamePutBundle)
		return
	}

	bundle, err := api.stateMachineBundleAPI.UpdateBundleState(ctx, bundleID, *etag, stateRequest.State, authData)

	if err != nil {
		handleErr(ctx, w, r, err, logData, RouteNamePutBundle)
		return
	}

	setETagAndCacheControlHeaders(ctx, w, r, bundle, logData)

	w.WriteHeader(http.StatusOK)
	logSuccessfulRequest(ctx, logData, RouteNamePutBundle)
}

func getUpdateStateRequestBody(r *http.Request) (*models.UpdateStateRequest, error) {
	stateRequest, err := utils.GetRequestBody[models.UpdateStateRequest](r)
	if err != nil {
		return nil, err
	}

	if stateRequest.State == "" || !stateRequest.State.IsValid() {
		return nil, errs.ErrInvalidBody
	}

	return stateRequest, nil
}

func handleErr(ctx context.Context, w http.ResponseWriter, r *http.Request, err error, logData log.Data, endpoint string) {
	errorEvent := fmt.Sprintf("%s endpoint: %s", endpoint, err.Error())
	log.Error(ctx, errorEvent, err, logData)
	errInfo := models.GetMatchingModelError(err)
	httpStatus := errs.GetStatusCodeForErr(err)
	utils.HandleBundleAPIErr(w, r, httpStatus, errInfo)
}

func logSuccessfulRequest(ctx context.Context, logData log.Data, endpoint string) {
	log.Info(ctx, fmt.Sprintf("%s endpoint: request successful", endpoint), logData)
}

func getBundleIDAndLogData(r *http.Request) (string, log.Data) {
	vars := mux.Vars(r)
	bundleID := vars[RouteVariableBundleID]
	logData := log.Data{RouteVariableBundleID: bundleID}
	return bundleID, logData
}

func getBundleIDAndContentIDAndLogData(r *http.Request) (bundleID, contentID string, logData log.Data) {
	vars := mux.Vars(r)
	bundleID = vars[RouteVariableBundleID]
	contentID = vars[RouteVariableContentID]
	logData = log.Data{
		RouteVariableBundleID:  bundleID,
		RouteVariableContentID: contentID,
	}
	return bundleID, contentID, logData
}

func setETagAndCacheControlHeaders(ctx context.Context, w http.ResponseWriter, r *http.Request, bundle *models.Bundle, logData log.Data) []byte {
	bundleBytes, err := json.Marshal(bundle)
	if err != nil {
		log.Error(ctx, "failed to marshal bundle into bytes", err, logData)
		errInfo := models.CreateModelError(models.CodeInternalServerError, errs.ErrMarshalJSONObject)
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return nil
	}

	w.Header().Set("Cache-Control", "no-store")

	// Set Etag
	ETag := bundle.ETag

	if ETag == "" {
		ETag = bundle.GenerateETag(&bundleBytes)
	}

	dpresponse.SetETag(w, ETag)

	return bundleBytes
}

func (api *BundleAPI) createBundle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var entityData *permSDK.EntityData
	entityData, err := api.authMiddleware.Parse(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if err != nil {
		log.Error(ctx, "createBundle: failed to parse auth token", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, e)
		return
	}

	bundle, err := models.CreateBundle(r.Body, entityData.UserID)
	if err != nil {
		if err == errs.ErrUnableToParseJSON {
			log.Error(ctx, "createBundle: failed to create bundle from request body", err)
			code := models.CodeBadRequest
			e := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionMalformedRequest,
			}
			utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, e)
			return
		} else if err == errs.ErrUnableToParseTime {
			log.Error(ctx, "createBundle: invalid time format in request body", err)
			code := models.ErrInvalidParameters
			e := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionInvalidTimeFormat,
				Source: &models.Source{
					Field: "scheduled_at",
				},
			}
			utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, e)
			return
		} else {
			log.Error(ctx, "createBundle: failed to read request body", err)
			code := models.CodeInternalServerError
			e := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionInternalError,
			}
			utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, e)
			return
		}
	}

	bundleErrs := models.ValidateBundle(bundle)
	err = api.stateMachineBundleAPI.ValidateScheduledAt(bundle)
	if err != nil {
		if err == errs.ErrScheduledAtRequired {
			log.Error(ctx, "createBundle: scheduled_at is required for scheduled bundles", err)
			code := models.CodeInvalidParameters
			e := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionScheduledAtIsRequired,
				Source: &models.Source{
					Field: "/scheduled_at",
				},
			}
			bundleErrs = append(bundleErrs, e)
		}
		if err == errs.ErrScheduledAtSet {
			log.Error(ctx, "createBundle: scheduled_at should not be set for manual bundles", err)
			code := models.CodeInvalidParameters
			e := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionScheduledAtShouldNotBeSet,
				Source: &models.Source{
					Field: "/scheduled_at",
				},
			}
			bundleErrs = append(bundleErrs, e)
		}
		if err == errs.ErrScheduledAtInPast {
			log.Error(ctx, "createBundle: scheduled_at cannot be in the past", err)
			code := models.CodeInvalidParameters
			e := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionScheduledAtIsInPast,
				Source: &models.Source{
					Field: "/scheduled_at",
				},
			}
			bundleErrs = append(bundleErrs, e)
		}
	}
	if len(bundleErrs) > 0 {
		log.Error(ctx, "createBundle: failed to validate bundle", nil, log.Data{"errors": bundleErrs})
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, bundleErrs...)
		return
	}

	statusCode, createdBundle, errObject, err := api.stateMachineBundleAPI.CreateBundle(ctx, bundle)
	if err != nil {
		log.Error(ctx, "createBundle: failed to create bundle", err)
		utils.HandleBundleAPIErr(w, r, statusCode, errObject)
		return
	}

	b, err := json.Marshal(createdBundle)
	if err != nil {
		log.Error(ctx, "createBundle: failed to marshal created bundle", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, e)
		return
	}

	dpresponse.SetETag(w, createdBundle.ETag)

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Location", "/bundles/"+createdBundle.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err = w.Write(b); err != nil {
		log.Error(ctx, "createBundle: error writing response body", err)
	}
}

func (api *BundleAPI) deleteBundle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bundleID, logData := getBundleIDAndLogData(r)

	var entityData *permSDK.EntityData
	entityData, err := api.authMiddleware.Parse(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if err != nil {
		log.Error(ctx, "deleteBundle: failed to parse auth token", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, e)
		return
	}

	statusCode, errObject, err := api.stateMachineBundleAPI.DeleteBundle(ctx, bundleID, entityData.UserID)
	if err != nil {
		log.Error(ctx, "deleteBundle endpoint: failed to delete bundle", err, logData)
		utils.HandleBundleAPIErr(w, r, statusCode, errObject)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
