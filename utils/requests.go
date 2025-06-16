package utils

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/gorilla/mux"
)

// GetBundleID gets the Bundle ID from route vars from the provided request
func GetBundleID(r *http.Request) (string, *models.Error) {
	return GetRouteVariableWithDefaultErrors(r, models.RouteBundleID)
}

// GetRouteVariableWithDefaultErrors gets the specified route variable from the provided request, and returns default error messages constructed using the variable's name
func GetRouteVariableWithDefaultErrors(r *http.Request, variable string) (string, *models.Error) {
	return GetRouteVariable(r, variable, apierrors.CreateMissingRouteVariableError(variable), apierrors.CreateMissingRouteVariableError(variable))
}

// GetRouteVariable gets the specified route variable from the provided request, and returns the provided errors if the variable is missing or empty.
func GetRouteVariable(r *http.Request, variable, missingError, emptyError string) (string, *models.Error) {
	vars := mux.Vars(r)
	if len(vars) == 0 {
		return "", models.CreateModelError(models.CodeBadRequest, apierrors.ErrorDescriptionMissingParameters)
	}

	variable, ok := vars[variable]
	if !ok {
		return "", models.CreateModelError(models.CodeBadRequest, missingError)
	}

	if variable == "" {
		return "", models.CreateModelError(models.CodeBadRequest, emptyError)
	}

	return variable, nil
}

// GetETag gets the ETag value from the If-Match header in the provided http.Request object
func GetETag(r *http.Request) (*string, *models.Error) {
	etag := r.Header.Get(HeaderIfMatch)
	if etag == "" {
		return nil, models.CreateModelError(models.CodeBadRequest, apierrors.ErrorDescriptionETagMissing)
	}

	return &etag, nil
}

// GetJSONRequestBody reads the body from the http.Request and tries to unmarshal it as a JSON to the specified type T
func GetJSONRequestBody[T any](r *http.Request) (*T, *models.Error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, models.CreateModelError(models.CodeBadRequest, apierrors.ErrorDescriptionMalformedRequest)
	}
	var requestBody T
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		return nil, models.CreateModelError(models.CodeBadRequest, apierrors.ErrorDescriptionMalformedRequest)
	}

	return &requestBody, nil
}
