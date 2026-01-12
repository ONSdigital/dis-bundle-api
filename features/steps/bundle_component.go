package steps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/mongo"
	"github.com/ONSdigital/dis-bundle-api/service"
	serviceMock "github.com/ONSdigital/dis-bundle-api/service/mock"
	"github.com/ONSdigital/dis-bundle-api/slack"
	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-component-test/utils"
	datasetAPIModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	permissionsAPIModels "github.com/ONSdigital/dp-permissions-api/models"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	permissionsAPISDKMock "github.com/ONSdigital/dp-permissions-api/sdk/mocks"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	datasetNotFound = "dataset-not-found"
	getMethod       = "GET"
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
	permissionsAPIClient    permissionsAPISDK.Clienter
	permissionsAPIPolicies  []*permissionsAPIModels.Policy
	permissionsAPIServer    *httptest.Server
	slackClient             slack.Clienter
	apiFeature              *componenttest.APIFeature
	AuthorisationMiddleware authorisation.Middleware
	DatasetAPIVersions      []*datasetAPIModels.Version
}

func NewBundleComponent(mongoURI string) (*BundleComponent, error) {
	c := &BundleComponent{
		HTTPServer: &http.Server{
			ReadHeaderTimeout: 60 * time.Second,
		},
		errorChan:              make(chan error),
		ServiceRunning:         false,
		permissionsAPIPolicies: []*permissionsAPIModels.Policy{},
	}

	var err error

	c.Config, err = config.Get()
	if err != nil {
		return nil, err
	}

	log.Info(context.Background(), "configuration for component test", log.Data{"config": c.Config})

	c.permissionsAPIServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == getMethod && r.URL.Path == "/v1/permissions-bundle" {
			bundle := getPermissionsBundle()
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(bundle); err != nil {
				log.Error(context.Background(), "failed to encode permissions bundle", err)
			}
			return
		}
		if r.Method == getMethod && strings.HasPrefix(r.URL.Path, "/v1/policies/") {
			policyID := strings.TrimPrefix(r.URL.Path, "/v1/policies/")
			for _, p := range c.permissionsAPIPolicies {
				if p.ID == policyID {
					w.WriteHeader(http.StatusOK)
					if err := json.NewEncoder(w).Encode(p); err != nil {
						log.Error(context.Background(), "failed to encode policy", err)
					}
					return
				}
			}
			policy := &permissionsAPIModels.Policy{ID: policyID}
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(policy); err != nil {
				log.Error(context.Background(), "failed to encode empty policy", err)
			}
			return
		}
		if r.Method == "PUT" && strings.HasPrefix(r.URL.Path, "/v1/policies/") {
			policyID := strings.TrimPrefix(r.URL.Path, "/v1/policies/")
			var policy permissionsAPIModels.Policy
			if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			policy.ID = policyID
			found := false
			for i, p := range c.permissionsAPIPolicies {
				if p.ID == policyID {
					c.permissionsAPIPolicies[i] = &policy
					found = true
					break
				}
			}
			if !found {
				c.permissionsAPIPolicies = append(c.permissionsAPIPolicies, &policy)
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}))

	c.Config.AuthConfig.PermissionsAPIURL = c.permissionsAPIServer.URL

	c.initialiser = &serviceMock.InitialiserMock{
		DoGetMongoDBFunc:                 c.DoGetMongoDB,
		DoGetDatasetAPIClientFunc:        c.DoGetDatasetAPIClient,
		DoGetPermissionsAPIClientFunc:    c.DoGetPermissionsAPIClient,
		DoGetDataBundleSlackClientFunc:   c.DoGetDataBundleSlackClient,
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

func getPermissionsBundle() *permissionsAPISDK.Bundle {
	return &permissionsAPISDK.Bundle{
		"bundles:read": {
			"groups/role-admin": {
				{
					ID: "1",
				},
			},
		},
		"bundles:create": {
			"groups/role-admin": {
				{
					ID: "1",
				},
			},
		},
		"bundles:update": {
			"groups/role-admin": {
				{
					ID: "1",
				},
			},
		},
		"bundles:delete": {
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
			if datasetID == datasetNotFound {
				return datasetAPIModels.Version{}, errors.New("dataset not found")
			}
			if editionID == "edition-not-found" {
				return datasetAPIModels.Version{}, errors.New("edition not found")
			}
			if versionID == "404" {
				return datasetAPIModels.Version{}, errors.New("version not found")
			}

			versionVersion, err := strconv.Atoi(versionID)
			if err != nil {
				return datasetAPIModels.Version{}, err
			}
			for _, version := range c.DatasetAPIVersions {
				if version.DatasetID == datasetID && version.Edition == editionID && version.Version == versionVersion {
					return *version, nil
				}
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
		PutVersionFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID, editionID, versionID string, version datasetAPIModels.Version) (datasetAPIModels.Version, error) {
			versionAsString, err := strconv.Atoi(versionID)
			if err != nil {
				return datasetAPIModels.Version{}, err
			}
			for _, versionInDatastore := range c.DatasetAPIVersions {
				if versionInDatastore.DatasetID == datasetID && versionInDatastore.Edition == editionID && versionInDatastore.Version == versionAsString {
					if version.ReleaseDate != "" {
						versionInDatastore.ReleaseDate = version.ReleaseDate
					}
				}
			}
			return version, nil
		},
		GetDatasetFunc: func(ctx context.Context, headers datasetAPISDK.Headers, datasetID string) (datasetAPIModels.Dataset, error) {
			if datasetID == "dataset-id-does-not-exist" {
				return datasetAPIModels.Dataset{}, errors.New("dataset not found")
			}
			if datasetID == "dataset1" {
				return datasetAPIModels.Dataset{State: "edition-confirmed", Title: "Test Dataset title"}, nil
			}
			return datasetAPIModels.Dataset{}, nil
		},
	}
	c.datasetAPIClient = datasetAPIClient
	return c.datasetAPIClient
}

func (c *BundleComponent) DoGetPermissionsAPIClient(permissionsAPIURL string) permissionsAPISDK.Clienter {
	permissionsAPIClient := &permissionsAPISDKMock.ClienterMock{
		GetPolicyFunc: func(ctx context.Context, id string, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
			for _, policy := range c.permissionsAPIPolicies {
				if policy.ID == id {
					return policy, nil
				}
			}
			return nil, errors.New("404 Not Found")
		},
		PostPolicyWithIDFunc: func(ctx context.Context, id string, policy permissionsAPIModels.PolicyInfo, headers permissionsAPISDK.Headers) (*permissionsAPIModels.Policy, error) {
			createdPolicy := &permissionsAPIModels.Policy{
				ID:        id,
				Entities:  policy.Entities,
				Role:      policy.Role,
				Condition: policy.Condition,
			}
			c.permissionsAPIPolicies = append(c.permissionsAPIPolicies, createdPolicy)
			return createdPolicy, nil
		},
		PutPolicyFunc: func(ctx context.Context, id string, policy permissionsAPIModels.Policy, headers permissionsAPISDK.Headers) error {
			for i, p := range c.permissionsAPIPolicies {
				if p.ID == id {
					c.permissionsAPIPolicies[i] = &policy
					return nil
				}
			}
			c.permissionsAPIPolicies = append(c.permissionsAPIPolicies, &policy)
			return nil
		},
	}
	c.permissionsAPIClient = permissionsAPIClient
	return c.permissionsAPIClient
}

func (c *BundleComponent) DoGetDataBundleSlackClient(slackConfig *slack.SlackConfig, apiToken string, enabled bool) (slack.Clienter, error) {
	c.slackClient = &slack.NoopClient{}
	return c.slackClient, nil
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
		DoGetDatasetAPIClientFunc:        c.DoGetDatasetAPIClient,
		DoGetPermissionsAPIClientFunc:    c.DoGetPermissionsAPIClient,
		DoGetDataBundleSlackClientFunc:   c.DoGetDataBundleSlackClient,
		DoGetHealthCheckFunc:             c.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:              c.DoGetHTTPServer,
		DoGetAuthorisationMiddlewareFunc: c.DoGetAuthorisationMiddleware,
	}
}

func (c *BundleComponent) Close() error {
	ctx := context.Background()

	if c.permissionsAPIServer != nil {
		c.permissionsAPIServer.Close()
	}

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
