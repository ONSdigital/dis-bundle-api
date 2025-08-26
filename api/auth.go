package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ONSdigital/dis-bundle-api/config"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	clientsidentity "github.com/ONSdigital/dp-api-clients-go/v2/identity"
	"github.com/ONSdigital/dp-net/v3/request"
	"github.com/ONSdigital/dp-permissions-api/sdk"
)

func (api *BundleAPI) GetAuthEntityData(r *http.Request) (*models.AuthEntityData, error) {

	fmt.Println("GOT INTO GET AUTH ENTITY")
	cfg, _ := config.Get()

	idClient := clientsidentity.NewWithHealthClient(&health.Client{
		Client: api.cli,
		URL:    cfg.ZebedeeURL,
		Name:   "identity",
	})

	fmt.Println("THE ID CLIENT IS")
	fmt.Println(idClient)
	bearerToken := strings.TrimPrefix(r.Header.Get(request.AuthHeaderKey), request.BearerPrefix)
	JWTEntityData, err := api.authMiddleware.Parse(bearerToken)
	if err != nil {
		// check service id token is valid
		resp, err := idClient.CheckTokenIdentity(r.Context(), bearerToken, clientsidentity.TokenTypeService)
		if err != nil {
			return nil, err
		} else {
			// valid
			JWTEntityData = &sdk.EntityData{UserID: resp.Identifier}
		}
	}

	return models.CreateAuthEntityData(JWTEntityData, bearerToken), nil
}
