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

func (api *BundleAPI) getBundleByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	bundleID := vars["bundle_id"]
	logData := log.Data{"bundle_id": bundleID}

	bundles, err := api.stateMachineBundleAPI.GetBundleByID(ctx, bundleID)
	if err != nil {
		if err == apierrors.ErrBundleNotFound {
			log.Error(ctx, "getBundleByID endpoint: bundle id not found", err, logData)
			code := models.CodeNotFound
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrResourceNotFound,
			}
			utils.HandleBundleAPIErr(w, r, errInfo, http.StatusNotFound)
		} else {
			log.Error(ctx, "An internal error occurred", err, logData)
			code := models.CodeInternalServerError
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrInternalErrorDescription,
			}
			utils.HandleBundleAPIErr(w, r, errInfo, http.StatusInternalServerError)
		}
		return
	}

	setMainHeaders(w)
	if bundles.ETag != "" {
		dpresponse.SetETag(w, bundles.ETag)
	}

	versionBytes, err := json.Marshal(bundles)
	if err != nil {
		log.Error(ctx, "failed to marshal version resource into bytes", err, logData)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrUnmarshallJSONObject,
		}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusInternalServerError)
		return
	}

	_, err = w.Write(versionBytes)
	if err != nil {
		log.Error(ctx, "failed writing bytes to response", err, logData)
		code := models.CodeNotFound
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrWritingBytesToResponse,
		}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusInternalServerError)
		return
	}

	log.Info(ctx, "getBundleById endpoint: request successful", logData)
}
