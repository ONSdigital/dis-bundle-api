package api

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
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

func (api *BundleAPI) getBundle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	bundleID := vars["bundle-id"]
	logData := log.Data{"bundle-id": bundleID}

	bundles, err := api.stateMachineBundleAPI.GetBundle(ctx, bundleID)
	if err != nil {
		if err == apierrors.ErrBundleNotFound {
			log.Error(ctx, "getBundle endpoint: bundle id not found", err, logData)
			code := models.CodeNotFound
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrResourceNotFound,
			}
			utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
		} else {
			log.Error(ctx, "An internal error occurred", err, logData)
			code := models.CodeInternalServerError
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrInternalErrorDescription,
			}
			utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		}
		return
	}

	// Set the required headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	dpresponse.SetETag(w, bundles.ETag)

	bundleBytes, err := json.Marshal(bundles)
	if err != nil {
		log.Error(ctx, "failed to marshal bundle into bytes", err, logData)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrMarshallJSONObject,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	_, err = w.Write(bundleBytes)
	if err != nil {
		log.Error(ctx, "failed writing bytes to response", err, logData)
		code := models.InternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrWritingBytesToResponse,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	log.Info(ctx, "getBundle endpoint: request successful", logData)
}
