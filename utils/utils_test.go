package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

// errorWriter simulates a failure when writing to the response body
type errorWriter struct {
	http.ResponseWriter
}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, errors.New("simulated write error")
}
func (e *errorWriter) WriteString(s string) (int, error) {
	return 0, errors.New("simulated write error")
}

var codebadRequest = models.CodeBadRequest

func TestHandleBundleAPIErr_Success(t *testing.T) {
	Convey("Given a valid error object and HTTP status code", t, func() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		errInfo := &models.Error{
			Code:        &codebadRequest,
			Description: "Invalid request",
		}

		Convey("When HandleBundleAPIErr is called", func() {
			HandleBundleAPIErr(w, r, errInfo, http.StatusBadRequest)

			Convey("Then it should write the correct response", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")

				var response models.Error
				err := json.NewDecoder(w.Body).Decode(&response)
				So(err, ShouldBeNil)
				So(&response, ShouldResemble, errInfo)
			})
		})
	})
}

func TestHandleBundleAPIErr_Failure(t *testing.T) {
	Convey("Given an invalid error object", t, func() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)

		Convey("When HandleBundleAPIErr is called", func() {
			HandleBundleAPIErr(w, r, nil, http.StatusBadRequest)

			Convey("Then it should return an internal server error", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")

				var response models.Error
				err := json.NewDecoder(w.Body).Decode(&response)
				So(err, ShouldBeNil)
				So(response.Code.String(), ShouldEqual, models.CodeInternalServerError.String())
				So(response.Description, ShouldEqual, "Failed to process the request due to an internal error")
			})
		})
	})

	Convey("Given a response write that fails during encoding", t, func() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		errInfo := &models.Error{
			Code:        &codebadRequest,
			Description: "Invalid request",
		}

		Convey("When the response writer fails", func() {
			errorWriter := &errorWriter{ResponseWriter: w}

			HandleBundleAPIErr(errorWriter, r, errInfo, http.StatusBadRequest)

			Convey("Then it should log the write error", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}
