package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/filters"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/slack"
	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/ONSdigital/dis-bundle-api/utils"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

type StateMachineBundleAPI struct {
	Datastore             store.Datastore
	StateMachine          *StateMachine
	DatasetAPIClient      datasetAPISDK.Clienter
	PermissionsAPIClient  permissionsAPISDK.Clienter
	DataBundleSlackClient slack.Clienter
	PreviewServiceURL     string
}

func Setup(datastore store.Datastore, stateMachine *StateMachine, datasetAPIClient datasetAPISDK.Clienter, permissionsAPIClient permissionsAPISDK.Clienter, dataBundleSlackClient slack.Clienter, previewServiceURL string) *StateMachineBundleAPI {
	return &StateMachineBundleAPI{
		Datastore:             datastore,
		StateMachine:          stateMachine,
		DatasetAPIClient:      datasetAPIClient,
		PermissionsAPIClient:  permissionsAPIClient,
		DataBundleSlackClient: dataBundleSlackClient,
		PreviewServiceURL:     previewServiceURL,
	}
}

func (s *StateMachineBundleAPI) ListBundles(ctx context.Context, offset, limit int, bundleFilters *filters.BundleFilters) ([]*models.Bundle, int, error) {
	results, totalCount, err := s.Datastore.ListBundles(ctx, offset, limit, bundleFilters)
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

func (s *StateMachineBundleAPI) CreateEvent(ctx context.Context, authEntityData *models.AuthEntityData, action models.Action, bundle *models.Bundle, contentItem *models.ContentItem) error {
	event, err := models.CreateEventModel(authEntityData.GetUserID(), authEntityData.GetUserEmail(), action, bundle, contentItem)
	if err != nil {
		log.Error(ctx, "failed to create event model", err)
		return err
	}

	return s.Datastore.CreateEvent(ctx, event)
}

func (s *StateMachineBundleAPI) UpdateBundle(ctx context.Context, bundleID string, bundle *models.Bundle) (*models.Bundle, error) {
	return s.Datastore.UpdateBundle(ctx, bundleID, bundle)
}

func (s *StateMachineBundleAPI) CheckBundleExistsByTitleUpdate(ctx context.Context, title, excludeID string) (bool, error) {
	return s.Datastore.CheckBundleExistsByTitleUpdate(ctx, title, excludeID)
}

func (s *StateMachineBundleAPI) GetContentItemsByBundleID(ctx context.Context, bundleID string) ([]*models.ContentItem, error) {
	return s.Datastore.GetContentItemsByBundleID(ctx, bundleID)
}

func (s *StateMachineBundleAPI) UpdateContentItemDatasetInfo(ctx context.Context, contentItemID, title, state string) error {
	return s.Datastore.UpdateContentItemDatasetInfo(ctx, contentItemID, title, state)
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
			return nil, apierrors.ErrUnableToParseJSON
		}

		bundle.ETag = bundle.GenerateETag(&bundleBytes)
	}

	if bundle.ETag != suppliedETag {
		log.Warn(ctx, "ETag validation failed", log.Data{"bundle-id": bundleID, "etag": bundle.ETag, "supplied-etag": suppliedETag})
		return nil, apierrors.ErrInvalidIfMatchHeader
	}

	return bundle, nil
}

func (s *StateMachineBundleAPI) UpdateBundleState(ctx context.Context, bundleID, suppliedETag string, targetState models.BundleState, authEntityData *models.AuthEntityData) (*models.Bundle, error) {
	logData := log.Data{"bundle_id": bundleID}
	userID := authEntityData.GetUserID()

	bundle, err := s.GetBundleAndValidateETag(ctx, bundleID, suppliedETag)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	bundle.UpdatedAt = &now
	bundle.LastUpdatedBy = &models.User{Email: userID}

	updatedBundle, err := s.StateMachine.Transition(ctx, s, bundle, targetState, *authEntityData)
	if err != nil {
		log.Error(ctx, "transition failed", err, logData)
		return nil, err
	}

	return updatedBundle, nil
}

