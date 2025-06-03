package utils

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateResponseBuilder(t *testing.T) {
	Convey("CreateResponseBuilder", t, func() {
		Convey("Should create a response builder", func() {
			responseBuilder := CreateHTTPResponseBuilder()

			So(responseBuilder, ShouldNotBeNil)
		})
	})

	Convey("CreateResponseBuilder", t, func() {
		Convey("Should create a response builder with default values", func() {
			httpResponseBuilder := CreateHTTPResponseBuilder()

			So(httpResponseBuilder, ShouldNotBeNil)

			responseBuilder := httpResponseBuilder.(*ResponseBuilder)
			So(responseBuilder.headers, ShouldNotBeNil)
			So(responseBuilder.headers, ShouldBeEmpty)

			So(responseBuilder.statusCode, ShouldEqual, 0)
			So(responseBuilder.body, ShouldBeNil)

			So(responseBuilder.etag, ShouldBeNil)
		})
	})

	Convey("WithHeader", t, func() {
		headerKey := "testing header"
		headerValue := "test header value"
		Convey("Should set header", func() {
			httpResponseBuilder := CreateHTTPResponseBuilder().WithHeader(headerKey, headerValue)

			responseBuilder := httpResponseBuilder.(*ResponseBuilder)

			So(responseBuilder.headers, ShouldNotBeEmpty)
			So(responseBuilder.headers, ShouldContainKey, headerKey)
			actualValue := responseBuilder.headers[headerKey]
			So(actualValue, ShouldEqual, headerValue)
		})
	})

	Convey("WithETag", t, func() {
		etagValue := "etag value"
		Convey("Should set etag", func() {
			httpResponseBuilder := CreateHTTPResponseBuilder().WithETag(etagValue)
			responseBuilder := httpResponseBuilder.(*ResponseBuilder)

			So(responseBuilder.etag, ShouldNotBeNil)

			So(*responseBuilder.etag, ShouldEqual, etagValue)
		})
	})

	Convey("WithCacheControl", t, func() {
		cacheControl := CacheControlNoStore

		Convey("Should set cache control header", func() {
			httpResponseBuilder := CreateHTTPResponseBuilder().WithCacheControl(cacheControl)
			responseBuilder := httpResponseBuilder.(*ResponseBuilder)

			So(responseBuilder.headers, ShouldNotBeEmpty)
			So(responseBuilder.headers, ShouldContainKey, HeaderCacheControl)
			actualValue := responseBuilder.headers[HeaderCacheControl]
			So(actualValue, ShouldEqual, cacheControl.String())
		})
	})

	Convey("WithStatusCode", t, func() {
		statusCode := 200

		Convey("Should set status code", func() {
			httpResponseBuilder := CreateHTTPResponseBuilder().WithStatusCode(statusCode)
			responseBuilder := httpResponseBuilder.(*ResponseBuilder)

			So(responseBuilder.statusCode, ShouldEqual, statusCode)
		})
	})

	Convey("WithBody", t, func() {
		body := models.Bundle{
			ID: "bundle-id",
		}

		responseType := ResponseBodyTypeJson

		Convey("Should set body", func() {
			httpResponseBuilder := CreateHTTPResponseBuilder().WithBody(ResponseBodyTypeJson, body)
			responseBuilder := httpResponseBuilder.(*ResponseBuilder)

			So(responseBuilder.body, ShouldNotBeNil)
			So(responseBuilder.body.responseType, ShouldEqual, responseType)
			So(responseBuilder.body.body, ShouldEqual, body)
		})
	})

	Convey("WithJsonBody", t, func() {
		body := models.Bundle{
			ID: "bundle-id",
		}
		responseType := ResponseBodyTypeJson

		Convey("Should set body", func() {
			httpResponseBuilder := CreateHTTPResponseBuilder().WithJsonBody(body)
			responseBuilder := httpResponseBuilder.(*ResponseBuilder)

			So(responseBuilder.body, ShouldNotBeNil)
			So(responseBuilder.body.responseType, ShouldEqual, responseType)
			So(responseBuilder.body.body, ShouldEqual, body)
		})
	})

	Convey("Build", t, func() {

		etagValue := "test-etag-value"
		cacheControl := CacheControlNoStore
		headers := map[string]string{
			"test-header-one":   "test-header-value",
			"other-test-header": "expected-test-header-value",
		}
		statusCode := 400
		responseBody := models.Bundle{
			ID:         "bundle-1234",
			BundleType: models.BundleTypeScheduled,
		}

		Convey("Should set response etag header", func() {
			rr := httptest.NewRecorder()
			CreateHTTPResponseBuilder().WithETag(etagValue).Build(rr)

			So(rr.Header().Get(HeaderETag), ShouldEqual, etagValue)
		})

		Convey("Should set response cache-control header", func() {
			rr := httptest.NewRecorder()
			CreateHTTPResponseBuilder().WithCacheControl(cacheControl).Build(rr)

			So(rr.Header().Get(HeaderCacheControl), ShouldEqual, cacheControl.String())
		})

		Convey("Should set response headers", func() {
			rr := httptest.NewRecorder()
			responseBuilder := CreateHTTPResponseBuilder()

			for key, value := range headers {
				responseBuilder = responseBuilder.WithHeader(key, value)
			}

			responseBuilder.Build(rr)

			for key, value := range headers {
				So(rr.Header().Get(key), ShouldEqual, value)
			}
		})

		Convey("Should set status code", func() {
			rr := httptest.NewRecorder()
			CreateHTTPResponseBuilder().WithStatusCode(statusCode).Build(rr)
			So(rr.Result().StatusCode, ShouldEqual, statusCode)
		})

		Convey("Should set response body + content-type header", func() {
			rr := httptest.NewRecorder()
			CreateHTTPResponseBuilder().WithJsonBody(responseBody).Build(rr)

			var responseBundle models.Bundle
			err := json.NewDecoder(rr.Body).Decode(&responseBundle)
			So(err, ShouldBeNil)
			So(responseBundle, ShouldEqual, responseBody)
		})

		Convey("Should set all", func() {
			rr := httptest.NewRecorder()
			builder := CreateHTTPResponseBuilder().
				WithCacheControl(cacheControl).
				WithStatusCode(statusCode).
				WithETag(etagValue).
				WithJsonBody(responseBody)

			for key, value := range headers {
				builder = builder.WithHeader(key, value)
			}

			builder.Build(rr)

			for key, value := range headers {
				So(rr.Header().Get(key), ShouldEqual, value)
			}

			var responseBundle models.Bundle
			err := json.NewDecoder(rr.Body).Decode(&responseBundle)
			So(err, ShouldBeNil)
			So(responseBundle, ShouldEqual, responseBody)

			So(rr.Result().StatusCode, ShouldEqual, statusCode)

			for key, value := range headers {
				So(rr.Header().Get(key), ShouldEqual, value)
			}

			So(rr.Header().Get(HeaderETag), ShouldEqual, etagValue)
			So(rr.Header().Get(HeaderCacheControl), ShouldEqual, cacheControl.String())
		})
	})
}

