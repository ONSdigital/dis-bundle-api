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
	return api
}

// func writeErrorResponse(ctx context.Context, w http.ResponseWriter, errorResponse *models.Error, statusCode int) {
// 	var jsonResponse []byte
// 	var err error
// 	w.Header().Set("Content-Type", "application/json")
// 	// process custom headers
// 	// if errorResponse.Headers != nil {
// 	// 	for key := range errorResponse.Headers {
// 	// 		w.Header().Set(key, errorResponse.Headers[key])
// 	// 	}
// 	// }
// 	// w.WriteHeader(errorResponse.Status)
// 	if statusCode == http.StatusInternalServerError {
// 		code := models.CodeInternalServerError
// 		jsonResponse, err = json.Marshal(models.Error{Code: &code, Description: "internal server error"})
// 	} else {
// 		jsonResponse, err = json.Marshal(errorResponse)
// 	}
// 	if err != nil {
// 		code := models.JSONMarshalError
// 		responseErr := models.Error{Code: &code, Description: "json marshal error"}
// 		http.Error(w, responseErr.Description, http.StatusInternalServerError)
// 		return
// 	}

// 	_, err = w.Write(jsonResponse)
// 	if err != nil {
// 		code := models.WriteResponseError
// 		responseErr := models.Error{Code: &code, Description: "write response error"}
// 		http.Error(w, responseErr.Description, http.StatusInternalServerError)
// 		return
// 	}
// }

// func writeSuccessResponse(w http.ResponseWriter, successResponse *models.SuccessResponse) {
// 	w.Header().Set("Content-Type", "application/json")
// 	// process custom headers
// 	if successResponse.Headers != nil {
// 		for key := range successResponse.Headers {
// 			w.Header().Set(key, successResponse.Headers[key])
// 		}
// 	}
// 	w.WriteHeader(successResponse.Status)

// 	_, err := w.Write(successResponse.Body)
// 	if err != nil {
// 		code := models.WriteResponseError
// 		responseErr := models.Error{Code: &code, Description: "write response error"}
// 		http.Error(w, responseErr.Description, http.StatusInternalServerError)
// 		return
// 	}
// }

// type baseHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request) (*models.SuccessResponse, *models.Error)

// func contextAndErrors(h baseHandler) http.HandlerFunc {
// 	return func(w http.ResponseWriter, req *http.Request) {
// 		ctx := req.Context()
// 		response, err := h(ctx, w, req)
// 		if err != nil {
// 			writeErrorResponse(ctx, w, err, 500)
// 			return
// 		}
// 		writeSuccessResponse(w, response)
// 	}
// }

// get registers a GET http.HandlerFunc.
func (api *BundleAPI) get(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodGet)
}

// type AuthHandler interface {
// 	Require(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc
// }

// func (api *BundleAPI) isAuthorised(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc {
// 	return api.permissions.Require(required, handler)
// }
