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
	dpresponse.SetETag(w, bundle.ETag)

	_, err = w.Write(bundleBytes)
	if err != nil {
		log.Error(ctx, "failed writing bytes to response", err, logData)
		return
	}

	log.Info(ctx, "getBundle endpoint: request successful", logData)
}

func (api *BundleAPI) createBundle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	bundle, err := models.CreateBundle(r.Body)
	if err != nil {
		if err == errs.ErrUnableToParseJSON {
			log.Error(ctx, "failed to create bundle from request body", err)
			code := models.CodeBadRequest
			e := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionMalformedRequest,
			}
			utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusBadRequest)
			return
		} else if err == errs.ErrUnableToParseTime {
			log.Error(ctx, "invalid time format in request body", err)
			code := models.ErrInvalidParameters
			e := &models.Error{
				Code:        &code,
				Description: "Invalid time format in request body",
				Source: &models.Source{
					Field: "scheduled_at",
				},
			}
			utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusBadRequest)
			return
		} else {
			log.Error(ctx, "failed to read request body", err)
			code := models.CodeInternalServerError
			e := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionInternalError,
			}
			utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusInternalServerError)
			return
		}
	}

	authToken := r.Header.Get("Authorization")
	if authToken == "" {
		log.Error(ctx, "authorization token is missing", nil)
		code := models.CodeUnauthorized
		e := &models.Error{
			Code:        &code,
			Description: "Authorization token is required",
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusUnauthorized)
		return
	}

	authToken = strings.TrimPrefix(authToken, "Bearer ")

	var entityData *permSDK.EntityData
	entityData, err = api.authMiddleware.Parse(authToken)
	if err != nil {
		log.Error(ctx, "failed to parse auth token", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusInternalServerError)
		return
	}

	bundle.CreatedBy = &models.User{
		Email: entityData.UserID,
	}

	bundle.LastUpdatedBy = &models.User{
		Email: entityData.UserID,
	}

	bundleErrs := models.ValidateBundle(bundle)
	if len(bundleErrs) > 0 {
		log.Error(ctx, "failed to validate bundle", nil, log.Data{"errors": bundleErrs})
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: bundleErrs}, http.StatusBadRequest)
		return
	}

	err = api.stateMachineBundleAPI.StateMachine.Transition(ctx, api.stateMachineBundleAPI, nil, bundle)
	if err != nil {
		log.Error(ctx, "failed to transition bundle state", err)
		code := models.CodeBadRequest
		e := &models.Error{
			Code:        &code,
			Description: fmt.Sprintf("Failed to transition bundle state: %s", err.Error()),
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusBadRequest)
		return
	}

	_, err = api.stateMachineBundleAPI.GetBundleByTitle(ctx, bundle.Title)
	if err != nil {
		if err != errs.ErrBundleNotFound {
			log.Error(ctx, "failed to check existing bundle by title", err)
			code := models.CodeInternalServerError
			e := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionInternalError,
			}
			utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusInternalServerError)
			return
		}
	} else {
		log.Error(ctx, "bundle with the same title already exists", nil)
		code := models.CodeConflict
		e := &models.Error{
			Code:        &code,
			Description: "A bundle with the same title already exists",
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusConflict)
		return
	}

	createdBundle, err := api.stateMachineBundleAPI.CreateBundle(ctx, bundle)
	if err != nil {
		if err == errs.ErrScheduledAtRequired {
			log.Error(ctx, "scheduled_at is required for scheduled bundles", err)
			code := models.CodeBadRequest
			e := &models.Error{
				Code:        &code,
				Description: "scheduled_at is required for scheduled bundles",
				Source: &models.Source{
					Field: "/scheduled_at",
				},
			}
			utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusBadRequest)
			return
		}
		if err == errs.ErrScheduledAtSet {
			log.Error(ctx, "scheduled_at should not be set for manual bundles", err)
			code := models.CodeBadRequest
			e := &models.Error{
				Code:        &code,
				Description: "scheduled_at should not be set for manual bundles",
				Source: &models.Source{
					Field: "/scheduled_at",
				},
			}
			utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusBadRequest)
			return
		}
		if err == errs.ErrScheduledAtInPast {
			log.Error(ctx, "scheduled_at cannot be in the past", err)
			code := models.CodeBadRequest
			e := &models.Error{
				Code:        &code,
				Description: "scheduled_at cannot be in the past",
				Source: &models.Source{
					Field: "/scheduled_at",
				},
			}
			utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusBadRequest)
			return
		}
		log.Error(ctx, "failed to create bundle in the database", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusInternalServerError)
		return
	}

	eventBundle := &models.EventBundle{
		ID:            createdBundle.ID,
		BundleType:    createdBundle.BundleType,
		CreatedBy:     createdBundle.CreatedBy,
		CreatedAt:     createdBundle.CreatedAt,
		LastUpdatedBy: createdBundle.LastUpdatedBy,
		PreviewTeams:  createdBundle.PreviewTeams,
		ScheduledAt:   createdBundle.ScheduledAt,
		State:         createdBundle.State,
		Title:         createdBundle.Title,
		UpdatedAt:     createdBundle.UpdatedAt,
		ManagedBy:     createdBundle.ManagedBy,
	}

	event := &models.Event{
		RequestedBy: &models.RequestedBy{
			ID:    entityData.UserID,
			Email: entityData.UserID,
		},
		Action:   models.ActionCreate,
		Resource: "/bundles",
		Bundle:   eventBundle,
	}

	err = models.ValidateEvent(event)
	if err != nil {
		log.Error(ctx, "failed to validate event", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusInternalServerError)
		return
	}

	err = api.stateMachineBundleAPI.CreateBundleEvent(ctx, event)
	if err != nil {
		log.Error(ctx, "failed to create bundle event", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusInternalServerError)
		return
	}

	err = writeResponse(ctx, w, createdBundle)
	if err != nil {
		log.Error(ctx, "failed to write response", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		utils.HandleBundleAPIErrors(w, r, models.ErrorList{Errors: []*models.Error{e}}, http.StatusInternalServerError)
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
