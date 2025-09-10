package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	"github.com/ONSdigital/log.go/v2/log"
)

func (api *BundleAPI) postBundleContents(w http.ResponseWriter, r *http.Request) {
	defer dphttp.DrainBody(r)

	ctx := r.Context()
	bundleID, logData := getBundleIDAndLogData(r)

	authEntityData, err := api.GetAuthEntityData(r)
	if err != nil {
		handleErr(ctx, w, r, err, logData, RouteNamePostBundleContents)
		return
	}

	contentItem, err := models.CreateContentItem(r.Body)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to create content item from request body", err, logData)
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
		log.Error(ctx, "postBundleContents endpoint: content item validation failed", apierrors.ErrInvalidBody, logData)
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, validationErrs...)
		return
	}

	bundleExists, err := api.stateMachineBundleAPI.CheckBundleExists(ctx, bundleID)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to check if bundle exists", err, logData)
		code := models.CodeInternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}
	if !bundleExists {
		log.Error(ctx, "postBundleContents endpoint: bundle not found", nil, logData)
		code := models.CodeNotFound
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionNotFound,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
		return
	}

	_, err = api.stateMachineBundleAPI.GetVersion(ctx, authEntityData.Headers, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, strconv.Itoa(contentItem.Metadata.VersionID))
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "dataset not found"):
			log.Error(ctx, "postBundleContents endpoint: dataset not found in dataset API", nil, logData)
			code := models.CodeNotFound
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionNotFound,
				Source:      &models.Source{Field: "/metadata/dataset_id"},
			}
			utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
			return

		case strings.Contains(err.Error(), "edition not found"):
			log.Error(ctx, "postBundleContents endpoint: edition not found in dataset API", nil, logData)
			code := models.CodeNotFound
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionNotFound,
				Source:      &models.Source{Field: "/metadata/edition_id"},
			}
			utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
			return

		case strings.Contains(err.Error(), "version not found"):
			log.Error(ctx, "postBundleContents endpoint: version not found in dataset API", nil, logData)
			code := models.CodeNotFound
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionNotFound,
				Source:      &models.Source{Field: "/metadata/version_id"},
			}
			utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
			return

		default:
			log.Error(ctx, "postBundleContents endpoint: failed to get version from dataset API", err, logData)
			code := models.CodeInternalError
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionInternalError,
			}
			utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
			return
		}
	}

	exists, err := api.stateMachineBundleAPI.CheckContentItemExistsByDatasetEditionVersion(ctx, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, contentItem.Metadata.VersionID)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to check if content item exists", err, logData)
		code := models.CodeInternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	if exists {
		log.Error(ctx, "postBundleContents endpoint: content item already exists for the given dataset, edition, and version", nil, logData)
		code := models.CodeConflict
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionConflict,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusConflict, errInfo)
		return
	}

	err = api.stateMachineBundleAPI.CreateContentItem(ctx, contentItem)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to create content item in database", err, logData)
		code := models.CodeInternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	location := "/bundles/" + bundleID + "/contents/" + contentItem.ID

	event := &models.Event{
		RequestedBy: &models.RequestedBy{
			ID:    authEntityData.GetUserID(),
			Email: authEntityData.GetUserEmail(),
		},
		Action:      models.ActionCreate,
		Resource:    location,
		ContentItem: contentItem,
	}

	err = models.ValidateEvent(event)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: event validation failed", err, logData)
		code := models.CodeInternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	err = api.stateMachineBundleAPI.CreateBundleEvent(ctx, event)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to create event in database", err, logData)
		code := models.CodeInternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	bundleUpdate, err := api.stateMachineBundleAPI.UpdateBundleETag(ctx, bundleID, authEntityData.GetUserEmail())
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to update bundle ETag", err, logData)
		code := models.CodeInternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	if bundleUpdate.BundleType == models.BundleTypeScheduled {
		err = api.stateMachineBundleAPI.UpdateDatasetVersionReleaseDate(ctx, bundleUpdate.ScheduledAt, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, contentItem.Metadata.VersionID, authEntityData.Headers)
		if err != nil {
			handleErr(ctx, w, r, err, logData, RouteNamePostBundleContents)
			return
		}
	}

	contentItemJSON, err := json.Marshal(contentItem)
	if err != nil {
		log.Error(ctx, "postBundleContents endpoint: failed to marshal content item to JSON", err, logData)
		code := models.CodeInternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
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
		log.Error(ctx, "postBundleContents endpoint: error writing response body", err, logData)
	}
}

