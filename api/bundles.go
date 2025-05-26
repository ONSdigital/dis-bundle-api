package api

import (
	"fmt"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/log.go/v2/log"
)

func (api *BundleAPI) getBundles(w http.ResponseWriter, r *http.Request, limit, offset int) (bundles any, errCode int, errBundles *models.Error) {
	ctx := r.Context()

	bundles, totalCount, err := api.stateMachineBundleAPI.ListBundles(ctx, offset, limit)
	if err != nil {
		fmt.Println("AN ERROR WAS RETURNED")
		fmt.Println(err)
		code := models.CodeInternalServerError
		log.Error(ctx, "failed to get bundles", err)
		return nil, http.StatusInternalServerError, &models.Error{Code: &code, Description: "failed to get bundle"}
	}

	return bundles, totalCount, nil
}