func (s *StateMachineBundleAPI) CreateBundle(ctx context.Context, bundle *models.Bundle, authEntityData *models.AuthEntityData) (int, *models.Bundle, *models.Error, error) {
	bundleExists, err := s.CheckBundleExistsByTitle(ctx, bundle.Title)
	if err != nil {
		log.Error(ctx, "failed to check existing bundle by title", err)
		code := models.CodeInternalError
		e := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		return http.StatusInternalServerError, nil, e, err
	}

	if bundleExists {
		log.Error(ctx, "bundle with the same title already exists", apierrors.ErrBundleTitleAlreadyExists)
		code := models.CodeConflict
		e := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionBundleTitleAlreadyExist,
			Source: &models.Source{
				Field: "/title",
			},
		}
		return http.StatusConflict, nil, e, apierrors.ErrBundleTitleAlreadyExists
	}

	err = s.Datastore.CreateBundle(ctx, bundle)
	if err != nil {
		log.Error(ctx, "failed to create bundle", err)
		code := models.CodeInternalError
		e := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		return http.StatusInternalServerError, nil, e, err
	}

	createdBundle, err := s.GetBundle(ctx, bundle.ID)
	if err != nil {
		log.Error(ctx, "failed to retrieve created bundle", err)
		code := models.CodeInternalError
		e := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		return http.StatusInternalServerError, nil, e, err
	}

	if err = s.CreateEvent(ctx, authEntityData, models.ActionCreate, createdBundle, nil); err != nil {
		log.Error(ctx, "failed to create event", err, log.Data{"bundle_id": createdBundle.ID})
		return http.StatusInternalServerError, nil, models.GetMatchingModelError(err), err
	}

	return http.StatusCreated, createdBundle, nil, nil
}

func (s *StateMachineBundleAPI) DeleteBundle(ctx context.Context, bundleID string, authEntityData *models.AuthEntityData) (int, *models.Error, error) {
	identityType := log.USER
	if authEntityData.IsServiceAuth {
		identityType = log.SERVICE
	}
	logAuth := log.Auth(identityType, authEntityData.EntityData.UserID)

	bundle, err := s.GetBundle(ctx, bundleID)
	if err != nil {
		if err == apierrors.ErrBundleNotFound {
			code := models.CodeNotFound
			e := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionNotFound,
			}
			return http.StatusNotFound, e, err
		} else {
			code := models.CodeInternalError
			e := &models.Error{
				Code:        &code,
				Description: apierrors.ErrorDescriptionInternalError,
			}
			return http.StatusInternalServerError, e, err
		}
	}

	if bundle.State == models.BundleStatePublished {
		code := models.CodeConflict
		err := apierrors.ErrDeleteBundleForbidden
		e := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionConflict,
		}
		return http.StatusConflict, e, err
	}

	bundleContents, err := s.Datastore.GetContentItemsByBundleID(ctx, bundleID)
	if err != nil {
		code := models.CodeInternalError
		e := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		return http.StatusInternalServerError, e, err
	}

	if len(bundleContents) > 0 {
		for _, contentItem := range bundleContents {
			err = s.DeleteContentItem(ctx, contentItem.ID)
			if err != nil {
				log.Error(ctx, "failed to delete content item", err, log.Data{"bundle_id": bundleID, "content_item_id": contentItem.ID})
				code := models.CodeInternalError
				e := &models.Error{
					Code:        &code,
					Description: apierrors.ErrorDescriptionInternalError,
				}
				return http.StatusInternalServerError, e, err
			}

			for _, team := range *bundle.PreviewTeams {
				policy, err := s.PermissionsAPIClient.GetPolicy(ctx, team.ID, permissionsAPISDK.Headers{Authorization: authEntityData.Headers.AccessToken})
				if err != nil {
					log.Error(ctx, "failed to get permissions policy for preview team", err, log.Data{"bundle_id": bundleID, "content_item_id": contentItem.ID, "preview_team": team.ID})
					code := models.CodeNotFound
					e := &models.Error{
						Code:        &code,
						Description: apierrors.ErrorDescriptionNotFound,
					}
					return http.StatusNotFound, e, err
				}

				toRemove := []string{
					contentItem.Metadata.DatasetID,
					fmt.Sprintf("%s/%s", contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID),
				}

				policy.Condition.Values = removeConditionValuesForBundlePreviewTeam(policy.Condition.Values, toRemove...)

				err = s.PermissionsAPIClient.PutPolicy(ctx, team.ID, *policy, permissionsAPISDK.Headers{Authorization: authEntityData.Headers.AccessToken})
				if err != nil {
					log.Error(ctx, "failed to update permissions policy for preview team", err, log.Data{"bundle_id": bundleID, "content_item_id": contentItem.ID, "preview_team": team.ID})
					code := models.CodeInternalError
					e := &models.Error{
						Code:        &code,
						Description: apierrors.ErrorDescriptionInternalError,
					}
					return http.StatusInternalServerError, e, err
				}
			}

			contentItemForEvent := models.ContentItem{
				ID: contentItem.ID,
			}

			if err = s.CreateEvent(ctx, authEntityData, models.ActionDelete, nil, &contentItemForEvent); err != nil {
				log.Error(ctx, "failed to create event", err, log.Classification(log.ProtectiveMonitoring), logAuth, log.Data{"bundle_id": bundleID, "content_item_id": contentItem.ID, "action": models.ActionDelete})
				return http.StatusInternalServerError, models.GetMatchingModelError(err), err
			}
			log.Info(ctx, "bundle event creation successful", log.Classification(log.ProtectiveMonitoring), logAuth, log.Data{"action": models.ActionDelete})
		}
	}

	err = s.Datastore.DeleteBundle(ctx, bundleID)
	if err != nil {
		code := models.CodeInternalError
		e := &models.Error{
			Code:        &code,
			Description: apierrors.ErrorDescriptionInternalError,
		}
		return http.StatusInternalServerError, e, err
	}

	if err = s.CreateEvent(ctx, authEntityData, models.ActionDelete, bundle, nil); err != nil {
		log.Error(ctx, "failed to create event", err, log.Classification(log.ProtectiveMonitoring), logAuth, log.Data{"bundle_id": bundleID})
		return http.StatusInternalServerError, models.GetMatchingModelError(err), err
	}
	log.Info(ctx, "bundle event creation successful", log.Classification(log.ProtectiveMonitoring), logAuth, log.Data{"action": models.ActionDelete})

	return http.StatusNoContent, nil, nil
}

