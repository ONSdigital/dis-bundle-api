package filters

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	StringParamName   = "validString"
	StringParamValue  = "some_string_should_be_parsed"
	StringParamResult = "this_should_be_the_output"

	DateParamValue = "2025-04-23T09:00:00Z"
)

func TestParseQueryParam(t *testing.T) {
	t.Parallel()
	Convey("When we call ParseQueryParam", t, func() {
		Convey("Then it should parse a valid string", func() {
			queryParams := url.Values{StringParamName: []string{StringParamValue}}
			req := &http.Request{
				URL: &url.URL{RawQuery: queryParams.Encode()},
			}

			parser := func(input string) (*string, error) {
				return &StringParamResult, nil
			}

			result, err := parseQueryParam(req, StringParamName, parser)
			So(err, ShouldBeNil)

			So(result, ShouldResemble, &StringParamResult)
		})

		Convey("Then it should return nil for a missing param", func() {
			paramValue := "some_string_should_be_parsed"
			queryParams := url.Values{StringParamName: []string{paramValue}}
			req := &http.Request{
				URL: &url.URL{RawQuery: queryParams.Encode()},
			}

			parser := func(input string) (*string, error) {
				return &StringParamResult, nil
			}

			result, err := parseQueryParam(req, "this_param_was_not_supplied", parser)
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Then it should return error for empty param", func() {
			paramName := "emptyString"
			paramValue := ""
			queryParams := url.Values{paramName: []string{paramValue}}

			req := &http.Request{
				URL: &url.URL{RawQuery: queryParams.Encode()},
			}

			parser := func(input string) (*string, error) {
				return nil, nil
			}

			result, err := parseQueryParam(req, paramName, parser)
			So(err.Error, ShouldEqual, fmt.Errorf("malformed %s parameter: empty value", paramName))
			So(result, ShouldBeNil)
		})

		Convey("Then it should return error if parser returns error", func() {
			paramName := "invalidString"
			paramValue := "this_should_error"
			queryParams := url.Values{paramName: []string{paramValue}}

			req := &http.Request{
				URL: &url.URL{RawQuery: queryParams.Encode()},
			}

			parserError := errors.New("some error here")

			parser := func(input string) (*string, error) {
				return nil, parserError
			}

			expectedError := fmt.Errorf("malformed %s parameter. Value: %s, Error: %w", paramName, paramValue, parserError)

			result, err := parseQueryParam(req, paramName, parser)
			So(err, ShouldNotBeNil)
			So(err.Error, ShouldEqual, expectedError)
			So(err.Source.Parameter, ShouldEqual, paramName)
			So(result, ShouldBeNil)
		})

		Convey("Then it should work with return parsed date time", func() {
			paramName := "publish_date"
			paramValue := DateParamValue
			expectedDateTime := time.Date(2025, 04, 23, 9, 0, 0, 0, time.UTC)

			queryParams := url.Values{paramName: []string{paramValue}}

			req := &http.Request{
				URL: &url.URL{RawQuery: queryParams.Encode()},
			}

			result, err := parseQueryParam(req, paramName, parseTimeRFC3339)
			So(err, ShouldBeNil)
			So(result.UnixNano(), ShouldAlmostEqual, expectedDateTime.UnixNano())
		})
	})
}

func TestParseTimeRFC3339(t *testing.T) {
	t.Parallel()

	Convey("When we call parseTimeRFC3339", t, func() {
		Convey("Then it should parse a valid RFC3999 time string", func() {
			value := DateParamValue
			expectedDateTime := time.Date(2025, 04, 23, 9, 0, 0, 0, time.UTC)

			result, err := parseTimeRFC3339(value)

			So(err, ShouldBeNil)
			So(result.UnixNano(), ShouldAlmostEqual, expectedDateTime.UnixNano())
		})

		Convey("Then it error for a string in another format", func() {
			value := "2025-04-23"

			result, err := parseTimeRFC3339(value)
			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}
