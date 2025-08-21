package sdk

import (
	"context"
	"time"

	"github.com/ONSdigital/dis-bundle-api/models"
	apiError "github.com/ONSdigital/dis-bundle-api/sdk/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out ./mocks/client.go -pkg mocks . Clienter

type Clienter interface {
	Checker(ctx context.Context, check *healthcheck.CheckState) error
	Health() *health.Client
	URL() string
	GetBundles(ctx context.Context, headers Headers, scheduledAt *time.Time, queryParams *QueryParams) (*BundlesList, apiError.Error)
	GetBundle(ctx context.Context, headers Headers, id string) (*ResponseInfo, apiError.Error)
	PutBundleState(ctx context.Context, headers Headers, id string, state models.BundleState) (*models.Bundle, apiError.Error)
}
