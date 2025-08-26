package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/application"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/pagination"
	"github.com/ONSdigital/dis-bundle-api/store"
	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	"github.com/gorilla/mux"
)

// API provides a struct to wrap the api around
type BundleAPI struct {
	Router                *mux.Router
	Store                 *store.Datastore
	stateMachineBundleAPI *application.StateMachineBundleAPI
	authMiddleware        auth.Middleware
	cli                   dphttp.Clienter
	config                *config.Config
}

// Setup function sets up the api and returns an api
func Setup(ctx context.Context, cfg *config.Config, router *mux.Router, dataStore *store.Datastore, stateMachineBundleAPI *application.StateMachineBundleAPI, authMiddleware auth.Middleware, cli dphttp.Clienter) *BundleAPI {
	api := &BundleAPI{
		Router:                router,
		Store:                 dataStore,
		stateMachineBundleAPI: stateMachineBundleAPI,
		authMiddleware:        authMiddleware,
		cli:                   cli,
		config:                cfg,
	}

	paginator := pagination.NewPaginator(cfg.DefaultLimit, cfg.DefaultOffset, cfg.DefaultMaxLimit)

	// get
	api.get(
		"/bundles",
		authMiddleware.Require("bundles:read", pagination.Paginate(paginator, api.getBundles)),
	)
	api.get(
		"/bundles/{bundle-id}",
		authMiddleware.Require("bundles:read", api.getBundle),
	)
	api.get(
		"/bundles/{bundle-id}/contents",
		authMiddleware.Require("bundles:read", paginator.Paginate(api.getBundleContents)),
	)
	api.get(
		"/bundle-events",
		authMiddleware.Require("bundles:read", paginator.Paginate(api.getBundleEvents)),
	)

	// post
	api.post(
		"/bundles",
		authMiddleware.Require("bundles:create", api.createBundle),
	)
	api.post(
		"/bundles/{bundle-id}/contents",
		authMiddleware.Require("bundles:create", api.postBundleContents),
	)

	// put
	api.put("/bundles/{bundle-id}",
		authMiddleware.Require("bundles:update", api.putBundle),
	)
	api.put("/bundles/{bundle-id}/state",
		authMiddleware.Require("bundles:update", api.putBundleState),
	)

	// delete
	api.delete(
		"/bundles/{bundle-id}",
		authMiddleware.Require("bundles:delete", api.deleteBundle),
	)
	api.delete(
		"/bundles/{bundle-id}/contents/{content-id}",
		authMiddleware.Require("bundles:delete", api.deleteContentItem),
	)

	return api
}

// get registers a GET http.HandlerFunc.
func (api *BundleAPI) get(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodGet)
}

// post registers a POST http.HandlerFunc.
func (api *BundleAPI) post(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodPost)
}

// put registers a PUT http.HandlerFunc.
func (api *BundleAPI) put(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodPut)
}

// delete registers a DELETE http.HandlerFunc.
func (api *BundleAPI) delete(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodDelete)
}