func (api *BundleAPI) deleteContentItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bundleID, contentID, logData := getBundleIDAndContentIDAndLogData(r)

	authEntityData, err := api.GetAuthEntityData(r)
	if err != nil {
		handleErr(ctx, w, r, err, logData, RouteNameDeleteContentItem)
		return
	}

	contentItem, err := api.stateMachineBundleAPI.Datastore.GetContentItemByBundleIDAndContentItemID(ctx, bundleID, contentID)
	if err != nil {
		if err == apierrors.ErrContentItemNotFound {
			log.Error(ctx, "deleteContentItem endpoint: content item not found", err, logData)
			code := models.CodeNotFound
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionNotFound,
			}
			utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
			return
		}
		log.Error(ctx, "deleteContentItem endpoint: failed to get content item by ID", err, logData)
		code := models.CodeInternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	if contentItem.State != nil && *contentItem.State == models.StatePublished {
		log.Error(ctx, "deleteContentItem endpoint: cannot delete content item in published state", nil, logData)
		code := models.CodeConflict
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionConflict,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusConflict, errInfo)
		return
	}

	err = api.stateMachineBundleAPI.DeleteContentItem(ctx, contentID)
	if err != nil {
		if err == apierrors.ErrContentItemNotFound {
			log.Error(ctx, "deleteContentItem endpoint: content item not found for deletion", err, logData)
			code := models.CodeNotFound
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionNotFound,
			}
			utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
			return
		}
		log.Error(ctx, "deleteContentItem endpoint: failed to delete content item", err, logData)
		code := models.CodeInternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	updatedBundle, err := api.stateMachineBundleAPI.UpdateBundleETag(ctx, bundleID, authEntityData.GetUserID())
	if err != nil {
		log.Error(ctx, "failed to update bundle ETag", err, logData)
		code := models.CodeInternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	event := &models.Event{
		RequestedBy: &models.RequestedBy{
			ID:    authEntityData.GetUserID(),
			Email: authEntityData.GetUserEmail(),
		},
		Action:      models.ActionDelete,
		Resource:    "/bundles/" + bundleID + "/contents/" + contentID,
		ContentItem: contentItem,
	}

	err = models.ValidateEvent(event)
	if err != nil {
		log.Error(ctx, "deleteContentItem endpoint: event validation failed", err, logData)
		code := models.CodeInternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}

	err = api.stateMachineBundleAPI.CreateBundleEvent(ctx, event)
	if err != nil {
		log.Error(ctx, "deleteContentItem endpoint: failed to create event in database", err, logData)
		code := models.CodeInternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return
	}
	dpresponse.SetETag(w, updatedBundle.ETag)
	w.WriteHeader(http.StatusNoContent)
}

func (api *BundleAPI) getBundleContents(w http.ResponseWriter, r *http.Request, limit, offset int) (contents any, totalCount int, contentErrors *models.Error) {
	ctx := r.Context()
	bundleID, logData := getBundleIDAndLogData(r)

	authEntityData, err := api.GetAuthEntityData(r)
	if err != nil {
		errInfo := models.GetMatchingModelError(err)
		return []*models.ContentItem{}, 0, errInfo
	}

	bundleExists, err := api.stateMachineBundleAPI.CheckBundleExists(ctx, bundleID)
	if err != nil {
		code := models.CodeInternalError
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		return []*models.ContentItem{}, 0, errInfo
	}

	if !bundleExists {
		code := models.CodeNotFound
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionNotFound,
		}
		return []*models.ContentItem{}, 0, errInfo
	}

	bundleContents, totalCount, err := api.stateMachineBundleAPI.GetBundleContents(ctx, bundleID, offset, limit, authEntityData.Headers)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Error(ctx, "getBundleContents endpoint: dataset not found in dataset API", nil, logData)
			code := models.CodeNotFound
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionNotFound,
				Source:      &models.Source{Field: "/metadata/dataset_id"},
			}
			return nil, 0, errInfo
		} else {
			log.Error(ctx, "getBundleContents endpoint: failed to get dataset from dataset API", err, logData)
			code := models.CodeInternalError
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionInternalError,
			}
			return nil, 0, errInfo
		}
	}
	return bundleContents, totalCount, nil
}
