package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/mongo"
	"github.com/ONSdigital/dis-bundle-api/slack"
	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	MongoDB                 bool
	DatasetAPIClient        bool
	PermissionsAPIClient    bool
	DataBundleSlackClient   bool
	AuthorisationMiddleware bool
	HealthCheck             bool
	Init                    Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		Init: initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetMongoDB creates a mongoDB client and sets the Mongo flag to true
func (e *ExternalServiceList) GetMongoDB(ctx context.Context, cfg config.MongoConfig) (store.MongoDB, error) {
	mongoDB, err := e.Init.DoGetMongoDB(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.MongoDB = true
	return mongoDB, nil
}

// DoGetMongoDB returns a MongoDB
func (e *Init) DoGetMongoDB(ctx context.Context, cfg config.MongoConfig) (store.MongoDB, error) {
	mongodb := &mongo.Mongo{
		MongoConfig: cfg,
	}
	if err := mongodb.Init(ctx); err != nil {
		return nil, err
	}
	log.Info(ctx, "listening to mongo db session", log.Data{"URI": mongodb.ClusterEndpoint})
	return mongodb, nil
}

// GetDatasetAPIClient creates a new Dataset API client with the provided datasetAPIURL and sets the DatasetAPIClient flag to true
func (e *ExternalServiceList) GetDatasetAPIClient(datasetAPIURL string) datasetAPISDK.Clienter {
	client := e.Init.DoGetDatasetAPIClient(datasetAPIURL)
	e.DatasetAPIClient = true
	return client
}

// DoGetDatasetAPIClient returns a new Dataset API client with the provided datasetAPIURL
func (e *Init) DoGetDatasetAPIClient(datasetAPIURL string) datasetAPISDK.Clienter {
	client := datasetAPISDK.New(datasetAPIURL)
	return client
}

// GetPermissionsAPIClient creates a new Permissions API client with the provided permissionsAPIURL and sets the PermissionsAPIClient flag to true
func (e *ExternalServiceList) GetPermissionsAPIClient(permissionsAPIURL string) permissionsAPISDK.Clienter {
	client := e.Init.DoGetPermissionsAPIClient(permissionsAPIURL)
	e.PermissionsAPIClient = true
	return client
}

// DoGetPermissionsAPIClient returns a new Permissions API client with the provided permissionsAPIURL
func (e *Init) DoGetPermissionsAPIClient(permissionsAPIURL string) permissionsAPISDK.Clienter {
	client := permissionsAPISDK.NewClient(permissionsAPIURL)
	return client
}

// GetDataBundleSlackClient creates a new Slack Client and sets the DataBundleSlackClient flag to true
func (e *ExternalServiceList) GetDataBundleSlackClient(slackConfig *slack.SlackConfig, apiToken string, enabled bool) (slack.Clienter, error) {
	client, err := e.Init.DoGetDataBundleSlackClient(slackConfig, apiToken, enabled)
	if err != nil {
		return nil, err
	}
	e.DataBundleSlackClient = true
	return client, nil
}

// DoGetDataBundleSlackClient creates a new Slack Client
// If enabled is false, a no-op slack client is returned
func (e *Init) DoGetDataBundleSlackClient(slackConfig *slack.SlackConfig, apiToken string, enabled bool) (slack.Clienter, error) {
	return slack.New(slackConfig, apiToken, enabled)
}

// GetAuthorisationMiddleware creates authorisation middleware for the given config and sets the AuthorisationMiddleware flag to true
func (e *ExternalServiceList) GetAuthorisationMiddleware(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
	authMiddleware, err := e.Init.DoGetAuthorisationMiddleware(ctx, authorisationConfig)
	if err != nil {
		return nil, err
	}
	e.AuthorisationMiddleware = true
	return authMiddleware, nil
}

// DoGetAuthorisationMiddleware creates authorisation middleware for the given config
func (e *Init) DoGetAuthorisationMiddleware(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
	return authorisation.NewFeatureFlaggedMiddleware(ctx, authorisationConfig, nil)
}

// GetHealthCheck creates a healthcheck with versionInfo and sets the HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// DoGetHealthCheck creates a healthcheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

// GetHTTPServer creates an http server
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	return s
}

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}
