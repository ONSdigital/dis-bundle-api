package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/slack"
	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/healthCheck.go -pkg mock . HealthChecker

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetMongoDB(ctx context.Context, cfg config.MongoConfig) (store.MongoDB, error)
	DoGetDatasetAPIClient(datasetAPIURL string) datasetAPISDK.Clienter
	DoGetPermissionsAPIClient(permissionsAPIURL string) permissionsAPISDK.Clienter
	DoGetDataBundleSlackClient(slackConfig *slack.SlackConfig, apiToken string, enabled bool) (slack.Clienter, error)
	DoGetAuthorisationMiddleware(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error)
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error)
	DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer
}

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// HealthChecker defines the required methods from Healthcheck
type HealthChecker interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}
