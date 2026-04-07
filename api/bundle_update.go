package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/slack"
	"github.com/ONSdigital/dis-bundle-api/utils"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

//nolint:gocognit,gocyclo // cognitive complexity 21 (> 20) is acceptable for now
func (api *BundleAPI) putBundle(w http.ResponseWriter, r *http.Request) {
	defer dphttp.DrainBody(r)

	ctx := r.Context()
	vars := mux.Vars(r)
	bundleID := vars["bundle-id"]

	logdata := log.Data{"bundle_id": bundleID}

	authEntityData, err := api.GetAuthEntityData(r)
	if err != nil {
		handleErr(ctx, w, r, err, logdata, RouteNamePutBundle)
		return
	}

	ifMatchHeader := r.Header.Get("If-Match")
	if ifMatchHeader == "" {
		log.Error(ctx, "putBundle endpoint: missing If-Match header", nil, logdata)
		code := models.CodeMissingParameters
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionMissingIfMatchHeader,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
		return
	}

	currentBundle, err := api.stateMachineBundleAPI.GetBundle(ctx, bundleID)
	if err != nil {
		api.handleGetBundleError(ctx, w, r, err, logdata)
		return
	}

	if currentBundle.ETag != ifMatchHeader {
		log.Error(ctx, "putBundle endpoint: ETag mismatch", nil, logdata)
		code := models.CodeConflict
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionConflict,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusConflict, errInfo)
		return
	}

	bundleUpdate, validationErrors, err := api.CreateAndValidateBundleUpdate(r, bundleID, currentBundle, authEntityData.GetUserID())
	if err != nil {
		api.handleBadRequestError(ctx, w, r, "bundle creation or validation failed", err, logdata)
		return
	}

	var allValidationErrors []*models.Error
	allValidationErrors = append(allValidationErrors, validationErrors...)

	validationErrs := api.stateMachineBundleAPI.ValidateBundleRules(ctx, bundleUpdate, currentBundle)
	allValidationErrors = append(allValidationErrors, validationErrs...)

	if bundleUpdate.State != "" && bundleUpdate.State.IsValid() && currentBundle.State != "" && bundleUpdate.State != currentBundle.State {
		err := api.stateMachineBundleAPI.StateMachine.Transition(context.Background(), api.stateMachineBundleAPI, currentBundle, bundleUpdate)
		if err != nil {
			if err == apierrors.ErrInvalidTransition {
				code := models.CodeInvalidParameters
				stateError := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionMalformedRequest,
					Source:      &models.Source{Field: "/state"},
				}
				allValidationErrors = append(allValidationErrors, stateError)
			} else {
				api.handleInternalError(ctx, w, r, "state transition validation failed", err, logdata)
				return
			}
		}
	}

	if len(allValidationErrors) > 0 {
		logdata["validation_errors"] = allValidationErrors
		log.Error(ctx, "putBundle endpoint: bundle validation failed", nil, logdata)
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, allValidationErrors...)
		return
	}

	// Create policies for any preview teams added in the update.
	if err := api.stateMachineBundleAPI.CreateBundlePolicies(ctx, authEntityData.Headers.AccessToken, bundleUpdate.PreviewTeams, models.RoleDatasetsPreviewer); err != nil {
		api.handleInternalError(ctx, w, r, "failed to create bundle policies", err, logdata)
		return
	}

	// Add policy conditions for newly added teams for existing content items.
	if err := api.stateMachineBundleAPI.AddPolicyConditionsForAddedPreviewTeams(ctx, authEntityData.Headers.AccessToken, bundleID, currentBundle.PreviewTeams, bundleUpdate.PreviewTeams); err != nil {
		api.handleInternalError(ctx, w, r, "failed to add policy conditions for added preview teams", err, logdata)
		return
	}

	// Remove policy conditions for any preview teams removed in the update.
	if err := api.stateMachineBundleAPI.RemovePolicyConditionsForRemovedPreviewTeams(ctx, authEntityData.Headers.AccessToken, bundleID, currentBundle.PreviewTeams, bundleUpdate.PreviewTeams); err != nil {
		api.handleInternalError(ctx, w, r, "failed to remove policy conditions for removed preview teams", err, logdata)
		return
	}

	var publishLogFields []slack.Field
	var slackMessageRef *slack.MessageRef

	isPublishTransition := currentBundle.State.String() == models.BundleStateApproved.String() && bundleUpdate.State.String() == models.BundleStatePublished.String()
	publishStartTime := time.Now()
	if isPublishTransition {
		contentItemCount, err := api.stateMachineBundleAPI.Datastore.Backend.CountBundleContents(ctx, bundleID)
		if err != nil {
			// Don't block publication if we can't get the content item count.
			// Log the error and continue with a count of 0.
			// If it is a critical issue then a Slack alarm be raised with the failed PutBundle call.
			log.Error(ctx, "failed to count bundle contents: continuing with count 0", err, logdata)
		}

		publishLogFields = []slack.Field{
			{Title: "Bundle ID", Value: bundleID},
			{Title: "Title", Value: currentBundle.Title},
			{Title: "Type", Value: currentBundle.BundleType.String()},
			{Title: "Number of Content Items", Value: strconv.Itoa(contentItemCount)},
			{Title: "Publish Start Date", Value: publishStartTime.Format(utils.SlackPublishTimeFormat)},
		}
		logdata["slack_fields"] = publishLogFields

		log.Info(ctx, "sending slack notification: Bundle publish started", logdata)
		slackMessageRef, err = api.stateMachineBundleAPI.DataBundleSlackClient.SendPublishLog(ctx, "Bundle publish started", publishLogFields)
		if err != nil {
			log.Error(ctx, "failed to send slack notification: Bundle publish started", err, logdata)
		}
	}

	updatedBundle, err := api.stateMachineBundleAPI.PutBundle(ctx, bundleID, bundleUpdate, currentBundle, authEntityData)
	if err != nil {
		log.Error(ctx, "putBundle endpoint: bundle update failed", err, logdata)

		if isPublishTransition {
			_, slackErr := api.stateMachineBundleAPI.DataBundleSlackClient.SendAlarm(ctx, "Failed to publish bundle", err, publishLogFields)
			if slackErr != nil {
				log.Error(ctx, "failed to send slack notification: Failed to publish bundle", slackErr, logdata)
			}
		}

		switch err {
		case apierrors.ErrInvalidTransition:
			code := models.CodeInvalidParameters
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionMalformedRequest,
				Source:      &models.Source{Field: "/state"},
			}
			utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
		case apierrors.ErrNotFound:
			code := models.CodeNotFound
			errInfo := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionNotFound,
				Source:      &models.Source{Field: "/dataset_id"},
			}
			utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
		default:
			api.handleInternalError(ctx, w, r, "bundle update failed", err, logdata)
		}
		return
	}

	if isPublishTransition {
		publishEndTime := time.Now()
		publishLogFields = append(publishLogFields,
			slack.Field{Title: "Publish End Date", Value: publishEndTime.Format(utils.SlackPublishTimeFormat)},
			slack.Field{Title: "Duration", Value: fmt.Sprintf("%.4f seconds", publishEndTime.Sub(publishStartTime).Seconds())},
		)
		logdata["slack_fields"] = publishLogFields

		log.Info(ctx, "updating slack notification: Bundle publish completed", logdata)
		_, err := api.stateMachineBundleAPI.DataBundleSlackClient.UpdatePublishLog(ctx, slackMessageRef, "Bundle publish completed", publishLogFields)
		if err != nil {
			log.Error(ctx, "failed to send slack notification: Bundle publish completed", err, logdata)
		}
	}

	dpresponse.SetETag(w, updatedBundle.ETag)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(updatedBundle); err != nil {
		log.Error(ctx, "putBundle endpoint: error encoding response body", err, logdata)
	}
}