func (s *StateMachineBundleAPI) CheckBundleExistsByTitle(ctx context.Context, title string) (bool, error) {
	exists, err := s.Datastore.CheckBundleExistsByTitle(ctx, title)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// func (s *StateMachineBundleAPI) ValidateScheduledAt(bundle *models.Bundle) error {
// 	if bundle.BundleType == models.BundleTypeScheduled && bundle.ScheduledAt == nil {
// 		return apierrors.apierrorscheduledAtRequired
// 	}

// 	if bundle.BundleType == models.BundleTypeManual && bundle.ScheduledAt != nil {
// 		return apierrors.apierrorscheduledAtSet
// 	}

// 	if bundle.ScheduledAt != nil && bundle.ScheduledAt.Before(time.Now()) {
// 		return apierrors.apierrorscheduledAtInPast
// 	}

// 	return nil
// }

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
		editionID := contentItem.Metadata.EditionID
		versionID := strconv.Itoa(contentItem.Metadata.VersionID)

		dataset, err := s.DatasetAPIClient.GetDataset(ctx, authHeaders, datasetID)
		if err != nil {
			log.Error(ctx, "failed to fetch dataset", err, log.Data{"dataset_id": datasetID})
			return nil, 0, err
		}

		version, err := s.DatasetAPIClient.GetVersion(ctx, authHeaders, datasetID, editionID, versionID)
		if err != nil {
			log.Error(ctx, "failed to fetch dataset version", err, log.Data{"dataset_id": datasetID, "edition_id": editionID, "version_id": versionID})
			return nil, 0, err
		}
		contentItem.State = (*models.State)(&version.State)
		contentItem.Metadata.Title = dataset.Title
	}

	return contentResults, totalCount, nil
}

