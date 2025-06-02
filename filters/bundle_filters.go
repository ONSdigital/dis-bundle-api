package filters

import (
	"net/http"
	"time"
)

const (
	PublishDate = "publish_date"
)

// Bundle filter option
type BundleFilters struct {
	PublishDate *time.Time
}

// Creates BundleFilters from the query parameters in the request
func CreateBundlefilters(r *http.Request) (*BundleFilters, *QueryParamParseError) {
	publishDate, err := parseQueryParam(r, PublishDate, parseTimeRFC3339)
	if err != nil {
		return nil, err
	}

	return &BundleFilters{
		PublishDate: publishDate,
	}, nil
}
