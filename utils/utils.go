package utils

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	HeaderIfMatch = "If-Match"
)

// HandleBundleAPIErr is a helper function to handle errors and set the HTTP response status code and headers accordingly
func HandleBundleAPIErr(w http.ResponseWriter, r *http.Request, httpStatusCode int, errors ...*models.Error) {
	var errList models.ErrorList

	for _, customError := range errors {
		if validationErr := models.ValidateError(customError); validationErr != nil {
			log.Error(r.Context(), "HandleBundleAPIErr: invalid error info provided", validationErr, log.Data{"invalid_error": customError})
			codeInternalError := models.CodeInternalError
			genericError := &models.Error{
				Code:        &codeInternalError,
				Description: apierrors.ErrorDescriptionInternalError,
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

// GetETag reads the If-Match header from the request, and returns an error if it doesn't exist, otherwise it will return the header value
func GetETag(r *http.Request) (*string, error) {
	etag := r.Header.Get(HeaderIfMatch)
	if etag == "" {
		return nil, apierrors.ErrMissingIfMatchHeader
	}

	return &etag, nil
}

// GetRequestBody attempts to read the request body as JSON and unmarshal it into the specified type T
func GetRequestBody[T any](r *http.Request) (*T, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, apierrors.ErrInvalidBody
	}

	var requestBody T
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		return nil, apierrors.ErrInvalidBody
	}

	return &requestBody, nil
}

func MapErrorCodeToStatus(code *models.Code) int {
	if code == nil {
		return http.StatusInternalServerError
	}

	switch *code {
	case models.CodeNotFound:
		return http.StatusNotFound
	case models.CodeBadRequest, models.CodeInvalidParameters, models.CodeMissingParameters:
		return http.StatusBadRequest
	case models.CodeUnauthorised:
		return http.StatusUnauthorized
	case models.CodeForbidden:
		return http.StatusForbidden
	case models.CodeConflict:
		return http.StatusConflict
	case models.CodeInternalError, models.CodeJSONMarshalError, models.CodeJSONUnmarshalError, models.CodeWriteResponseError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
