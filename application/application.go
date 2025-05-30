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
	results, totalCount, err := s.Datastore.ListBundles(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	return results, totalCount, nil
}

func (s *StateMachineBundleAPI) CheckAllBundleContentsAreApproved(ctx context.Context, bundleID string) (bool, error) {
	return s.Datastore.CheckAllBundleContentsAreApproved(ctx, bundleID)
}

func (s *StateMachineBundleAPI) CreateBundle(ctx context.Context, bundle *models.Bundle) (*models.Bundle, error) {
	err := s.Datastore.CreateBundle(ctx, bundle)
	if err != nil {
		return nil, err
	}
	createdBundle, err := s.Datastore.GetBundle(ctx, bundle.ID)
	if err != nil {
		return nil, err
	}
	return createdBundle, nil
}

func (s *StateMachineBundleAPI) GetBundleByTitle(ctx context.Context, title string) (*models.Bundle, error) {
	bundle, err := s.Datastore.GetBundleByTitle(ctx, title)
	if err != nil {
		return nil, err
	}
	return bundle, nil
}

func (s *StateMachineBundleAPI) CreateBundleEvent(ctx context.Context, event *models.Event) error {
	err := s.Datastore.CreateBundleEvent(ctx, event)
	if err != nil {
		return err
	}
	return nil
}
