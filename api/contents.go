package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

func (api *BundleAPI) postBundleContents(w http.ResponseWriter, r *http.Request) {
	defer dphttp.DrainBody(r)

	ctx := r.Context()
	vars := mux.Vars(r)
	bundleID := vars["bundle-id"]

	logdata := log.Data{"bundle_id": bundleID}

	contentItem, err := models.CreateContentItem(r.Body)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to create content item from request body", err, logdata)
		code := models.CodeBadRequest
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionMalformedRequest,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
		return
	}

	contentItem.BundleID = bundleID
	models.CleanContentItem(contentItem)

	validationErrs := models.ValidateContentItem(contentItem)
	if len(validationErrs) > 0 {
		logdata["validation_errors"] = validationErrs
		log.Error(ctx, "postBundleContents endpoint: content item validation failed", apierrors.ErrInvalidBody, logdata)
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, validationErrs...)
		return
	}

	bundleExists, err := api.stateMachineBundleAPI.CheckBundleExists(ctx, bundleID)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to check if bundle exists", err, logdata)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to check if bundle exists",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}
	if !bundleExists {
		log.Error(ctx, "postBundleContents endpoint: bundle not found", nil, logdata)
		code := models.CodeNotFound
		errInfo := &models.Error{
			Code:        &code,
			Description: "Bundle not found",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
		return
	}

	var authHeaders datasetAPISDK.Headers
	if r.Header.Get("X-Florence-Token") != "" {
		authHeaders.ServiceToken = r.Header.Get("X-Florence-Token")
	} else {
		authHeaders.ServiceToken = r.Header.Get("Authorization")
	}

	_, err = api.datasetAPIClient.GetVersion(ctx, authHeaders, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, strconv.Itoa(contentItem.Metadata.VersionID))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Error(ctx, "postBundleContents endpoint: version not found in dataset API", nil, logdata)
			code := models.CodeNotFound
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionNotFound,
				Source:      &models.Source{Field: "/metadata/dataset_id"},
			}
			utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
			return
		} else {
			log.Error(ctx, "postBundleContents endpoint: failed to get version from dataset API", err, logdata)
			code := models.CodeInternalServerError
			errInfo := &models.Error{
				Code:        &code,
				Description: "Failed to get version from dataset API",
			}
			utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
			return
		}
	}

	exists, err := api.stateMachineBundleAPI.CheckContentItemExistsByDatasetEditionVersion(ctx, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, contentItem.Metadata.VersionID)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to check if content item exists", err, logdata)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to check if content item exists",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	if exists {
		log.Error(ctx, "postBundleContents endpoint: content item already exists for the given dataset, edition, and version", nil, logdata)
		code := models.CodeConflict
		errInfo := &models.Error{
			Code:        &code,
			Description: "Content item already exists for the given dataset, edition and version",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusConflict, errInfo)
		return
	}

	err = api.stateMachineBundleAPI.CreateContentItem(ctx, contentItem)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to create content item in database", err, logdata)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to create content item in the datastore",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	JWTEntityData, err := api.authMiddleware.Parse(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to parse JWT from authorization header", err, logdata)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to get user identity from JWT",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	location := "/bundles/" + bundleID + "/contents/" + contentItem.ID

	event := &models.Event{
		RequestedBy: &models.RequestedBy{
			ID:    JWTEntityData.UserID,
			Email: JWTEntityData.UserID,
		},
		Action:      models.ActionCreate,
		Resource:    location,
		ContentItem: contentItem,
	}

	err = models.ValidateEvent(event)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: event validation failed", err, logdata)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to validate event",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	err = api.stateMachineBundleAPI.CreateBundleEvent(ctx, event)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to create event in database", err, logdata)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to create event",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	bundleUpdate, err := api.stateMachineBundleAPI.UpdateBundleETag(ctx, bundleID, JWTEntityData.UserID)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to update bundle ETag", err, logdata)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to update bundle ETag",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	contentItemJSON, err := json.Marshal(contentItem)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to marshal content item to JSON", err, logdata)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to marshal content item to JSON",
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	dpresponse.SetETag(w, bundleUpdate.ETag)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Location", location)
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(contentItemJSON); err != nil {
		log.Error(ctx, "postBundleContents endpoint: error writing response body", err, logdata)
	}
}

func (api *BundleAPI) getBundleContents(w http.ResponseWriter, r *http.Request, limit, offset int) (contents any, totalCount int, contentErrors *models.Error) {
	//fetch bundle ID
	ctx := r.Context()
	vars := mux.Vars(r)
	bundleID := vars["bundle-id"]
	logdata := log.Data{"bundle_id": bundleID}

	//check if the bundle exists
	bundleExists, err := api.stateMachineBundleAPI.CheckBundleExists(ctx, bundleID)
	if err != nil {
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to check if bundle exists",
		}
		return []*models.ContentItem{}, 0, errInfo
	}

	if !bundleExists {
		code := models.CodeNotFound
		errInfo := &models.Error{
			Code:        &code,
			Description: "Bundle not found",
		}
		return []*models.ContentItem{}, 0, errInfo
	}

	authHeaders := datasetAPISDK.Headers{}
	if r.Header.Get("X-Florence-Token") != "" {
		authHeaders.ServiceToken = r.Header.Get("X-Florence-Token")
	} else {
		authHeaders.ServiceToken = r.Header.Get("Authorization")
	}

	bundleContents, totalCount, err := api.stateMachineBundleAPI.GetBundleContents(ctx, bundleID, offset, limit, authHeaders)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Error(ctx, "getBundleContents endpoint: dataset not found in dataset API", nil, logdata)
			code := models.CodeNotFound
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionNotFound,
				Source:      &models.Source{Field: "/metadata/dataset_id"},
			}
			return nil, 0, errInfo
		} else {
			log.Error(ctx, "getBundleContents endpoint: failed to get dataset from dataset API", err, logdata)
			code := models.CodeInternalServerError
			errInfo := &models.Error{
				Code:        &code,
				Description: "Failed to get dataset from dataset API",
			}
			return nil, 0, errInfo
		}
	}
	return bundleContents, totalCount, nil
}