func (s *StateMachineBundleAPI) PutBundle(ctx context.Context, bundleID string, bundleUpdate *models.Bundle, authEntityData *models.AuthEntityData, eTag string) (*models.Bundle, error) {
	logData := log.Data{"bundle_id": bundleID}
	userID := authEntityData.GetUserID()

	originalBundle, err := s.GetBundleAndValidateETag(ctx, bundleID, eTag)
	if err != nil {
		return nil, err
	}

	if bundleUpdate.Title != originalBundle.Title {
		exists, err := s.CheckBundleExistsByTitleUpdate(ctx, bundleUpdate.Title, bundleUpdate.ID)
		if err != nil {
			log.Error(ctx, "failed to check bundle title uniqueness", err)
			return nil, err
		}
		if exists {
			log.Error(ctx, "bundle title already exists", err, logData)
			return nil, apierrors.ErrBundleTitleAlreadyExists
		}
	}

	// Create policies for any preview teams added in the update.
	// NOTE: This does not currently handle the case where existing preview teams are removed.
	// If a preview team is removed from the bundle, their policies will still exist.
	if err := s.CreateBundlePolicies(ctx, authEntityData.Headers.AccessToken, bundleUpdate.PreviewTeams, models.RoleDatasetsPreviewer); err != nil {
		log.Error(ctx, "failed to create bundle policies", err, logData)
		return nil, apierrors.ErrBundlePolicyFailedToCreate
	}

	// Add policy conditions for newly added teams for existing content items.
	if err := s.AddPolicyConditionsForAddedPreviewTeams(ctx, authEntityData.Headers.AccessToken, bundleID, originalBundle.PreviewTeams, bundleUpdate.PreviewTeams); err != nil {
		log.Error(ctx, "failed to add policy conditions for added preview teams", err, logData)
		return nil, apierrors.ErrBundleFailedToAddPolicyPreviewTeams
	}

	// Remove policy conditions for any preview teams removed in the update.
	if err := s.RemovePolicyConditionsForRemovedPreviewTeams(ctx, authEntityData.Headers.AccessToken, bundleID, originalBundle.PreviewTeams, bundleUpdate.PreviewTeams); err != nil {
		log.Error(ctx, "failed to remove policy conditions for removed preview teams", err, logData)
		return nil, apierrors.ErrBundleFailedToRemovePolicyPreviewTeams
	}

	now := time.Now()
	bundleUpdate.UpdatedAt = &now
	bundleUpdate.LastUpdatedBy = &models.User{Email: userID}

	bundleUpdate.CreatedAt = originalBundle.CreatedAt
	bundleUpdate.CreatedBy = originalBundle.CreatedBy

	// Store the state to move to incase this has changes
	nextState := bundleUpdate.State

	// Set the state to be the previous state to check for the state transition but holds all other updates to the record
	// the new state is applied in the enter function
	bundleUpdate.State = originalBundle.State

	updatedBundle, err := s.StateMachine.Transition(ctx, s, bundleUpdate, nextState, *authEntityData)
	if err != nil {
		log.Error(ctx, "transition failed", err, logData)
		return nil, err
	}

	return updatedBundle, nil
}

func (s *StateMachineBundleAPI) UpdateContentItemsWithDatasetInfo(ctx context.Context, bundleID string, authHeaders datasetAPISDK.Headers) error {
	contentItems, err := s.GetContentItemsByBundleID(ctx, bundleID)
	if err != nil {
		log.Error(ctx, "failed to get content items", err, log.Data{"bundle_id": bundleID})
		return err
	}

	if len(contentItems) == 0 {
		log.Info(ctx, "no content items found for bundle", log.Data{"bundle_id": bundleID})
		return nil
	}

	for _, contentItem := range contentItems {
		dataset, err := s.DatasetAPIClient.GetDataset(ctx, authHeaders, contentItem.Metadata.DatasetID)
		if err != nil {
			log.Error(ctx, "dataset api client call failed", err, log.Data{
				"content_item_id": contentItem.ID,
				"dataset_id":      contentItem.Metadata.DatasetID,
			})

			errorMsg := strings.ToLower(err.Error())
			if strings.Contains(errorMsg, "client failed to read datasetapi body") ||
				strings.Contains(errorMsg, "not found") {
				log.Error(ctx, "dataset not found", err, log.Data{
					"content_item_id": contentItem.ID,
					"dataset_id":      contentItem.Metadata.DatasetID,
				})
				return apierrors.ErrNotFound
			}

			return err
		}

		err = s.UpdateContentItemDatasetInfo(ctx, contentItem.ID, dataset.Title, strings.ToUpper(dataset.State))
		if err != nil {
			log.Error(ctx, "update content item failed", err, log.Data{
				"content_item_id": contentItem.ID,
				"dataset_id":      contentItem.Metadata.DatasetID,
			})
			continue
		}
	}
	return nil
}

