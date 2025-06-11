package filters

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ONSdigital/dis-bundle-api/models"
)

type QueryParamParseError struct {
	Error  error
	Source *models.Source
}

func CreateQueryParamParseError(err error, paramName string) *QueryParamParseError {
	return &QueryParamParseError{
		Error: err,
		Source: &models.Source{
			Parameter: paramName,
		},
	}
}

func parseQueryParam[T any](r *http.Request, paramName string, parser func(string) (*T, error)) (*T, *QueryParamParseError) {
	if !r.URL.Query().Has(paramName) {
		return nil, nil
	}

	value := r.URL.Query().Get(paramName)
	if value == "" {
		return nil, CreateQueryParamParseError(fmt.Errorf("malformed %s parameter: empty value", paramName), paramName)
	}

	parsed, err := parser(value)
	if err != nil {
		return nil, CreateQueryParamParseError(fmt.Errorf("malformed %s parameter. Value: %s, Error: %w", paramName, value, err), paramName)
	}

	return parsed, nil
}

func parseTimeRFC3339(value string) (*time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)

	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
