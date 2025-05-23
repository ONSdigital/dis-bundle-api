package api

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/application"
	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
	"github.com/ONSdigital/log.go/v2/log"
)

func (api *BundleAPI) getBundles(w http.ResponseWriter, r *http.Request, limit, offset int) (interface{}, int, error) {
	ctx := r.Context()

	bundles, totalCount, err := application.BundleStore.ListBundles(api.bundleAPI.Store, ctx, offset, limit)
	if err != nil {
		log.Error(ctx, "failed to get bundles", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return nil, 0, err
	}

	bodyBytes, err := json.Marshal(bundles)
	if err != nil {
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return nil, 0, err
	}

	etag := dpresponse.GenerateETag(bodyBytes, false)
	dpresponse.SetETag(w, etag)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	w.WriteHeader(http.StatusOK)
	w.Write(bodyBytes)

	return bundles, totalCount, nil
}
