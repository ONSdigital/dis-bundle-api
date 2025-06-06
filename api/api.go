package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/application"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/pagination"
	"github.com/ONSdigital/dis-bundle-api/store"
	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/gorilla/mux"
)

// API provides a struct to wrap the api around
type BundleAPI struct {
	Router                *mux.Router
	Store                 *store.Datastore
	stateMachineBundleAPI *application.StateMachineBundleAPI
	authMiddleware        auth.Middleware
}

// Setup function sets up the api and returns an api
func Setup(ctx context.Context, cfg *config.Config, router *mux.Router, store *store.Datastore, stateMachineBundleAPI *application.StateMachineBundleAPI, authMiddleware auth.Middleware) *BundleAPI {
	api := &BundleAPI{
		Router:                router,
		Store:                 store,
		stateMachineBundleAPI: stateMachineBundleAPI,
		authMiddleware:        authMiddleware,
	}

	paginator := pagination.NewPaginator(cfg.DefaultLimit, cfg.DefaultOffset, cfg.DefaultMaxLimit)

	api.get(
		"/bundles",
		authMiddleware.Require("bundles:read", paginator.Paginate(api.getBundles)),
	)

	api.get(
		"/bundles/{bundle_id}",
		authMiddleware.Require("bundles:read", api.getBundleById),
	)

	return api
}

// get registers a GET http.HandlerFunc.
func (api *BundleAPI) get(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodGet)
}

func setMainHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
}
