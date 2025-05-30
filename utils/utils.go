package utils

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/log.go/v2/log"
)

// HandleBundleAPIErr is a helper function to handle errors and set the HTTP response status code and headers accordingly
func HandleBundleAPIErr(w http.ResponseWriter, r *http.Request, errInfo *models.Error, httpStatusCode int) {
	if validationErr := models.ValidateError(errInfo); validationErr != nil {
		log.Error(r.Context(), "HandleBundleAPIErr: invalid error info provided", validationErr)
		codeInternalServerError := models.CodeInternalServerError
		errInfo = &models.Error{
			Code:        &codeInternalServerError,
			Description: "Failed to process the request due to an internal error",
		}
		httpStatusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)

	if err := json.NewEncoder(w).Encode(errInfo); err != nil {
		log.Error(r.Context(), "HandleBundleAPIErr: failed to encode error info", err)
	}
}
