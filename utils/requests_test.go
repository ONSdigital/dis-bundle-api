package utils

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/dis-bundle-api/models"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func createRouteVariableTestRequest(key, value string) *http.Request {
	r, _ := http.NewRequest("GET", "/", nil)
	vars := map[string]string{}

	if key != "" {
		vars[key] = value
	}

	r = mux.SetURLVars(r, vars)
	return r

}

func TestGetRouteVariable(t *testing.T) {
	const (
		ExpectedKey   = "testing-key"
		ExpectedValue = "testing-value"
		MissingError  = "testing missing error"
		EmptyError    = "testing empty error"
	)

	Convey("When the route variable exists and is not empty", t, func() {
		r := createRouteVariableTestRequest(ExpectedKey, ExpectedValue)

		Convey("Then it should return the expected value", func() {
			value, err := GetRouteVariable(r, ExpectedKey, MissingError, EmptyError)

			So(err, ShouldBeNil)
			So(value, ShouldEqual, ExpectedValue)
		})
	})

	Convey("When the no route variables expst", t, func() {
		r := createRouteVariableTestRequest("", "")

		Convey("Then it should return the missing params error", func() {
			value, err := GetRouteVariable(r, ExpectedKey, MissingError, EmptyError)

			So(value, ShouldEqual, "")

			So(err, ShouldNotBeNil)
			So(*err.Code, ShouldEqual, models.CodeBadRequest)
			So(err.Description, ShouldEqual, apierrors.ErrorDescriptionMissingParameters)
		})
	})

	Convey("When the route variable is missing", t, func() {
		r := createRouteVariableTestRequest("some-other-key", "")

		Convey("Then it should return the missing error", func() {
			value, err := GetRouteVariable(r, ExpectedKey, MissingError, EmptyError)

			So(value, ShouldEqual, "")

			So(err, ShouldNotBeNil)
			So(*err.Code, ShouldEqual, models.CodeBadRequest)
			So(err.Description, ShouldEqual, MissingError)
		})
	})

	Convey("When the route variable is empty", t, func() {
		r := createRouteVariableTestRequest(ExpectedKey, "")

		Convey("Then it should return the empty error", func() {
			value, err := GetRouteVariable(r, ExpectedKey, MissingError, EmptyError)

			So(value, ShouldEqual, "")

			So(err, ShouldNotBeNil)
			So(*err.Code, ShouldEqual, models.CodeBadRequest)
			So(err.Description, ShouldEqual, EmptyError)
		})
	})
}

func createGetETagTestRequest(value *string) *http.Request {
	r, _ := http.NewRequest("GET", "/", nil)

	if value != nil {
		r.Header.Set(HeaderIfMatch, *value)
	}
	return r
}

func TestGetETag(t *testing.T) {
	Convey("When the etag header is set", t, func() {
		value := "test-etag-value"
		r := createGetETagTestRequest(&value)

		Convey("Then it should be returned with no error", func() {
			etag, err := GetETag(r)

			So(err, ShouldBeNil)
			So(etag, ShouldEqual, &value)
		})
	})

	Convey("When the etag header is not set", t, func() {
		r := createGetETagTestRequest(nil)

		Convey("Then it should be returned with no error", func() {
			etag, err := GetETag(r)

			So(etag, ShouldBeNil)

			So(err, ShouldNotBeNil)
			So(*err.Code, ShouldEqual, models.CodeBadRequest)
			So(err.Description, ShouldEqual, apierrors.ErrorDescriptionETagMissing)
		})
	})
}

type TestRequestBody struct {
	BundleID string `json:"bundle_id"`
	Version  int    `json:"version"`
}

func createGetRequestBodyTestRequest(body []byte) *http.Request {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	r, _ := http.NewRequest("POST", "/test", bodyReader)
	return r
}

func createValidJson(bundleId string, version int) []byte {
	jsonString := fmt.Sprintf(`{"bundle_id":"%s","version":%d}`, bundleId, version)

	return []byte(jsonString)
}

func TestGetRequestBody(t *testing.T) {
	const (
		BundleId = "bundle-1234"
		Version  = 1
	)
	Convey("When the request body contains valid JSON", t, func() {
		validJson := createValidJson(BundleId, Version)
		r := createGetRequestBodyTestRequest(validJson)

		Convey("Then it should parse successfully and return the struct", func() {
			result, err := GetRequestBody[TestRequestBody](r)

			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.BundleID, ShouldEqual, BundleId)
			So(result.Version, ShouldEqual, Version)
		})
	})

	Convey("When the request body is empty", t, func() {
		r := createGetRequestBodyTestRequest([]byte{})

		Convey("Then it should return an error", func() {
			result, err := GetRequestBody[TestRequestBody](r)

			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(*err.Code, ShouldEqual, models.CodeBadRequest)
			So(err.Description, ShouldEqual, apierrors.ErrorDescriptionMalformedRequest)
		})
	})

	Convey("When the request body contains invalid JSON", t, func() {
		invalidJSON := []byte(`{"bundle_id":"bundle","version":}`)
		r := createGetRequestBodyTestRequest(invalidJSON)

		Convey("Then it should return an error", func() {
			result, err := GetRequestBody[TestRequestBody](r)

			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(*err.Code, ShouldEqual, models.CodeBadRequest)
			So(err.Description, ShouldEqual, apierrors.ErrorDescriptionMalformedRequest)
		})
	})

	Convey("When the request body contains malformed JSON", t, func() {
		malformedJSON := []byte(`{this is not valid json}`)
		r := createGetRequestBodyTestRequest(malformedJSON)

		Convey("Then it should return an error", func() {
			result, err := GetRequestBody[TestRequestBody](r)

			So(result, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(*err.Code, ShouldEqual, models.CodeBadRequest)
			So(err.Description, ShouldEqual, apierrors.ErrorDescriptionMalformedRequest)
		})
	})
}
