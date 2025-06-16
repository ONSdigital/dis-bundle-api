package eventsmocks

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/events"
	"github.com/ONSdigital/dis-bundle-api/models"
)

func CreateSuccessMockBundleEventsManager() events.BundleEventsManager {
	return &BundleEventsManagerMock{
		CreateBundleUpdatedEventFunc: func(ctx context.Context, r *http.Request, bundle *models.Bundle) *models.Error {
			return nil
		},
	}
}

func CreateErrorMockBundleEventsManager(err models.Error) events.BundleEventsManager {
	return &BundleEventsManagerMock{
		CreateBundleUpdatedEventFunc: func(ctx context.Context, r *http.Request, bundle *models.Bundle) *models.Error {
			return &err
		},
	}
}
