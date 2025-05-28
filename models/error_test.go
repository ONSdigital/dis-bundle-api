package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	. "github.com/smartystreets/goconvey/convey"
)

type ErrorReader struct{}

func (e *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("mock read error")
}

var internalServerErrorCode = CodeInternalServerError
var invalidErrorCode = Code("invalid_code")

var exampleError = Error{
	Code:        &internalServerErrorCode,
	Description: "An example error occurred",
	Source:      &Source{Field: "example_field"},
}

var allCodes = []Code{
	CodeInternalServerError,
	CodeNotFound,
	CodeBadRequest,
	CodeUnauthorized,
	CodeForbidden,
	CodeConflict,
}

func TestCreateError_Success(t *testing.T) {
	Convey("Given a valid Error", t, func() {
		data, err := json.Marshal(exampleError)
		So(err, ShouldBeNil)

		reader := bytes.NewReader(data)

		Convey("When CreateError is called", func() {
			result, err := CreateError(reader)

			Convey("Then it should return the Error without any error", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.Code, ShouldEqual, &internalServerErrorCode)
				So(result.Description, ShouldEqual, "An example error occurred")
				So(result.Source.Field, ShouldEqual, "example_field")
				So(result.Source.Parameter, ShouldEqual, "")
				So(result.Source.Header, ShouldEqual, "")
			})
		})
	})
}

func TestCreateError_Failure(t *testing.T) {
	Convey("Given a reader that returns an error", t, func() {
		reader := &ErrorReader{}

		Convey("When CreateError is called", func() {
			_, err := CreateError(reader)

			Convey("Then it should return an error indicating it was unable to read the message", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errs.ErrUnableToReadMessage.Error())
			})
		})
	})

	Convey("Given a reader with invalid JSON", t, func() {
		invalidJSON := `{"code": "123}`
		reader := bytes.NewReader([]byte(invalidJSON))

		Convey("When CreateError is called", func() {
			_, err := CreateError(reader)

			Convey("Then it should return an error indicating it was unable to parse JSON", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, errs.ErrUnableToParseJSON.Error())
			})
		})
	})
}

func TestValidateError_Success(t *testing.T) {
	Convey("Given a valid Error", t, func() {
		Convey("When ValidateError is called", func() {
			err := ValidateError(&exampleError)

			Convey("Then there should be no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestValidateError_Failure(t *testing.T) {
	Convey("Given an Error that is nil", t, func() {
		Convey("When ValidateError is called", func() {
			err := ValidateError(nil)

			Convey("Then it should return an error indicating the error cannot be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "error cannot be nil")
			})
		})
	})

	Convey("Given an Error with an invalid code", t, func() {
		errorWithInvalidCode := Error{
			Code:        &invalidErrorCode,
			Description: "An error with an invalid code",
			Source:      &Source{Field: "example_field"},
		}

		Convey("When ValidateError is called", func() {
			err := ValidateError(&errorWithInvalidCode)

			Convey("Then it should return an error indicating the code is invalid", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "invalid error code: invalid_code")
			})
		})
	})

	Convey("Given an Error with multiple Source fields set", t, func() {
		errorWithMultipleSources := Error{
			Code:        &internalServerErrorCode,
			Description: "An error with multiple source fields",
			Source:      &Source{Field: "field_name", Parameter: "param_name", Header: "header_name"},
		}

		Convey("When ValidateError is called", func() {
			err := ValidateError(&errorWithMultipleSources)

			Convey("Then it should return an error indicating only one Source field can be set", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "only one of Source.Field, Source.Parameter, Source.Header can be set")
			})
		})
	})
}

func TestCode_IsValid_Success(t *testing.T) {
	Convey("Given a valid Code", t, func() {
		for _, code := range allCodes {
			Convey(fmt.Sprintf("When IsValid is called for code %s", code), func() {
				valid := code.IsValid()

				Convey("Then it should return true", func() {
					So(valid, ShouldBeTrue)
				})
			})
		}
	})
}

func TestCode_IsValid_Failure(t *testing.T) {
	Convey("Given an invalid Code", t, func() {
		Convey("When IsValid is called", func() {
			valid := invalidErrorCode.IsValid()

			Convey("Then it should return false", func() {
				So(valid, ShouldBeFalse)
			})
		})
	})
}
