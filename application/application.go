package application

import (
	"context"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/filters"
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

func (s *StateMachineBundleAPI) CreateBundleEvent(ctx context.Context, event *models.Event) error {
	return s.Datastore.CreateBundleEvent(ctx, event)
}

func (s *StateMachineBundleAPI) CreateBundle(ctx context.Context, bundle *models.Bundle) error {
	return s.Datastore.CreateBundle(ctx, bundle)
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
