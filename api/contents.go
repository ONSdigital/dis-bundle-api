package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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

	bundleExists, err := api.stateMachineBundleAPI.CheckBundleExists(ctx, bundleID)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to check if bundle exists", err, logdata)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to check if bundle exists",
		}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusInternalServerError)
		return
	}
	if !bundleExists {
		log.Error(ctx, "postBundleContents endpoint: bundle not found", nil, logdata)
		code := models.CodeNotFound
		errInfo := &models.Error{
			Code:        &code,
			Description: "Bundle not found",
		}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusNotFound)
		return
	}

	contentItem, err := models.CreateContentItem(r.Body)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to create content item from request body", err, logdata)
		code := models.CodeBadRequest
		errInfo := &models.Error{
			Code:        &code,
			Description: "Unable to process request due to a malformed or invalid request body or query parameter",
		}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusBadRequest)
		return
	}

	contentItem.BundleID = bundleID
	models.CleanContentItem(contentItem)

	err = models.ValidateContentItem(contentItem)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: content item validation failed", err, logdata)
		code := models.CodeBadRequest
		errInfo := &models.Error{
			Code:        &code,
			Description: "Content item validation failed",
			Source:      &models.Source{Field: err.Error()},
		}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusBadRequest)
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
				Description: "version not found",
			}
			utils.HandleBundleAPIErr(w, r, errInfo, http.StatusNotFound)
			return
		} else {
			log.Error(ctx, "postBundleContents endpoint: failed to get version from dataset API", err, logdata)
			code := models.CodeInternalServerError
			errInfo := &models.Error{
				Code:        &code,
				Description: "Failed to get version from dataset API",
			}
			utils.HandleBundleAPIErr(w, r, errInfo, http.StatusInternalServerError)
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
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusInternalServerError)
		return
	}

	if exists {
		log.Error(ctx, "postBundleContents endpoint: content item already exists for the given dataset, edition, and version", nil, logdata)
		code := models.CodeConflict
		errInfo := &models.Error{
			Code:        &code,
			Description: "Content item already exists for the given dataset, edition, and version",
		}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusConflict)
		return
	}

	err = api.stateMachineBundleAPI.CreateContentItem(ctx, contentItem)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to create content item in the datastore", err, logdata)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to create content item in the datastore",
		}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusInternalServerError)
		return
	}

	location := "/bundles/" + bundleID + "/contents/" + contentItem.ID

	event := &models.Event{
		RequestedBy: &models.RequestedBy{
			ID:    "placeholder-user-id",
			Email: "placeholder-email",
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
			Description: "Event validation failed",
			Source:      &models.Source{Field: err.Error()},
		}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusInternalServerError)
		return
	}

	err = api.stateMachineBundleAPI.CreateBundleEvent(ctx, event)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to create bundle event", err, logdata)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to create bundle event",
		}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusInternalServerError)
		return
	}

	userEmail, err := api.authMiddleware.Parse(r.Header.Get("Authorization"))
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to parse user email from authorization header", err, logdata)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to parse user email from authorization header",
		}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusInternalServerError)
		return
	}

	bundleUpdate, err := api.stateMachineBundleAPI.UpdateBundleETag(ctx, bundleID, userEmail.UserID)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to update bundle ETag", err, logdata)
		code := models.CodeInternalServerError
		errInfo := &models.Error{
			Code:        &code,
			Description: "Failed to update bundle ETag",
		}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusInternalServerError)
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
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusInternalServerError)
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
