package application

import (
	"context"

	"github.com/ONSdigital/dis-bundle-api/models"
)

type BundleStore interface {
	ListBundles(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error)
}

type BundleService struct {
	Store BundleStore
}

func NewBundleService(store BundleStore) *BundleService {
	return &BundleService{
		Store: store,
	}
}

func (s *BundleService) ListBundles(ctx context.Context, offset, limit int) ([]*models.Bundle, int, error) {
	return s.Store.ListBundles(ctx, offset, limit)
}
