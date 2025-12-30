package service_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/service"
	serviceMock "github.com/ONSdigital/dis-bundle-api/service/mock"
	"github.com/ONSdigital/dis-bundle-api/slack"
	slackMock "github.com/ONSdigital/dis-bundle-api/slack/mocks"
	"github.com/ONSdigital/dis-bundle-api/store"
	storeMock "github.com/ONSdigital/dis-bundle-api/store/datastoretest"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	authorisationMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	datasetAPISDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	permissionsAPISDK "github.com/ONSdigital/dp-permissions-api/sdk"
	permissionsAPISDKMock "github.com/ONSdigital/dp-permissions-api/sdk/mocks"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx           = context.Background()
	testBuildTime = "BuildTime"
	testGitCommit = "GitCommit"
	testVersion   = "Version"
)

var (
	errMongo                 = errors.New("MongoDB error")
	errDataBundleSlackClient = errors.New("Data Bundle Slack Client error")
	errAuthMiddleware        = errors.New("Authorisation Middleware error")
	errHealthcheck           = errors.New("healthCheck error")
	errServer                = errors.New("HTTP Server error")
)

var funcDoGetMongoDBErr = func(ctx context.Context, cfg config.MongoConfig) (store.MongoDB, error) {
	return nil, errMongo
}

func funcDoGetDataBundleSlackClientErr(slackConfig *slack.SlackConfig, apiToken string, enabled bool) (slack.Clienter, error) {
	return nil, errDataBundleSlackClient
}

var funcDoGetAuthMiddlewareErr = func(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
	return nil, errAuthMiddleware
}

var funcDoGetHealthcheckErr = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	return nil, errHealthcheck
}