func (s *StateMachineBundleAPI) UpdateDatasetVersionReleaseDate(ctx context.Context, releaseDate *time.Time, datasetID, editionID string, versionID int, authHeaders datasetAPISDK.Headers) error {
	versionUpdate := datasetAPIModels.Version{
		Type:        "static",
		ReleaseDate: releaseDate.UTC().Format("2006-01-02T15:04:05.000Z"),
	}

	_, err := s.DatasetAPIClient.PutVersion(ctx, authHeaders, datasetID, editionID, strconv.Itoa(versionID), versionUpdate)
	if err != nil {
		log.Error(ctx, "failed to update dataset version", err, log.Data{
			"dataset_id":   datasetID,
			"edition_id":   editionID,
			"version_id":   versionID,
			"release_date": releaseDate,
		})
		return err
	}

	return nil
}

func removeConditionValuesForBundlePreviewTeam(values []string, toRemove ...string) []string {
	removed := []string{}
	for _, v := range toRemove {
		if slices.Contains(values, v) && !slices.Contains(removed, v) {
			idx := slices.Index(values, v)
			values = slices.Delete(values, idx, idx+1)
			removed = append(removed, v)
		}
	}
	return values
}

func UpdateContentItemsForupdate(ctx context.Context, smBundle StateMachineBundleAPI, authEntityData *models.AuthEntityData, contentItem *models.ContentItem, ch chan string, wg *sync.WaitGroup, state, bundleTitle string, errCh chan error) {
	defer wg.Done()

	if err := smBundle.DatasetAPIClient.PutVersionState(ctx, authEntityData.Headers, contentItem.Metadata.DatasetID, contentItem.Metadata.EditionID, strconv.Itoa(contentItem.Metadata.VersionID), strings.ToLower(state)); err != nil {
		log.Warn(ctx, fmt.Sprintf("Error occurred transitioning content item for bundle: %s", err.Error()), log.Data{"bundle-id": contentItem.BundleID, "content-item-id": contentItem.ID})
		previewURL := smBundle.PreviewServiceURL + contentItem.Links.Preview

		alarmFields := []slack.Field{
			{Title: "Bundle ID", Value: contentItem.BundleID},
			{Title: "Bundle Title", Value: bundleTitle},
			{Title: "Dataset ID", Value: contentItem.Metadata.DatasetID},
			{Title: "Edition", Value: contentItem.Metadata.EditionID},
			{Title: "Version", Value: strconv.Itoa(contentItem.Metadata.VersionID)},
			{Title: "Preview Link", Value: previewURL},
		}

		_, alarmErr := smBundle.DataBundleSlackClient.SendAlarm(ctx, "Bundle content item failed to update", err, alarmFields)
		if alarmErr != nil {
			log.Error(ctx, "failed to send slack alarm for content item failure", alarmErr, log.Data{"bundle-id": contentItem.BundleID, "content-item-id": contentItem.ID})
		}

		log.Info(ctx, "sending slack alarm for content item failure", log.Data{
			"bundle-id":       contentItem.BundleID,
			"content-item-id": contentItem.ID,
			"alarm_fields":    alarmFields,
		})
		errCh <- err
		return
	}

	if err := smBundle.Datastore.UpdateContentItemState(ctx, contentItem.ID, state); err != nil {
		errCh <- err
		return
	}

	if state == models.BundleStateApproved.String() {
		contentItem.State = models.CastContentItemStateToState(models.BundleStateApproved.String())
	}

	if state == models.BundleStatePublished.String() {
		contentItem.State = models.CastContentItemStateToState(models.BundleStatePublished.String())
	}

	if err := smBundle.CreateEvent(ctx, authEntityData, models.ActionUpdate, nil, contentItem); err != nil {
		log.Error(ctx, "failed to create event", err, log.Data{"bundle_id": contentItem.BundleID, "content_item_id": contentItem.ID})
		errCh <- err
		return
	}

	fmt.Println("CREATED EVENT")

	ch <- contentItem.BundleID
}

