package application

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/filters"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

type StateMachineBundleAPI struct {
	Datastore        store.Datastore
	StateMachine     *StateMachine
	DatasetAPIClient datasetAPISDK.Clienter
}

func Setup(datastore store.Datastore, stateMachine *StateMachine, datasetAPIClient datasetAPISDK.Clienter) *StateMachineBundleAPI {
	return &StateMachineBundleAPI{
		Datastore:        datastore,
		StateMachine:     stateMachine,
		DatasetAPIClient: datasetAPIClient,
	}
}

func (s *StateMachineBundleAPI) ListBundles(ctx context.Context, offset, limit int, filters *filters.BundleFilters) ([]*models.Bundle, int, error) {
	results, totalCount, err := s.Datastore.ListBundles(ctx, offset, limit, filters)
	if err != nil {
		return nil, 0, err
	}
	return results, totalCount, nil
}

func (s *StateMachineBundleAPI) GetBundle(ctx context.Context, bundleID string) (*models.Bundle, error) {
	return s.Datastore.GetBundle(ctx, bundleID)
}

func (s *StateMachineBundleAPI) UpdateBundleETag(ctx context.Context, bundleID, email string) (*models.Bundle, error) {
	return s.Datastore.UpdateBundleETag(ctx, bundleID, email)
}

func (s *StateMachineBundleAPI) CheckBundleExists(ctx context.Context, bundleID string) (bool, error) {
	return s.Datastore.CheckBundleExists(ctx, bundleID)
}

func (s *StateMachineBundleAPI) CreateContentItem(ctx context.Context, contentItem *models.ContentItem) error {
	return s.Datastore.CreateContentItem(ctx, contentItem)
}

func (s *StateMachineBundleAPI) GetContentItemByBundleIDAndContentItemID(ctx context.Context, bundleID, contentItemID string) (*models.ContentItem, error) {
	return s.Datastore.GetContentItemByBundleIDAndContentItemID(ctx, bundleID, contentItemID)
}

func (s *StateMachineBundleAPI) ListBundleEvents(ctx context.Context, offset, limit int, bundleID string, after, before *time.Time) ([]*models.Event, int, error) {
	results, totalCount, err := s.Datastore.ListBundleEvents(ctx, offset, limit, bundleID, after, before)
	if err != nil {
		return nil, 0, err
	}
	return results, totalCount, nil
}

func (s *StateMachineBundleAPI) CheckAllBundleContentsAreApproved(ctx context.Context, bundleID string) (bool, error) {
	return s.Datastore.CheckAllBundleContentsAreApproved(ctx, bundleID)
}

func (s *StateMachineBundleAPI) CheckContentItemExistsByDatasetEditionVersion(ctx context.Context, datasetID, editionID string, versionID int) (bool, error) {
	return s.Datastore.CheckContentItemExistsByDatasetEditionVersion(ctx, datasetID, editionID, versionID)
}

func (s *StateMachineBundleAPI) DeleteContentItem(ctx context.Context, contentItemID string) error {
	return s.Datastore.DeleteContentItem(ctx, contentItemID)
}

