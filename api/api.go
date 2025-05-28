package api

import (
	"context"

	"github.com/ONSdigital/dis-bundle-api/application"
	"github.com/ONSdigital/dis-bundle-api/store"
	"github.com/gorilla/mux"
)

// API provides a struct to wrap the API around
type API struct {
	Router                *mux.Router
	stateMachineBundleAPI *application.StateMachineBundleAPI
}

// Setup function sets up the API and returns an API
func Setup(ctx context.Context, r *mux.Router, store *store.Datastore, stateMachineBundleAPI *application.StateMachineBundleAPI) *API {
	api := &API{
		Router:                r,
		stateMachineBundleAPI: stateMachineBundleAPI,
	}

	// TODO: remove hello world example handler route
	r.HandleFunc("/hello", HelloHandler(ctx)).Methods("GET")
	return api
}
