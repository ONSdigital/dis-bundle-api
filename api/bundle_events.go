package api

import (
	"net/http"
	"time"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
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

	var validationErrors []*models.Error

	for param := range r.URL.Query() {
		if !allowedParams[param] {
			code := models.CodeInvalidParameters
			errInfo := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionMalformedRequest,
				Source:      &models.Source{Parameter: param},
			}
			validationErrors = append(validationErrors, errInfo)
		}
	}

	bundleID := r.URL.Query().Get("bundle")
	afterParam := r.URL.Query().Get("after")
	beforeParam := r.URL.Query().Get("before")

	var after, before *time.Time

	if afterParam != "" {
		afterTime, err := time.Parse(time.RFC3339, afterParam)
		if err != nil {
			code := models.CodeInvalidParameters
			errInfo := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionMalformedRequest,
				Source:      &models.Source{Parameter: "after"},
			}
			validationErrors = append(validationErrors, errInfo)
		} else {
			after = &afterTime
		}
	}

	if beforeParam != "" {
		beforeTime, err := time.Parse(time.RFC3339, beforeParam)
		if err != nil {
			code := models.CodeInvalidParameters
			errInfo := &models.Error{
				Code:        &code,
				Description: errs.ErrorDescriptionMalformedRequest,
				Source:      &models.Source{Parameter: "before"},
			}
			validationErrors = append(validationErrors, errInfo)
		} else {
			before = &beforeTime
		}
	}

	if len(validationErrors) > 0 {
		utils.HandleBundleAPIErr(w, r, http.StatusBadRequest, validationErrors...)
		return nil, 0, validationErrors[0]
	}

	events, totalCount, err := api.stateMachineBundleAPI.ListBundleEvents(ctx, offset, limit, bundleID, after, before)
	if err != nil {
		code := models.CodeInternalError
		log.Error(ctx, "failed to get bundle events", err)
		errInfo := &models.Error{Code: &code, Description: "Failed to process the request due to an internal error"}
		utils.HandleBundleAPIErr(w, r, http.StatusInternalServerError, errInfo)
		return nil, 0, nil
	}

	if totalCount == 0 {
		code := models.CodeNotFound
		errInfo := &models.Error{Code: &code, Description: "The requested resource does not exist."}
		utils.HandleBundleAPIErr(w, r, http.StatusNotFound, errInfo)
		return nil, 0, errInfo
	}

	return events, totalCount, nil
}
