package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/filters"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
	permSDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
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
	vars := mux.Vars(r)
	bundleID := vars["bundle-id"]
	logData := log.Data{"bundle-id": bundleID}

	bundle, err := api.stateMachineBundleAPI.GetBundle(ctx, bundleID)
	if err != nil {
		if err == errs.ErrBundleNotFound {
			log.Error(ctx, "getBundle endpoint: bundle id not found", err, logData)
			code := models.CodeNotFound
			errInfo := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionNotFound,
			}
			utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
		} else {
			log.Error(ctx, "An internal error occurred", err, logData)
			code := models.CodeInternalServerError
			errInfo := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionInternalError,
			}
			utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		}
		return
	}

	bundleBytes, err := json.Marshal(bundle)
	if err != nil {
		log.Error(ctx, "failed to marshal bundle into bytes", err, logData)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: errs.ErrMarshalJSONObject,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	// Set the required headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	// Set Etag
	ETag := bundle.ETag
	if ETag == "" {
		ETag = dpresponse.GenerateETag(bundleBytes, false)
	}
	dpresponse.SetETag(w, ETag)

	_, err = w.Write(bundleBytes)
	if err != nil {
		log.Error(ctx, "failed writing bytes to response", err, logData)
		return
	}

	log.Info(ctx, "getBundle endpoint: request successful", logData)
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
