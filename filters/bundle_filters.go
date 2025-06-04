package filters

import (
	"net/http"
	"time"
)

const (
	PublishDate = "publish_date"
)

// Bundle filter optio
type Bundlefilters struct {
	PublishDate *time.Time
}

func CreateBundlefilters(r *http.Request) (*Bundlefilters, error) {
	publishDate, err := parseQueryParam(r, PublishDate, parseTimeRFC3339)
	if err != nil {
		return nil, err
	}

	return &Bundlefilters{
		PublishDate: publishDate,
	}, nil
}