func TestRun(t *testing.T) {
	Convey("Having a set of mocked dependencies", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		authorisationMiddleware := &authorisationMock.MiddlewareMock{
			RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
				return handlerFunc
			},
			CloseFunc: func(ctx context.Context) error {
				return nil
			},
		}

		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(string, healthcheck.Checker) error { return nil },
			StartFunc:    func(context.Context) {},
		}

		serverWg := &sync.WaitGroup{}

		serverMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return nil
			},
		}

		failingServerMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return errServer
			},
		}

		funcDoGetMongoDBOk := func(context.Context, config.MongoConfig) (store.MongoDB, error) {
			return &storeMock.MongoDBMock{}, nil
		}

		funcDoGetDatasetAPIClientOk := func(datasetAPIURL string) datasetAPISDK.Clienter {
			return &datasetAPISDKMock.ClienterMock{}
		}

		funcDoGetPermissionsAPIClientOk := func(permissionsAPIURL string) permissionsAPISDK.Clienter {
			return &permissionsAPISDKMock.ClienterMock{}
		}

		funcDoGetDataBundleSlackClientOk := func(slackConfig *slack.SlackConfig, apiToken string, enabled bool) (slack.Clienter, error) {
			return &slackMock.ClienterMock{}, nil
		}

		funcDoGetAuthMiddlewareOk := func(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
			return authorisationMiddleware, nil
		}

		funcDoGetHealthcheckOk := func(*config.Config, string, string, string) (service.HealthChecker, error) {
			return hcMock, nil
		}

		funcDoGetHTTPServerOk := func(string, http.Handler) service.HTTPServer {
			return serverMock
		}

		funcDoGetFailingHTTPServer := func(string, http.Handler) service.HTTPServer {
			return failingServerMock
		}

		Convey("Given that initialising MongoDB returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetMongoDBFunc: funcDoGetMongoDBErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errMongo)
				So(svcList.MongoDB, ShouldBeFalse)
				So(svcList.DatasetAPIClient, ShouldBeFalse)
				So(svcList.PermissionsAPIClient, ShouldBeFalse)
				So(svcList.DataBundleSlackClient, ShouldBeFalse)
				So(svcList.AuthorisationMiddleware, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising DataBundleSlackClient returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetMongoDBFunc:               funcDoGetMongoDBOk,
				DoGetDatasetAPIClientFunc:      funcDoGetDatasetAPIClientOk,
				DoGetPermissionsAPIClientFunc:  funcDoGetPermissionsAPIClientOk,
				DoGetDataBundleSlackClientFunc: funcDoGetDataBundleSlackClientErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errDataBundleSlackClient)
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.DatasetAPIClient, ShouldBeTrue)
				So(svcList.PermissionsAPIClient, ShouldBeTrue)
				So(svcList.DataBundleSlackClient, ShouldBeFalse)
				So(svcList.AuthorisationMiddleware, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising Authorisation Middleware returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetMongoDBFunc:                 funcDoGetMongoDBOk,
				DoGetDatasetAPIClientFunc:        funcDoGetDatasetAPIClientOk,
				DoGetPermissionsAPIClientFunc:    funcDoGetPermissionsAPIClientOk,
				DoGetDataBundleSlackClientFunc:   funcDoGetDataBundleSlackClientOk,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthMiddlewareErr,
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errAuthMiddleware)
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.DatasetAPIClient, ShouldBeTrue)
				So(svcList.PermissionsAPIClient, ShouldBeTrue)
				So(svcList.DataBundleSlackClient, ShouldBeTrue)
				So(svcList.AuthorisationMiddleware, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising Healthcheck returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetMongoDBFunc:                 funcDoGetMongoDBOk,
				DoGetDatasetAPIClientFunc:        funcDoGetDatasetAPIClientOk,
				DoGetPermissionsAPIClientFunc:    funcDoGetPermissionsAPIClientOk,
				DoGetDataBundleSlackClientFunc:   funcDoGetDataBundleSlackClientOk,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthMiddlewareOk,
				DoGetHealthCheckFunc:             funcDoGetHealthcheckErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set. No further initialisations are attempted", func() {
				So(err, ShouldResemble, errHealthcheck)
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.DatasetAPIClient, ShouldBeTrue)
				So(svcList.PermissionsAPIClient, ShouldBeTrue)
				So(svcList.DataBundleSlackClient, ShouldBeTrue)
				So(svcList.AuthorisationMiddleware, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetMongoDBFunc:                 funcDoGetMongoDBOk,
				DoGetDatasetAPIClientFunc:        funcDoGetDatasetAPIClientOk,
				DoGetPermissionsAPIClientFunc:    funcDoGetPermissionsAPIClientOk,
				DoGetDataBundleSlackClientFunc:   funcDoGetDataBundleSlackClientOk,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthMiddlewareOk,
				DoGetHealthCheckFunc:             funcDoGetHealthcheckOk,
				DoGetHTTPServerFunc:              funcDoGetHTTPServerOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			serverWg.Add(1)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run succeeds and all the flags are set", func() {
				So(err, ShouldBeNil)
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.DatasetAPIClient, ShouldBeTrue)
				So(svcList.PermissionsAPIClient, ShouldBeTrue)
				So(svcList.DataBundleSlackClient, ShouldBeTrue)
				So(svcList.AuthorisationMiddleware, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
			})

			Convey("And the checkers are registered and the healthcheck and http server started", func() {
				So(len(hcMock.AddCheckCalls()), ShouldEqual, 3)
				So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Mongo DB")
				So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Zebedee")
				So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Dataset API Client")
				So(len(initMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
				So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, ":29800")
				So(len(hcMock.StartCalls()), ShouldEqual, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})

		Convey("Given that Checkers cannot be registered", func() {
			errAddCheckFail := errors.New("Error(s) registering checkers for healthcheck")
			hcMockAddFail := &serviceMock.HealthCheckerMock{
				AddCheckFunc: func(string, healthcheck.Checker) error { return errAddCheckFail },
				StartFunc:    func(context.Context) {},
			}

			initMock := &serviceMock.InitialiserMock{
				DoGetMongoDBFunc:                 funcDoGetMongoDBOk,
				DoGetDatasetAPIClientFunc:        funcDoGetDatasetAPIClientOk,
				DoGetPermissionsAPIClientFunc:    funcDoGetPermissionsAPIClientOk,
				DoGetDataBundleSlackClientFunc:   funcDoGetDataBundleSlackClientOk,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthMiddlewareOk,
				DoGetHealthCheckFunc: func(*config.Config, string, string, string) (service.HealthChecker, error) {
					return hcMockAddFail, nil
				},
				DoGetHTTPServerFunc: funcDoGetHTTPServerOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails, but all checks try to register", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, fmt.Sprintf("unable to register checkers: %s", errAddCheckFail.Error()))
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
				So(len(hcMockAddFail.AddCheckCalls()), ShouldEqual, 3)
				So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "Mongo DB")
				So(hcMockAddFail.AddCheckCalls()[1].Name, ShouldResemble, "Zebedee")
				So(hcMockAddFail.AddCheckCalls()[2].Name, ShouldResemble, "Dataset API Client")
			})
		})

		Convey("Given that all dependencies are successfully initialised but the http server fails", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetMongoDBFunc:                 funcDoGetMongoDBOk,
				DoGetDatasetAPIClientFunc:        funcDoGetDatasetAPIClientOk,
				DoGetPermissionsAPIClientFunc:    funcDoGetPermissionsAPIClientOk,
				DoGetDataBundleSlackClientFunc:   funcDoGetDataBundleSlackClientOk,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthMiddlewareOk,
				DoGetHealthCheckFunc:             funcDoGetHealthcheckOk,
				DoGetHTTPServerFunc:              funcDoGetFailingHTTPServer,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			serverWg.Add(1)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			Convey("Then the error is returned in the error channel", func() {
				sErr := <-svcErrors
				So(sErr.Error(), ShouldResemble, fmt.Sprintf("failure in http listen and serve: %s", errServer.Error()))
				So(len(failingServerMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})
	})
}

