package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/application"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/pagination"
	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/gorilla/mux"
)

var (
	readPermission = auth.Permissions{Read: true}
)

// API provides a struct to wrap the api around
type BundleAPI struct {
	Router                *mux.Router
	Store                 *store.Datastore
	bundleAPI             *application.BundleService
	stateMachineBundleAPI *application.StateMachineBundleAPI
	permissions           AuthHandler
}

// Setup function sets up the api and returns an api
func Setup(ctx context.Context, cfg *config.Config, router *mux.Router, store *store.Datastore, stateMachineBundleAPI *application.StateMachineBundleAPI, permissions AuthHandler) *BundleAPI {
	api := &BundleAPI{
		Router:                router,
		Store:                 store,
		bundleAPI:             application.NewBundleService(store),
		stateMachineBundleAPI: stateMachineBundleAPI,
		permissions:           permissions,
	}

	paginator := pagination.NewPaginator(cfg.DefaultLimit, cfg.DefaultOffset, cfg.DefaultMaxLimit)
	// TODO: remove hello world example handler route
	api.get(
		"/bundles",
		api.isAuthorised(readPermission, paginator.Paginate(api.getBundles)),
	)
	return api
}

// get registers a GET http.HandlerFunc.
func (api *BundleAPI) get(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodGet)
}

type AuthHandler interface {
	Require(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc
}

func (api *BundleAPI) isAuthorised(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc {
	return api.permissions.Require(required, handler)
}
