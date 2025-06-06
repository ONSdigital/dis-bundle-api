package application

import (
	"context"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
)

type StateMachineBundleAPI struct {
	Datastore    store.Datastore
	StateMachine *StateMachine
}

func Setup(datastore store.Datastore, stateMachine *StateMachine) *StateMachineBundleAPI {
	return &StateMachineBundleAPI{
		Datastore:    datastore,
		StateMachine: stateMachine,
	}
}

func (s *StateMachineBundleAPI) ListBundles(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error) {
	return s.Datastore.ListBundles(ctx, offset, limit)
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

func (s *StateMachineBundleAPI) CheckAllBundleContentsAreApproved(ctx context.Context, bundleID string) (bool, error) {
	return s.Datastore.CheckAllBundleContentsAreApproved(ctx, bundleID)
}

func (s *StateMachineBundleAPI) CheckContentItemExistsByDatasetEditionVersion(ctx context.Context, datasetID, editionID string, versionID int) (bool, error) {
	return s.Datastore.CheckContentItemExistsByDatasetEditionVersion(ctx, datasetID, editionID, versionID)
}

func (s *StateMachineBundleAPI) CreateBundleEvent(ctx context.Context, event *models.Event) error {
	return s.Datastore.CreateBundleEvent(ctx, event)
}
