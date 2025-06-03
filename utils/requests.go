package utils

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/gorilla/mux"
)

// Get Bundle ID from route (assuming route slug is named bundle_id)
func GetBundleID(r *http.Request) (string, *models.Error) {
	return GetRouteVariableWithDefaultErrors(r, models.RouteBundleId)
}

func GetRouteVariableWithDefaultErrors(r *http.Request, variable string) (string, *models.Error) {
	return GetRouteVariable(r, variable, apierrors.CreateMissingRouteVariableError(variable), apierrors.CreateMissingRouteVariableError(variable))
}

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

func GetETag(r *http.Request) (*string, *models.Error) {
	etag := r.Header.Get(HeaderIfMatch)
	if etag == "" {
		return nil, models.CreateModelError(models.CodeBadRequest, apierrors.ErrorDescriptionETagMissing)
	}

	return &etag, nil
}

func GetRequestBody[T any](r *http.Request) (*T, *models.Error) {
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
