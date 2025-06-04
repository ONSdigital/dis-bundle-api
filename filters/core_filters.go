package filters

import (
	"fmt"
	"net/http"
	"time"
)

func parseQueryParam[T any](r *http.Request, paramName string, parser func(string) (*T, error)) (*T, error) {
	if !r.URL.Query().Has(paramName) {
		return nil, nil
	}

	value := r.URL.Query().Get(paramName)
	if value == "" {
		return nil, fmt.Errorf("malformed %s parameter: empty value", paramName)
	}

	parsed, err := parser(value)
	if err != nil {
		return nil, fmt.Errorf("malformed %s parameter. Value: %s, Error: %w", paramName, value, err)
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
