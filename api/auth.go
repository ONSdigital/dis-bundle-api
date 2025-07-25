package api

import (
	"net/http"
	"strings"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dp-net/v3/request"
)

func (api *BundleAPI) GetAuthEntityData(r *http.Request) (*models.AuthEntityData, error) {
	bearerToken := strings.TrimPrefix(r.Header.Get(request.AuthHeaderKey), request.BearerPrefix)

	JWTEntityData, err := api.authMiddleware.Parse(bearerToken)
	if err != nil {
		return nil, err
	}

	return models.CreateAuthEntityData(JWTEntityData, bearerToken), nil
}
