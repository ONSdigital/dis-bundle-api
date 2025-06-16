package filters

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBundleFilters(t *testing.T) {
	t.Parallel()

	Convey("When we call CreateBundleFilters", t, func() {
		Convey("Then it creates valid bundle filters from the request", func() {
			paramValue := "2025-01-01T01:00:00Z"
			queryParams := url.Values{PublishDate: []string{paramValue}}
			req := &http.Request{
				URL: &url.URL{RawQuery: queryParams.Encode()},
			}

			publishDate := time.Date(2025, 1, 1, 01, 0, 0, 0, time.UTC)

			expectedResult := BundleFilters{
				PublishDate: &publishDate,
			}
			result, err := CreateBundlefilters(req)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, &expectedResult)
		})

		Convey("Then it returns error if parsing error", func() {
			paramValue := "not-an-actual-date"
			queryParams := url.Values{PublishDate: []string{paramValue}}
			req := &http.Request{
				URL: &url.URL{RawQuery: queryParams.Encode()},
			}

			result, err := CreateBundlefilters(req)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}