func PublishBundle(ctx context.Context, smBundle StateMachineBundleAPI, bundle *models.Bundle, authEntityData *models.AuthEntityData) (*models.Bundle, error) {
	logData := log.Data{"bundle_id": bundle.ID, "bundle_type": bundle.BundleType, "title": bundle.Title}
	contents, err := smBundle.Datastore.GetBundleContentsForBundle(ctx, bundle.ID)
	if err != nil {
		return nil, err
	}

	publishStartTime := time.Now()
	publishLogFields := []slack.Field{
		{Title: "Bundle ID", Value: bundle.ID},
		{Title: "Title", Value: bundle.Title},
		{Title: "Type", Value: bundle.BundleType.String()},
		{Title: "Number of Content Items", Value: strconv.Itoa(len(*contents))},
		{Title: "Publish Start Date", Value: publishStartTime.Format(utils.SlackPublishTimeFormat)},
	}
	logData["slack_fields"] = publishLogFields

	c1 := make(chan *slack.MessageRef)
	go func() {
		log.Info(ctx, "sending slack notification: Bundle publish started", logData)
		slackMessageRef, err := smBundle.DataBundleSlackClient.SendPublishLog(ctx, "Bundle publish started", publishLogFields)
		if err != nil {
			log.Error(ctx, "failed to send slack notification: Bundle publish started", err, logData)
		}
		c1 <- slackMessageRef
	}()

	slackMessageRef := <-c1

	var wg sync.WaitGroup
	ch := make(chan string, len(*contents))
	errCh := make(chan error, len(*contents))

	for index := range *contents {
		contentItem := &(*contents)[index]
		wg.Add(1)
		go UpdateContentItemsForupdate(ctx, smBundle, authEntityData, contentItem, ch, &wg, models.BundleStatePublished.String(), bundle.Title, errCh)
	}

	wg.Wait()

	close(errCh)
	for err := range errCh {
		log.Error(ctx, "something went wrong when processing content items", err, logData)
	}

	bundle.State = models.BundleStatePublished
	bundle.LastUpdatedBy.Email = authEntityData.GetUserEmail()

	updatedBundle, err := smBundle.Datastore.UpdateBundle(ctx, bundle.ID, bundle)
	if err != nil {
		_, err := smBundle.DataBundleSlackClient.SendAlarm(ctx, "Failed to publish bundle", err, publishLogFields)
		if err != nil {
			log.Error(ctx, "failed to send slack notification: Failed to publish bundle", err, logData)
		}
		return nil, err
	}

	publishEndTime := time.Now()
	publishLogFields = append(publishLogFields,
		slack.Field{Title: "Publish End Date", Value: publishEndTime.Format(utils.SlackPublishTimeFormat)},
		slack.Field{Title: "Duration", Value: fmt.Sprintf("%.4f seconds", publishEndTime.Sub(publishStartTime).Seconds())},
	)
	logData["slack_fields"] = publishLogFields

	log.Info(ctx, "updating slack notification: Bundle publish completed", logData)
	_, err = smBundle.DataBundleSlackClient.UpdatePublishLog(ctx, slackMessageRef, "Bundle publish completed", publishLogFields)
	if err != nil {
		log.Error(ctx, "failed to send slack notification: Bundle publish completed", err, logData)
	}

	identityType := log.USER
	if authEntityData.IsServiceAuth {
		identityType = log.SERVICE
	}
	logAuth := log.Auth(identityType, authEntityData.EntityData.UserID)

	if err = smBundle.CreateEvent(ctx, authEntityData, models.ActionUpdate, updatedBundle, nil); err != nil {
		log.Error(ctx, "failed to create event", err, log.Classification(log.ProtectiveMonitoring), logAuth, log.Data{"bundle_id": updatedBundle.ID, "action": models.ActionUpdate})
		return nil, err
	}
	log.Info(ctx, "bundle event creation successful", log.Classification(log.ProtectiveMonitoring), logAuth, log.Data{"bundle_id": updatedBundle.ID, "action": models.ActionUpdate})

	return updatedBundle, nil
}

