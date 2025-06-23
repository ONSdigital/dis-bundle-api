package application

import (
	"context"
	"time"

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

func (s *StateMachineBundleAPI) CreateBundleEvent(ctx context.Context, event *models.Event) error {
	return s.Datastore.CreateBundleEvent(ctx, event)
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