func TestClose(t *testing.T) {
	Convey("Having a correctly initialised service", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		hcStopped := false
		serverStopped := false

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(string, healthcheck.Checker) error { return nil },
			StartFunc:    func(context.Context) {},
			StopFunc:     func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error { return nil },
			ShutdownFunc: func(context.Context) error {
				if !hcStopped {
					return errors.New("Server was stopped before healthcheck")
				}
				serverStopped = true
				return nil
			},
		}

		funcClose := func(context.Context) error {
			if !hcStopped {
				return errors.New("Dependency was closed before healthcheck")
			}
			if !serverStopped {
				return errors.New("Dependency was closed before http server")
			}
			return nil
		}

		// mongoDB will fail if healthcheck or http server are not stopped
		mongoMock := &storeMock.MongoDBMock{
			CloseFunc: funcClose,
		}

		Convey("Closing a service does not close uninitialised dependencies", func() {
			svcList := service.NewServiceList(nil)
			svcList.HealthCheck = true
			svc := service.New(cfg, svcList)
			svc.SetServer(serverMock)
			svc.SetHealthCheck(hcMock)
			err = svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
		})

		fullSvcList := &service.ExternalServiceList{
			HealthCheck: true,
			MongoDB:     true,
			Init:        nil,
		}

		Convey("Closing the service results in all the initialised dependencies being closed in the expected order", func() {
			svc := service.New(cfg, fullSvcList)
			svc.SetServer(serverMock)
			svc.SetHealthCheck(hcMock)
			svc.SetMongoDB(mongoMock)
			err = svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
			So(len(mongoMock.CloseCalls()), ShouldEqual, 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {
			failingserverMock := &serviceMock.HTTPServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(context.Context) error {
					return errors.New("Failed to stop http server")
				},
			}

			svc := service.New(cfg, fullSvcList)
			svc.SetServer(failingserverMock)
			svc.SetHealthCheck(hcMock)
			svc.SetMongoDB(mongoMock)
			err = svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "failed to shutdown gracefully")
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(failingserverMock.ShutdownCalls()), ShouldEqual, 1)
			So(len(mongoMock.CloseCalls()), ShouldEqual, 1)
		})
	})
}