func ApproveBundle(ctx context.Context, smBundle StateMachineBundleAPI, bundle *models.Bundle, authEntityData *models.AuthEntityData) (*models.Bundle, error) {
	logData := log.Data{"bundle_id": bundle.ID, "bundle_type": bundle.BundleType, "title": bundle.Title}
	contents, err := smBundle.Datastore.GetBundleContentsForBundle(ctx, bundle.ID)
	if err != nil {
		return nil, err
	}

	if contents == nil || len(*contents) == 0 {
		return nil, apierrors.ErrBundleHasNoContentItems
	}

	var wg sync.WaitGroup
	ch := make(chan string, len(*contents))
	errCh := make(chan error, len(*contents))

	for index := range *contents {
		contentItem := &(*contents)[index]
		wg.Add(1)
		//nolint:errcheck // errors are returned in an error channel
		go UpdateContentItemsForupdate(ctx, smBundle, authEntityData, contentItem, ch, &wg, models.BundleStateApproved.String(), bundle.Title, errCh)
	}

	wg.Wait()

	close(errCh)
	for err := range errCh {
		log.Error(ctx, "Something went wrong in processing the content items", err, logData)
	}

	bundle.State = models.BundleStateApproved
	bundle.LastUpdatedBy.Email = authEntityData.GetUserEmail()

	updatedBundle, err := smBundle.updateBundleAndCreateEvent(ctx, bundle, authEntityData, logData)
	if err != nil {
		return nil, err
	}

	return updatedBundle, nil
}

func ReviewBundle(ctx context.Context, smBundle StateMachineBundleAPI, bundle *models.Bundle, authEntityData *models.AuthEntityData) (*models.Bundle, error) {
	logData := log.Data{"bundle_id": bundle.ID, "bundle_type": bundle.BundleType, "title": bundle.Title}

	bundle.State = models.BundleStateInReview
	bundle.LastUpdatedBy.Email = authEntityData.GetUserEmail()

	updatedBundle, err := smBundle.updateBundleAndCreateEvent(ctx, bundle, authEntityData, logData)
	if err != nil {
		return nil, err
	}

	return updatedBundle, nil
}

func DraftBundle(ctx context.Context, smBundle StateMachineBundleAPI, bundle *models.Bundle, authEntityData *models.AuthEntityData) (*models.Bundle, error) {
	logData := log.Data{"bundle_id": bundle.ID, "bundle_type": bundle.BundleType, "title": bundle.Title}

	bundle.State = models.BundleStateDraft
	bundle.LastUpdatedBy.Email = authEntityData.GetUserEmail()

	updatedBundle, err := smBundle.updateBundleAndCreateEvent(ctx, bundle, authEntityData, logData)
	if err != nil {
		return nil, err
	}

	return updatedBundle, nil
}

func (s *StateMachineBundleAPI) updateBundleAndCreateEvent(ctx context.Context, bundle *models.Bundle, authEntityData *models.AuthEntityData, logData log.Data) (*models.Bundle, error) {
	updatedBundle, err := s.Datastore.UpdateBundle(ctx, bundle.ID, bundle)
	if err != nil {
		return nil, err
	}

	log.Info(ctx, "Update bundle completed", logData)

	identityType := log.USER
	if authEntityData.IsServiceAuth {
		identityType = log.SERVICE
	}
	logAuth := log.Auth(identityType, authEntityData.EntityData.UserID)

	if err = s.CreateEvent(ctx, authEntityData, models.ActionUpdate, updatedBundle, nil); err != nil {
		log.Error(ctx, "failed to create event", err, log.Classification(log.ProtectiveMonitoring), logAuth, log.Data{"bundle_id": updatedBundle.ID, "action": models.ActionUpdate})
		return nil, err
	}
	log.Info(ctx, "bundle event creation successful", log.Classification(log.ProtectiveMonitoring), logAuth, log.Data{"bundle_id": updatedBundle.ID, "action": models.ActionUpdate})

	return updatedBundle, nil
}
