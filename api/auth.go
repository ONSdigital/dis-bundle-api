package api

import (
	"net/http"
	"strings"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dp-net/v3/request"
	"github.com/ONSdigital/dp-permissions-api/sdk"
)

func (api *BundleAPI) GetAuthEntityData(r *http.Request) (*models.AuthEntityData, error) {
	bearerTokenValue := getBearerTokenValue(r)
	var JWTEntityData *sdk.EntityData
	if bearerTokenValue != nil {
		var err error
		JWTEntityData, err = api.authMiddleware.Parse(*bearerTokenValue)
		if err != nil {
			return nil, err
		}
	}

	return models.CreateAuthEntityData(JWTEntityData, *bearerTokenValue), nil
}

func getBearerTokenValue(r *http.Request) *string {
	authHeader := r.Header.Get(request.AuthHeaderKey)
	if authHeader == "" {
		return &authHeader
	}

	if !strings.HasPrefix(authHeader, request.BearerPrefix) {
		return nil
	}

	authHeaderValue := r.Header.Get(request.AuthHeaderKey)
	trimmed := strings.TrimPrefix(authHeaderValue, request.BearerPrefix)
	return &trimmed
}

func (api *BundleAPI) getAuthData(r *http.Request) (*sdk.EntityData, error) {
	authToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	return api.authMiddleware.Parse(authToken)
}
