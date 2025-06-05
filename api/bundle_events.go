package api

import (
	"net/http"
	"time"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/dis-bundle-api/utils"
	"github.com/ONSdigital/log.go/v2/log"
)

func (api *BundleAPI) getBundleEvents(w http.ResponseWriter, r *http.Request, limit, offset int) (events any, totalCount int, eventErrors *models.Error) {
	ctx := r.Context()

	allowedParams := map[string]bool{
		"bundle": true,
		"after":  true,
		"before": true,
		"limit":  true,
		"offset": true,
	}

	for param := range r.URL.Query() {
		if !allowedParams[param] {
			code := models.ErrInvalidParameters
			errInfo := &models.Error{
				Code:        &code,
				Description: "Unable to process request due to a malformed or invalid request body or query parameter",
				Source:      &models.Source{Parameter: param},
			}
			utils.HandleBundleAPIErr(w, r, errInfo, http.StatusBadRequest)
			return []*models.Event{}, 0, nil
		}
	}

	bundleID := r.URL.Query().Get("bundle")
	afterParam := r.URL.Query().Get("after")
	beforeParam := r.URL.Query().Get("before")

	var after, before *time.Time

	if afterParam != "" {
		afterTime, err := time.Parse(time.RFC3339, afterParam)
		if err != nil {
			code := models.ErrInvalidParameters
			errInfo := &models.Error{
				Code:        &code,
				Description: "Unable to process request due to a malformed or invalid request body or query parameter",
				Source:      &models.Source{Parameter: "after"},
			}
			utils.HandleBundleAPIErr(w, r, errInfo, http.StatusBadRequest)
			return []*models.Event{}, 0, nil
		}
		after = &afterTime
	}

	if beforeParam != "" {
		beforeTime, err := time.Parse(time.RFC3339, beforeParam)
		if err != nil {
			code := models.ErrInvalidParameters
			errInfo := &models.Error{
				Code:        &code,
				Description: "Unable to process request due to a malformed or invalid request body or query parameter",
				Source:      &models.Source{Parameter: "before"},
			}
			utils.HandleBundleAPIErr(w, r, errInfo, http.StatusBadRequest)
			return []*models.Event{}, 0, nil
		}
		before = &beforeTime
	}

	events, totalCount, err := api.stateMachineBundleAPI.ListBundleEvents(ctx, offset, limit, bundleID, after, before)
	if err != nil {
		code := models.InternalError
		log.Error(ctx, "failed to get bundle events", err)
		errInfo := &models.Error{Code: &code, Description: "Failed to process the request due to an internal error"}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusInternalServerError)
		return nil, 0, nil
	}

	if totalCount == 0 && bundleID == "" && after == nil && before == nil {
		code := models.NotFound
		errInfo := &models.Error{Code: &code, Description: "The requested resource does not exist."}
		utils.HandleBundleAPIErr(w, r, errInfo, http.StatusNotFound)
		return []*models.Event{}, 0, nil
	}

	return events, totalCount, nil
}
