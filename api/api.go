package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/application"
	"github.com/ONSdigital/dis-bundle-api/auth"
	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/pagination"
	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/ONSdigital/dis-bundle-api/utils"
	datasetAPISDK "github.com/ONSdigital/dp-dataset-api/sdk"
	"github.com/gorilla/mux"
)

const (
	AuthRoleBundlesRead   = "bundles:read"
	AuthRoleBundlesCreate = "bundles:create"
	AuthRoleBundlesUpdate = "bundles:update"
)

// API provides a struct to wrap the api around
type BundleAPI struct {
	Router                *mux.Router
	Store                 *store.Datastore
	stateMachineBundleAPI *application.StateMachineBundleAPI
	datasetAPIClient      datasetAPISDK.Clienter
	authMiddleware        auth.AuthorisationMiddleware
}

// Setup function sets up the api and returns an api
func Setup(ctx context.Context, cfg *config.Config, router *mux.Router, store *store.Datastore, stateMachineBundleAPI *application.StateMachineBundleAPI, datasetAPIClient datasetAPISDK.Clienter, authMiddleware auth.AuthorisationMiddleware) *BundleAPI {
	api := &BundleAPI{
		Router:                router,
		Store:                 store,
		stateMachineBundleAPI: stateMachineBundleAPI,
		datasetAPIClient:      datasetAPIClient,
		authMiddleware:        authMiddleware,
	}

	paginator := pagination.NewPaginator(cfg.DefaultLimit, cfg.DefaultOffset, cfg.DefaultMaxLimit)

	api.get(
		"/bundles",
		authMiddleware.Require(AuthRoleBundlesRead, pagination.Paginate(paginator, api.getBundles)),
	)
	api.get(
		"/bundles/{bundle-id}",
		authMiddleware.Require(AuthRoleBundlesRead, api.getBundle),
	)
	api.get(
		"/bundle-events",
		authMiddleware.Require(AuthRoleBundlesRead, paginator.Paginate(api.getBundleEvents)),
	)
	api.post(
		"/bundles/{bundle-id}/contents",
		authMiddleware.Require(AuthRoleBundlesCreate, api.postBundleContents),
	)
	api.put(
		"/bundles/{bundle-id}/state",
		authMiddleware.Require(AuthRoleBundlesUpdate, wrapHandler(api.putBundleState)),
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

func (api *BundleAPI) put(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodPut)
}

type Handler = func(w http.ResponseWriter, r *http.Request) (errBundles *models.Error)

func wrapHandler(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)

		if err != nil {
			utils.HandleBundleAPIErr(w, r, err.Code.HTTPStatusCode(), err)
			return
		}
	}
}