func (s *StateMachineBundleAPI) CreateEventFromBundle(ctx context.Context, bundle *models.Bundle, email string, action models.Action) (*models.Error, error) {
	bundleEvent, err := models.ConvertBundleToBundleEvent(bundle)
	if err != nil {
		log.Error(ctx, "failed to convert bundle to bundle event", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		return e, err
	}

	event := &models.Event{
		RequestedBy: &models.RequestedBy{
			ID:    email,
			Email: email,
		},
		Action:   action,
		Resource: "/bundles/" + bundle.ID,
		Bundle:   bundleEvent,
	}

	err = models.ValidateEvent(event)
	if err != nil {
		log.Error(ctx, "failed to validate event", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		return e, err
	}

	err = s.CreateBundleEvent(ctx, event)
	if err != nil {
		log.Error(ctx, "failed to create bundle event", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		return e, err
	}

	return nil, nil
}

func (s *StateMachineBundleAPI) CreateEventFromContentItem(ctx context.Context, contentItem *models.ContentItem, email string, action models.Action) (*models.Error, error) {
	event := &models.Event{
		RequestedBy: &models.RequestedBy{
			ID:    email,
			Email: email,
		},
		Action:      action,
		Resource:    "/bundles/" + contentItem.BundleID + "/contents/" + contentItem.ID,
		ContentItem: contentItem,
	}

	err := models.ValidateEvent(event)
	if err != nil {
		log.Error(ctx, "failed to validate event", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		return e, err
	}

	err = s.CreateBundleEvent(ctx, event)
	if err != nil {
		log.Error(ctx, "failed to create bundle event", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		return e, err
	}

	return nil, nil
}

func (s *StateMachineBundleAPI) CreateBundleEvent(ctx context.Context, event *models.Event) error {
	return s.Datastore.CreateBundleEvent(ctx, event)
}

func (s *StateMachineBundleAPI) GetBundleAndValidateETag(ctx context.Context, bundleID, suppliedETag string) (*models.Bundle, error) {
	bundle, err := s.Datastore.GetBundle(ctx, bundleID)

	if err != nil {
		return nil, err
	}

	if bundle.ETag == "" {
		log.Warn(ctx, "ETag for bundle is empty; generating new", log.Data{"bundle-id": bundleID, "etag": bundle.ETag, "supplied-etag": suppliedETag})
		bundleBytes, err := json.Marshal(bundle)
		if err != nil {
			return nil, errs.ErrUnableToParseJSON
		}

		bundle.ETag = bundle.GenerateETag(&bundleBytes)
	}

	if bundle.ETag != suppliedETag {
		log.Warn(ctx, "ETag validation failed", log.Data{"bundle-id": bundleID, "etag": bundle.ETag, "supplied-etag": suppliedETag})
		return nil, errs.ErrInvalidIfMatchHeader
	}

	return bundle, nil
}

func (s *StateMachineBundleAPI) UpdateBundleState(ctx context.Context, bundleID, suppliedETag string, targetState models.BundleState, authEntityData *models.AuthEntityData) (*models.Bundle, error) {
	bundle, err := s.GetBundleAndValidateETag(ctx, bundleID, suppliedETag)

	if err != nil {
		return nil, err
	}

	bundle, err = s.StateMachine.TransitionBundle(ctx, s, bundle, &targetState, authEntityData)
	if err != nil {
		return nil, err
	}

	event, err := models.CreateEventModel(authEntityData.GetUserID(), authEntityData.GetUserEmail(), models.ActionUpdate, models.CreateBundleResourceLocation(bundle), nil, bundle)
	if err != nil {
		return nil, err
	}

	if err := s.Datastore.CreateBundleEvent(ctx, event); err != nil {
		return nil, err
	}

	return bundle, nil
}

func (s *StateMachineBundleAPI) updateVersionStateForContentItem(ctx context.Context, contentItem *models.ContentItem, targetState *models.BundleState, authToken string) error {
	headers := datasetAPISDK.Headers{
		UserAccessToken: authToken,
	}

	versionID := strconv.Itoa(contentItem.Metadata.VersionID)

	version, err := s.DatasetAPIClient.GetVersion(ctx, headers, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, versionID)

	if err != nil {
		return err
	}

	if !strings.EqualFold(version.State, contentItem.State.String()) {
		log.Warn(ctx, "Version state does not match ContentItem state", log.Data{"content-item-id": contentItem.ID, "version-state": version.State, "content-item-state": contentItem.State.String()})
		return errs.ErrVersionStateMismatched
	}

	if err := s.DatasetAPIClient.PutVersionState(ctx, headers, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, versionID, strings.ToLower(targetState.String())); err != nil {
		return err
	}

	return nil
}

func (s *StateMachineBundleAPI) CreateBundle(ctx context.Context, bundle *models.Bundle) (int, *models.Bundle, *models.Error, error) {
	err := s.StateMachine.Transition(ctx, s, nil, bundle)
	if err != nil {
		log.Error(ctx, "failed to transition bundle state", err)
		code := models.CodeBadRequest
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionStateNotAllowedToTransition,
		}
		return http.StatusBadRequest, nil, e, err
	}

	bundleExists, err := s.CheckBundleExistsByTitle(ctx, bundle.Title)
	if err != nil {
		log.Error(ctx, "failed to check existing bundle by title", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		return http.StatusInternalServerError, nil, e, err
	}

	if bundleExists {
		log.Error(ctx, "bundle with the same title already exists", errs.ErrBundleTitleAlreadyExists)
		code := models.CodeConflict
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionBundleTitleAlreadyExist,
			Source: &models.Source{
				Field: "/title",
			},
		}
		return http.StatusConflict, nil, e, errs.ErrBundleTitleAlreadyExists
	}

	err = s.Datastore.CreateBundle(ctx, bundle)
	if err != nil {
		log.Error(ctx, "failed to create bundle", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		return http.StatusInternalServerError, nil, e, err
	}

	createdBundle, err := s.GetBundle(ctx, bundle.ID)
	if err != nil {
		log.Error(ctx, "failed to retrieve created bundle", err)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		return http.StatusInternalServerError, nil, e, err
	}

	errObject, err := s.CreateEventFromBundle(ctx, bundle, createdBundle.CreatedBy.Email, models.ActionCreate)
	if err != nil {
		log.Error(ctx, "failed to create event from bundle", err)
		return http.StatusInternalServerError, nil, errObject, err
	}

	return http.StatusCreated, createdBundle, nil, nil
}

func (s *StateMachineBundleAPI) DeleteBundle(ctx context.Context, bundleID, email string) (int, *models.Error, error) {
	logData := log.Data{"bundle_id": bundleID}

	bundle, err := s.GetBundle(ctx, bundleID)
	if err != nil {
		if err == errs.ErrBundleNotFound {
			log.Error(ctx, "bundle not found", err, logData)
			code := models.CodeNotFound
			e := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionNotFound,
			}
			return http.StatusNotFound, e, err
		} else {
			log.Error(ctx, "failed to get bundle", err, logData)
			code := models.CodeInternalServerError
			e := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionInternalError,
			}
			return http.StatusInternalServerError, e, err
		}
	}

	err = s.StateMachine.Transition(ctx, s, bundle, nil)
	if err != nil {
		log.Error(ctx, "failed to transition bundle state", err, logData)
		code := models.CodeConflict
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionAlreadyPublished,
		}
		return http.StatusConflict, e, err
	}

	bundleContents, err := s.Datastore.ListBundleContentIDsWithoutLimit(ctx, bundleID)

	if err != nil {
		log.Error(ctx, "failed to retrieve bundle contents", err, logData)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		return http.StatusInternalServerError, e, err
	}

	if len(bundleContents) > 0 {
		for _, contentItem := range bundleContents {
			err = s.DeleteContentItem(ctx, contentItem.ID)
			if err != nil {
				log.Error(ctx, "failed to delete content item", err, log.Data{"bundle_id": bundleID, "content_item_id": contentItem.ID})
				code := models.CodeInternalServerError
				e := &models.Error{
					Code:        &code,
					Description: errs.ErrorDescriptionInternalError,
				}
				return http.StatusInternalServerError, e, err
			}

			errObject, err := s.CreateEventFromContentItem(ctx, contentItem, email, models.ActionDelete)
			if err != nil {
				log.Error(ctx, "failed to create event from content item", err, log.Data{"bundle_id": bundleID, "content_item_id": contentItem.ID})
				return http.StatusInternalServerError, errObject, err
			}
		}
	}

	err = s.Datastore.DeleteBundle(ctx, bundleID)
	if err != nil {
		log.Error(ctx, "failed to delete bundle", err, logData)
		code := models.CodeInternalServerError
		e := &models.Error{
			Code:        &code,
			Description: errs.ErrorDescriptionInternalError,
		}
		return http.StatusInternalServerError, e, err
	}

	errObject, err := s.CreateEventFromBundle(ctx, bundle, email, models.ActionDelete)
	if err != nil {
		log.Error(ctx, "failed to create event from bundle", err, logData)
		return http.StatusInternalServerError, errObject, err
	}

	return http.StatusNoContent, nil, nil
}

func (s *StateMachineBundleAPI) CheckBundleExistsByTitle(ctx context.Context, title string) (bool, error) {
	exists, err := s.Datastore.CheckBundleExistsByTitle(ctx, title)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *StateMachineBundleAPI) ValidateScheduledAt(bundle *models.Bundle) error {
	if bundle.BundleType == models.BundleTypeScheduled && bundle.ScheduledAt == nil {
		return errs.ErrScheduledAtRequired
	}

	if bundle.BundleType == models.BundleTypeManual && bundle.ScheduledAt != nil {
		return errs.ErrScheduledAtSet
	}

	if bundle.ScheduledAt != nil && bundle.ScheduledAt.Before(time.Now()) {
		return errs.ErrScheduledAtInPast
	}

	return nil
}

func (s *StateMachineBundleAPI) GetBundleContents(ctx context.Context, bundleID string, offset, limit int, authHeaders datasetAPISDK.Headers) ([]*models.ContentItem, int, error) {
	// Get bundle
	bundle, err := s.Datastore.GetBundle(ctx, bundleID)
	if err != nil {
		return nil, 0, err
	}
	bundleState := bundle.State

	totalCount := 0

	// If bundle is published, return its contents directly
	if bundleState.String() == models.BundleStatePublished.String() {
		contentResults, totalCount, err := s.Datastore.ListBundleContents(ctx, bundleID, offset, limit)
		if err != nil {
			return nil, 0, err
		}
		return contentResults, totalCount, nil
	}

	// If bundle is not published, populate state & title by calling dataset API Client
	contentResults, totalCount, err := s.Datastore.ListBundleContents(ctx, bundleID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	for _, contentItem := range contentResults {
		datasetID := contentItem.Metadata.DatasetID
		dataset, err := s.DatasetAPIClient.GetDataset(ctx, authHeaders, "", datasetID)

		if err != nil {
			log.Error(ctx, "failed to fetch dataset", err, log.Data{"dataset_id": datasetID})
			return nil, 0, err
		}

		contentItem.State = (*models.State)(&dataset.State)
		contentItem.Metadata.Title = dataset.Title
	}

	return contentResults, totalCount, nil
}
