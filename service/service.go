package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/api"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the API
type Service struct {
	Config      *config.Config
	Server      HTTPServer
	Router      *mux.Router
	API         *api.API
	ServiceList *ExternalServiceList
	HealthCheck HealthChecker
	mongoDB     store.MongoDB
}

type BundleAPIStore struct {
	store.MongoDB
}

// New creates a new service
func New(cfg *config.Config, serviceList *ExternalServiceList) *Service {
	svc := &Service{
		Config:      cfg,
		ServiceList: serviceList,
	}
	return svc
}

// SetServer sets the http server for a service
func (svc *Service) SetServer(server HTTPServer) {
	svc.Server = server
}

// SetHealthCheck sets the healthchecker for a service
func (svc *Service) SetHealthCheck(healthCheck HealthChecker) {
	svc.HealthCheck = healthCheck
}

// SetMongoDB sets the mongoDB connection for a service
func (svc *Service) SetMongoDB(mongoDB store.MongoDB) {
	svc.mongoDB = mongoDB
}

// Run the service
func (svc *Service) Run(ctx context.Context, buildTime, gitCommit, version string, svcErrors chan error) (err error) {
	log.Info(ctx, "running service")
	cfg := svc.Config
	log.Info(ctx, "using service configuration", log.Data{"config": cfg})

	// Get MongoDB client
	svc.mongoDB, err = svc.ServiceList.GetMongoDB(ctx, cfg.MongoConfig)
	if err != nil {
		log.Fatal(ctx, "failed to initialise mongo DB", err)
		return err
	}

	// Get HealthCheck
	svc.HealthCheck, err = svc.ServiceList.GetHealthCheck(svc.Config, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return err
	}

	if err := svc.registerCheckers(ctx); err != nil {
		return errors.Wrap(err, "unable to register checkers")
	}
	// Get HTTP Server and ... // TODO: Add any middleware that your service requires
	r := mux.NewRouter()
	middle := svc.createMiddleware(svc.Config)
	svc.Server = svc.ServiceList.GetHTTPServer(svc.Config.BindAddr, middle.Then(r))

	// Set up the API
	s := store.DataStore{Backend: BundleAPIStore{svc.mongoDB}}
	svc.API = api.Setup(ctx, r, &s)

	svc.HealthCheck.Start(ctx)

	// Run the http server in a new go-routine
	go func() {
		if err := svc.Server.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()
	return nil
}

// CreateMiddleware creates an Alice middleware chain of handlers
// to forward collectionID from cookie from header
func (svc *Service) createMiddleware(cfg *config.Config) alice.Chain {
	// healthcheck
	healthcheckHandler := healthcheckMiddleware(svc.HealthCheck.Handler, "/health")
	middleware := alice.New(healthcheckHandler)

	return middleware
}

// healthcheckMiddleware creates a new http.Handler to intercept /health requests.
func healthcheckMiddleware(healthcheckHandler func(http.ResponseWriter, *http.Request), path string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == "GET" && req.URL.Path == path {
				healthcheckHandler(w, req)
				return
			}

			h.ServeHTTP(w, req)
		})
	}
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Config.GracefulShutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	shutdownContext, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	// Gracefully shutdown the application closing any open resources.
	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.ServiceList.HealthCheck {
			svc.HealthCheck.Stop()
		}

		// stop any incoming requests
		if err := svc.Server.Shutdown(shutdownContext); err != nil {
			log.Error(shutdownContext, "failed to shutdown http server", err)
			hasShutdownError = true
		}

		// Close MongoDB (if it exists)
		if svc.ServiceList.MongoDB {
			if err := svc.mongoDB.Close(shutdownContext); err != nil {
				log.Error(shutdownContext, "failed to close mongo db session", err)
				hasShutdownError = true
			}
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-shutdownContext.Done()

	// timeout expired
	if shutdownContext.Err() == context.DeadlineExceeded {
		log.Error(shutdownContext, "shutdown timed out", shutdownContext.Err())
		return shutdownContext.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Error(shutdownContext, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(shutdownContext, "graceful shutdown was successful")
	return nil
}

func (svc *Service) registerCheckers(ctx context.Context) (err error) {
	hasErrors := false
	// TODO: add other health checks here, as per dp-upload-service
	if err = svc.HealthCheck.AddCheck("Mongo DB", svc.mongoDB.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for mongo db", err)
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
