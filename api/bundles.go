package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/filters"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"

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
	dpresponse.SetETag(w, bundle.ETag)

	_, err = w.Write(bundleBytes)
	if err != nil {
		log.Error(ctx, "failed writing bytes to response", err, logData)
		return
	}

	log.Info(ctx, "getBundle endpoint: request successful", logData)
}

func (api *BundleAPI) putBundleState(w http.ResponseWriter, r *http.Request) (errBundles *models.Error) {
	ctx := r.Context()

	etag, err := utils.GetETag(r)
	if err != nil {
		return err
	}

	bundleId, err := utils.GetBundleID(r)
	if err != nil {
		return err
	}

	updateRequest, err := utils.GetRequestBody[models.UpdateBundleStateRequest](r)
	if err != nil {
		return err
	}

	updatedBundle, err := api.stateMachineBundleAPI.UpdateBundleState(ctx, r, bundleId, updateRequest.State, *etag)
	if err != nil {
		return err
	}

	utils.
		CreateHTTPResponseBuilder().
		WithETag(updatedBundle.ETag).
		WithCacheControl(utils.CacheControlNoStore).
		WithStatusCode(http.StatusOK).
		Build(w)

	w.WriteHeader(http.StatusOK)

	return nil
}
