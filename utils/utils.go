package utils

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/log.go/v2/log"
)

// HandleBundleAPIErr is a helper function to handle errors and set the HTTP response status code and headers accordingly
func HandleBundleAPIErr(w http.ResponseWriter, r *http.Request, httpStatusCode int, errors ...*models.Error) {
	var errList models.ErrorList

	for _, customError := range errors {
		if validationErr := models.ValidateError(customError); validationErr != nil {
			log.Error(r.Context(), "HandleBundleAPIErr: invalid error info provided", validationErr, log.Data{"invalid_error": customError})
			codeInternalServerError := models.CodeInternalServerError
			genericError := &models.Error{
				Code:        &codeInternalServerError,
				Description: "Failed to process the request due to an internal error",
			}
			httpStatusCode = http.StatusInternalServerError
			errList.Errors = append(errList.Errors, genericError)
		} else {
			errList.Errors = append(errList.Errors, customError)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)

	if err := json.NewEncoder(w).Encode(errList); err != nil {
		log.Error(r.Context(), "HandleBundleAPIErr: failed to encode error info", err)
	}
}

func HandleBundleAPIErrors(w http.ResponseWriter, r *http.Request, errInfos models.ErrorList, httpStatusCode int) {
	for _, errInfo := range errInfos.Errors {
		validationErr := models.ValidateError(errInfo)
		if validationErr == nil {
			continue
		}
		log.Error(r.Context(), "HandleBundleAPIErrors: invalid error info provided", validationErr)
		codeInternalServerError := models.CodeInternalServerError
		errInfo.Code = &codeInternalServerError
		errInfo.Description = "Failed to process the request due to an internal error"
		httpStatusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)

	if err := json.NewEncoder(w).Encode(errInfos); err != nil {
		log.Error(r.Context(), "HandleBundleAPIErrors: failed to encode error infos", err)
	}
}
