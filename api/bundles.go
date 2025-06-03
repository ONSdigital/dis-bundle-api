package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
	permsdk "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

var (
	ErrDescription              = "Unable to process request due to a malformed or invalid request body or query parameter."
	ErrInternalErrorDescription = "An internal error occurred."
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

func (api *BundleAPI) createBundle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.Info(ctx, "createBundle: creating a new bundle")

	bundle, err := models.CreateBundle(r.Body)
	if err != nil {
		log.Error(ctx, "failed to create bundle from request body", err)
		code := models.ErrInvalidParameters
		e := models.Error{
			Code:        &code,
			Description: ErrDescription,
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusBadRequest)
		return
	}

	log.Info(ctx, "createBundle: created bundle from request body", log.Data{"bundle": bundle})

	err = models.ValidateBundle(bundle)
	if err != nil {
		bundleErrs := models.GetBundleErrors(bundle)
		log.Error(ctx, "failed to validate bundle", err)
		utils.HandleBundleAPIErrors(w, r, bundleErrs, http.StatusBadRequest)
		return
	}

	err = api.stateMachineBundleAPI.StateMachine.Transition(ctx, api.stateMachineBundleAPI, nil, bundle)
	if err != nil {
		log.Error(ctx, "failed to transition bundle state", err)
		code := models.CodeBadRequest
		e := models.Error{
			Code:        &code,
			Description: fmt.Sprintf("Failed to transition bundle state: %s", err.Error()),
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusBadRequest)
		return
	}

	_, err = api.stateMachineBundleAPI.GetBundleByTitle(ctx, bundle.Title)
	if err != nil {
		if err != apierrors.ErrBundleNotFound {
			log.Error(ctx, "failed to check existing bundle by title", err)
			code := models.CodeInternalServerError
			e := models.Error{
				Code:        &code,
				Description: ErrInternalErrorDescription,
			}
			utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusInternalServerError)
			return
		}
	} else {
		log.Error(ctx, "bundle with the same title already exists", nil)
		code := models.CodeConflict
		e := models.Error{
			Code:        &code,
			Description: "A bundle with the same title already exists",
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusConflict)
		return
	}

	if bundle.BundleType == models.BundleTypeScheduled && bundle.ScheduledAt == nil {
		log.Error(ctx, "scheduled_at is required for scheduled bundles", nil)
		code := models.CodeBadRequest
		e := models.Error{
			Code:        &code,
			Description: "scheduled_at is required for scheduled bundles",
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusBadRequest)
		return
	}

	if bundle.BundleType == models.BundleTypeManual && bundle.ScheduledAt != nil {
		log.Error(ctx, "scheduled_at should not be set for manual bundles", nil)
		code := models.CodeBadRequest
		e := models.Error{
			Code:        &code,
			Description: "scheduled_at should not be set for manual bundles",
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusBadRequest)
		return
	}

	if bundle.ScheduledAt != nil && bundle.ScheduledAt.Before(time.Now()) {
		log.Error(ctx, "scheduled_at cannot be in the past", nil)
		code := models.CodeBadRequest
		e := models.Error{
			Code:        &code,
			Description: "scheduled_at cannot be in the past",
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusBadRequest)
		return
	}

	authToken := r.Header.Get("Authorization")
	if authToken == "" {
		log.Error(ctx, "authorization token is missing", nil)
		code := models.CodeUnauthorized
		e := models.Error{
			Code:        &code,
			Description: "Authorization token is required",
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusUnauthorized)
		return
	}

	authToken = strings.TrimPrefix(authToken, "Bearer ")

	var entityData *permsdk.EntityData
	if strings.Contains(authToken, ".") {
		entityData, err = api.authMiddleware.Parse(authToken)
		if err != nil {
			log.Error(ctx, "failed to parse auth token", err)
			code := models.CodeInternalServerError
			e := models.Error{
				Code:        &code,
				Description: ErrInternalErrorDescription,
			}
			utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusInternalServerError)
			return
		}
	}

	log.Info(ctx, "createBundle: successfully parsed JWT", log.Data{"entityData": entityData})

	etag, err := generateEtag(bundle)
	if err != nil {
		log.Error(ctx, "failed to generate ETag for bundle", err)
		code := models.CodeInternalServerError
		e := models.Error{
			Code:        &code,
			Description: ErrInternalErrorDescription,
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusInternalServerError)
		return
	}

	bundle.ETag = etag

	createdBundle, err := api.stateMachineBundleAPI.CreateBundle(ctx, bundle)
	if err != nil {
		log.Error(ctx, "failed to create bundle in the database", err)
		code := models.CodeInternalServerError
		e := models.Error{
			Code:        &code,
			Description: ErrInternalErrorDescription,
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusInternalServerError)
		return
	}

	event := &models.Event{
		RequestedBy: &models.RequestedBy{
			ID:    entityData.UserID,
			Email: entityData.UserID,
		},
		Action:   models.ActionCreate,
		Resource: "/bundles",
		Bundle:   bundle,
	}

	err = models.ValidateEvent(event)
	if err != nil {
		log.Error(ctx, "failed to validate event", err)
		code := models.CodeInternalServerError
		e := models.Error{
			Code:        &code,
			Description: ErrInternalErrorDescription,
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusInternalServerError)
		return
	}

	err = api.stateMachineBundleAPI.CreateBundleEvent(ctx, event)
	if err != nil {
		log.Error(ctx, "failed to create bundle event", err)
		code := models.CodeInternalServerError
		e := models.Error{
			Code:        &code,
			Description: ErrInternalErrorDescription,
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusInternalServerError)
		return
	}

	err = writeResponse(ctx, w, createdBundle)
	if err != nil {
		log.Error(ctx, "failed to write response", err)
		code := models.CodeInternalServerError
		e := models.Error{
			Code:        &code,
			Description: ErrInternalErrorDescription,
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: &[]models.Error{e}}, http.StatusInternalServerError)
		return
	}
}

func writeResponse(ctx context.Context, w http.ResponseWriter, bundle *models.Bundle) error {
	b, err := json.Marshal(bundle)
	if err != nil {
		return err
	}
	dpresponse.SetETag(w, bundle.ETag)
	w.Header().Set("Cache-Control", "no-store")
	location := "/bundles/" + bundle.ID
	w.Header().Set("Location", location)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	log.Info(ctx, "createBundle: successfully created bundle", log.Data{"bundle_id": bundle.ID})
	return nil
}

func generateEtag(bundle *models.Bundle) (string, error) {
	b, err := json.Marshal(bundle)
	if err != nil {
		return "", err
	}
	etag := dpresponse.GenerateETag(b, false)
	etag = strings.Trim(etag, `"`)
	return etag, nil
}