func (api *BundleAPI) handleBadRequestError(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, logdata log.Data) {
	log.Error(ctx, "putBundle endpoint: "+message, err, logdata)
	code := models.CodeInvalidParameters
	errInfo := &models.Error{
		Code:        &code,
		Description: apierrors.ErrorDescriptionMalformedRequest,
	}
	utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, errInfo)
}

func (api *BundleAPI) handleInternalError(ctx context.Context, w http.ResponseWriter, r *http.Request, message string, err error, logdata log.Data) {
	log.Error(ctx, "putBundle endpoint: "+message, err, logdata)
	code := models.CodeInternalError
	errInfo := &models.Error{
		Code:        &code,
		Description: apierrors.ErrorDescriptionInternalError,
	}
	utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
}

func (api *BundleAPI) handleGetBundleError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error, logdata log.Data) {
	if err == apierrors.ErrBundleNotFound {
		log.Error(ctx, "putBundle endpoint: bundle not found", err, logdata)
		code := models.CodeNotFound
		errInfo := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionNotFound,
		}
		utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
		return
	}
	api.handleInternalError(ctx, w, r, "failed to get bundle", err, logdata)
}

// Helper function to create and validate bundle update
func (api BundleAPI) CreateAndValidateBundleUpdate(r *http.Request, bundleID string, currentBundle *models.Bundle, email string) (*models.Bundle, []*models.Error, error) {
	bundleUpdate, err := models.CreateBundle(r.Body, email)
	if err != nil {
		return nil, nil, err
	}

	bundleUpdate.ID = bundleID
	bundleUpdate.CreatedAt = currentBundle.CreatedAt
	bundleUpdate.CreatedBy = currentBundle.CreatedBy

	models.CleanBundle(bundleUpdate)

	validationErrors := models.ValidateBundle(bundleUpdate)
	if len(validationErrors) > 0 {
		return bundleUpdate, validationErrors, nil
	}

	return bundleUpdate, nil, nil
}
