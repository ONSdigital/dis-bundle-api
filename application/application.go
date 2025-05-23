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

func checkAllBundleContentsAreApproved(contents []models.BundleContent) bool {
	for _, bundleContent := range contents {
		if bundleContent.State != Approved.String() {
			return false
		}
	}
	return true
}

type Bundlestore interface {
	ListBundles(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error)
}

type BundleService struct {
	Store Bundlestore
}

func NewBundleService(store Bundlestore) *BundleService {
	return &BundleService{
		Store: store,
	}
}

func (s *BundleService) ListBundles(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error) {
	return s.Store.ListBundles(ctx, offset, limit)
}
