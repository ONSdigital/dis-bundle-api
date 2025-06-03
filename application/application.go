package application

import (
	"context"
	"net/http"
	"time"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/datasets"
	"github.com/ONSdigital/dis-bundle-api/events"
	"github.com/ONSdigital/dis-bundle-api/filters"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/store"
)

type StateMachineBundleAPI struct {
	Datastore      store.Datastore
	Events         events.BundleEventsManager
	StateMachine   *StateMachine
	DatasetsClient datasets.DatasetsClient
}

func Setup(datastore store.Datastore, stateMachine *StateMachine, datasetsClient datasets.DatasetsClient, eventsManager events.BundleEventsManager) *StateMachineBundleAPI {
	return &StateMachineBundleAPI{
		Datastore:      datastore,
		StateMachine:   stateMachine,
		DatasetsClient: datasetsClient,
		Events:         eventsManager,
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

func (s *StateMachineBundleAPI) UpdateBundleState(ctx context.Context, r *http.Request, bundleID string, state models.BundleState, suppliedEtag string) (*models.Bundle, *models.Error) {
	bundle, err := s.getBundleAndValidateETag(ctx, bundleID, suppliedEtag)
	if err != nil {
		return nil, err
	}

	transitionError := s.StateMachine.Transition(ctx, s, r, bundle, state)

	if transitionError != nil {
		return nil, transitionError
	}

	bundle, err = s.getBundle(ctx, bundleID)

	if err != nil {
		return nil, models.CreateModelError(models.CodeInternalServerError, apierrors.ErrorDescriptionInternalError)
	}

	return bundle, nil
}

func (s *StateMachineBundleAPI) getBundleAndValidateETag(ctx context.Context, bundleID string, suppliedEtag string) (*models.Bundle, *models.Error) {
	bundle, err := s.getBundle(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	err = s.validateETag(bundle, suppliedEtag)
	if err != nil {
		return nil, err
	}

	return bundle, nil
}

func (s *StateMachineBundleAPI) getBundle(ctx context.Context, bundleID string) (*models.Bundle, *models.Error) {
	bundle, err := s.Datastore.GetBundle(ctx, bundleID)

	if err != nil {
		return nil, models.CreateModelError(getCodeForError(err), err.Error())
	}

	if bundle == nil {
		return nil, models.CreateModelError(models.CodeNotFound, apierrors.ErrorDescriptionNotFound)
	}

	return bundle, nil
}

func getCodeForError(err error) models.Code {
	if err == apierrors.ErrBundleNotFound {
		return models.CodeNotFound
	} else {
		return models.CodeInternalServerError
	}
}

func (*StateMachineBundleAPI) validateETag(bundle *models.Bundle, suppliedEtag string) *models.Error {
	etagMatches := bundle.ETag == suppliedEtag
	if !etagMatches {
		return models.CreateModelError(models.CodeConflict, apierrors.ErrorDescriptionETagMismatch)
	}

	return nil
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
