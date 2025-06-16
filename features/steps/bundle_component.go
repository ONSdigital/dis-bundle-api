package steps

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ONSdigital/dis-bundle-api/api"
	"github.com/ONSdigital/dis-bundle-api/auth"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/mongo"
	"github.com/ONSdigital/dis-bundle-api/service"
	serviceMock "github.com/ONSdigital/dis-bundle-api/service/mock"
	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-component-test/utils"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	permissionsSDK "github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

var mockDatasetVersions = []*datasetAPIModels.Version{
	{ID: "1"},
	{ID: "2"},
	{ID: "inreview-version", State: "IN_REVIEW"},
	{ID: "approved-version", State: "APPROVED"},
	{ID: "published-version", State: "PUBLISHED"},
}

type BundleComponent struct {
	ErrorFeature            componenttest.ErrorFeature
	svc                     *service.Service
	errorChan               chan error
	MongoClient             *mongo.Mongo
	Config                  *config.Config
	HTTPServer              *http.Server
	ServiceRunning          bool
	initialiser             service.Initialiser
	datasetAPIClient        datasetAPISDK.Clienter
	apiFeature              *componenttest.APIFeature
	AuthorisationMiddleware auth.AuthorisationMiddleware
	DatasetAPIVersions      []*datasetAPIModels.Version
}

func NewBundleComponent(mongoURI string) (*BundleComponent, error) {
	c := &BundleComponent{
		HTTPServer: &http.Server{
			ReadHeaderTimeout: 60 * time.Second,
		},
		errorChan:      make(chan error),
		ServiceRunning: false,
	}

	var err error

	c.Config, err = config.Get()
	if err != nil {
		return nil, err
	}

	log.Info(context.Background(), "configuration for component test", log.Data{"config": c.Config})

	fakePermissionsAPI := setupFakePermissionsAPI()
	c.Config.AuthConfig.PermissionsAPIURL = fakePermissionsAPI.URL()

	c.initialiser = &serviceMock.InitialiserMock{
		DoGetMongoDBFunc:                 c.DoGetMongoDB,
		DoGetDatasetAPIClientFunc:        c.DoGetDatasetAPIClient,
		DoGetHealthCheckFunc:             c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:              c.DoGetHTTPServer,
		DoGetAuthorisationMiddlewareFunc: c.DoGetAuthorisationMiddleware,
	}

	mongodb := &mongo.Mongo{
		MongoConfig: config.MongoConfig{
			MongoDriverConfig: mongodriver.MongoDriverConfig{
				ClusterEndpoint: mongoURI,
				Database:        utils.RandomDatabase(),
				Collections:     c.Config.Collections,
				ConnectTimeout:  c.Config.ConnectTimeout,
				QueryTimeout:    c.Config.QueryTimeout,
			},
		}}

	if err := mongodb.Init(context.Background()); err != nil {
		return nil, err
	}

	c.ServiceRunning = true
	c.MongoClient = mongodb
	c.apiFeature = componenttest.NewAPIFeature(c.InitialiseService)
	c.DatasetAPIVersions = mockDatasetVersions
	return c, nil
}

func setupFakePermissionsAPI() *authorisationtest.FakePermissionsAPI {
	fakePermissionsAPI := authorisationtest.NewFakePermissionsAPI()
	bundle := getPermissionsBundle()
	fakePermissionsAPI.Reset()
	if err := fakePermissionsAPI.UpdatePermissionsBundleResponse(bundle); err != nil {
		log.Error(context.Background(), "failed to update permissions bundle response", err)
	}
	return fakePermissionsAPI
}

func getPermissionsBundle() *permissionsSDK.Bundle {
	return &permissionsSDK.Bundle{
		api.AuthRoleBundlesRead: { // role
			"groups/role-admin": { // group
				{
					ID: "1", // policy
				},
			},
		},
		api.AuthRoleBundlesCreate: {
			"groups/role-admin": {
				{
					ID: "1",
				},
			},
		},
		api.AuthRoleBundlesUpdate: {
			"groups/role-admin": {
				{
					ID: "1",
				},
			},
		},
	}
}

