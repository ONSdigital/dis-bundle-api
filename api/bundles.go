package api

import (
	"encoding/json"
	"net/http"

	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
	"github.com/ONSdigital/log.go/v2/log"
)

func (api *BundleAPI) getBundles(w http.ResponseWriter, r *http.Request, limit, offset int) (bundles any, errCode int, err error) {
	ctx := r.Context()

	bundles, totalCount, err := api.stateMachineBundleAPI.ListBundles(ctx, offset, limit)

	if err != nil {
		log.Error(ctx, "failed to get bundles", err)
		return nil, 0, err
	}

	bodyBytes, err := json.Marshal(bundles)
	if err != nil {
		log.Error(ctx, "failed writing bytes to response", err)
		return nil, 0, err
	}

	etag := dpresponse.GenerateETag(bodyBytes, false)
	dpresponse.SetETag(w, etag)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bodyBytes)
	if err != nil {
		log.Error(ctx, "failed writing bytes to response", err)
	}

	return bundles, totalCount, nil
}