func TestResponseBody(t *testing.T) {
	testBundle := models.Bundle{
		BundleType: models.BundleTypeManual,
		ID:         "testing-bundle",
	}

	Convey("When type is json", t, func() {
		responseBody := createResponseBody(ResponseBodyTypeJson, testBundle)
		Convey("and the body can be marashalled", func() {
			Convey("Write should write the appropriate body + header", func() {
				rr := httptest.NewRecorder()
				responseBody.Write(rr)

				var responseBundle models.Bundle
				err := json.NewDecoder(rr.Body).Decode(&responseBundle)
				So(err, ShouldBeNil)
				So(responseBundle, ShouldEqual, responseBody.body)

				So(rr.Header().Get(HeaderContentType), ShouldEqual, string(ResponseBodyTypeJson))
			})
		})

		Convey("and the body cannot be marashalled", func() {
			responseBody := createResponseBody(ResponseBodyTypeJson, func() {})

			Convey("Write should return an error and not set values", func() {
				rr := httptest.NewRecorder()
				err := responseBody.Write(rr)

				So(err, ShouldNotBeNil)
				var responseBundle models.Bundle
				err = json.NewDecoder(rr.Body).Decode(&responseBundle)
				So(err, ShouldNotBeNil)

				So(rr.Header().Get(HeaderContentType), ShouldEqual, string(ResponseBodyTypeJson))
			})
		})
	})

	Convey("When the type is not supported", t, func() {
		var unsupportedResponseBodyType ResponseBodyType = "unsupported"

		responseBody := createResponseBody(unsupportedResponseBodyType, testBundle)
		Convey("Write should return an error", func() {
			rr := httptest.NewRecorder()
			err := responseBody.Write(rr)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "not implemented")
			var responseBundle models.Bundle
			err = json.NewDecoder(rr.Body).Decode(&responseBundle)
			So(err, ShouldNotBeNil)

			So(rr.Header().Get(HeaderContentType), ShouldEqual, string(unsupportedResponseBodyType))
		})

	})
}