func (c *BundleComponent) Reset() error {
	ctx := context.Background()

	c.MongoClient.Database = utils.RandomDatabase()
	if err := c.MongoClient.Init(ctx); err != nil {
		log.Warn(ctx, "error initialising MongoClient during Reset", log.Data{"err": err.Error()})
	}

	c.setInitialiserMock()

	return nil
}

func (c *BundleComponent) InitialiseService() (http.Handler, error) {
	// Initialiser before Run to allow switching out of Initialiser between tests.
	c.svc = service.New(c.Config, service.NewServiceList(c.initialiser))

	if err := c.svc.Run(context.Background(), "1", "", "", c.errorChan); err != nil {
		return nil, err
	}
	c.ServiceRunning = true
	return c.HTTPServer.Handler, nil
}

func (c *BundleComponent) DoGetHealthcheckOk(*config.Config, string, string, string) (service.HealthChecker, error) {
	return &serviceMock.HealthCheckerMock{
		AddCheckFunc: func(string, healthcheck.Checker) error { return nil },
		StartFunc:    func(context.Context) {},
		StopFunc:     func() {},
	}, nil
}

func (c *BundleComponent) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	c.HTTPServer.Addr = bindAddr
	c.HTTPServer.Handler = router
	return c.HTTPServer
}

func (c *BundleComponent) DoGetMongoDB(context.Context, config.MongoConfig) (store.MongoDB, error) {
	return c.MongoClient, nil
}

func (c *BundleComponent) DoGetDatasetAPIClient(datasetAPIURL string) datasetAPISDK.Clienter {
	datasetAPIClient := &datasetAPISDKMock.ClienterMock{
		GetVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID string, versionID string) (datasetAPIModels.Version, error) {
			versionVersion, err := strconv.Atoi(versionID)
			if err != nil {
				return datasetAPIModels.Version{}, err
			}
			for _, version := range c.DatasetAPIVersions {
				if version.DatasetID == datasetID && version.Edition == editionID && version.Version == versionVersion {
					return *version, nil
				}
			}

			if datasetID == "dataset-not-found" {
				return datasetAPIModels.Version{}, errors.New("dataset not found")
			}
			if editionID == "edition-not-found" {
				return datasetAPIModels.Version{}, errors.New("edition not found")
			}
			if versionID == "404" {
				return datasetAPIModels.Version{}, errors.New("version not found")
			}

			return datasetAPIModels.Version{}, errors.New("version not found")
		},
		PutVersionStateFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID, state string) error {
			versionVersion, err := strconv.Atoi(versionID)
			if err != nil {
				return err
			}
			for _, version := range c.DatasetAPIVersions {
				if version.DatasetID == datasetID && version.Edition == editionID && version.Version == versionVersion {
					version.State = state
					return nil
				}
			}

			return fmt.Errorf("version %s not found for dataset %s edition %s", versionID, datasetID, editionID)
		},
	}
	c.datasetAPIClient = datasetAPIClient
	return c.datasetAPIClient
}

func (c *BundleComponent) DoGetAuthorisationMiddleware(ctx context.Context, cfg *authorisation.Config) (auth.AuthorisationMiddleware, error) {
	cfg.Enabled = true
	middleware, err := auth.CreateAuthorisationMiddlewareFromConfig(ctx, cfg, true)
	if err != nil {
		return nil, err
	}

	c.AuthorisationMiddleware = middleware
	return c.AuthorisationMiddleware, nil
}

func (c *BundleComponent) setInitialiserMock() {
	c.initialiser = &serviceMock.InitialiserMock{
		DoGetMongoDBFunc:                 c.DoGetMongoDB,
		DoGetDatasetAPIClientFunc:        c.DoGetDatasetAPIClient,
		DoGetHealthCheckFunc:             c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:              c.DoGetHTTPServer,
		DoGetAuthorisationMiddlewareFunc: c.DoGetAuthorisationMiddleware,
	}
}

func (c *BundleComponent) Close() error {
	ctx := context.Background()

	// Closing Mongo DB
	if c.svc != nil && c.ServiceRunning {
		if err := c.MongoClient.Connection.DropDatabase(ctx); err != nil {
			log.Warn(ctx, "error dropping database on Close", log.Data{"err": err.Error()})
		}
		if err := c.svc.Close(ctx); err != nil {
			return err
		}
		c.ServiceRunning = false
	}
	return nil
}
