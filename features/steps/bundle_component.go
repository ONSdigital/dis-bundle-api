package steps

import (
	"context"
	"net/http"
	"time"

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/mongo"
	"github.com/ONSdigital/dis-bundle-api/service"
	serviceMock "github.com/ONSdigital/dis-bundle-api/service/mock"
	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/dp-authorisation/v2/authorisationtest"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-component-test/utils"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"github.com/ONSdigital/dp-permissions-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

type BundleComponent struct {
	ErrorFeature            componenttest.ErrorFeature
	svc                     *service.Service
	errorChan               chan error
	MongoClient             *mongo.Mongo
	Config                  *config.Config
	HTTPServer              *http.Server
	ServiceRunning          bool
	initialiser             service.Initialiser
	apiFeature              *componenttest.APIFeature
	AuthorisationMiddleware authorisation.Middleware
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

func getPermissionsBundle() *sdk.Bundle {
	return &sdk.Bundle{
		"bundles:read": { // role
			"groups/role-admin": { // group
				{
					ID: "1", // policy
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

func (c *BundleComponent) DoGetAuthorisationMiddleware(ctx context.Context, cfg *authorisation.Config) (authorisation.Middleware, error) {
	middleware, err := authorisation.NewMiddlewareFromConfig(ctx, cfg, cfg.JWTVerificationPublicKeys)
	if err != nil {
		return nil, err
	}

	c.AuthorisationMiddleware = middleware
	return c.AuthorisationMiddleware, nil
}

func (c *BundleComponent) setInitialiserMock() {
	c.initialiser = &serviceMock.InitialiserMock{
		DoGetMongoDBFunc:                 c.DoGetMongoDB,
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
