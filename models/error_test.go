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

func TestCodeIsValid(t *testing.T) {
	Convey("Given a valid Code: CodeInternalServerError", t, func() {
		code := CodeInternalServerError
		Convey("It should return true", func() {
			So(code.IsValid(), ShouldBeTrue)
		})
	})

	Convey("Given a valid Code: CodeNotFound", t, func() {
		code := CodeNotFound
		Convey("It should return true", func() {
			So(code.IsValid(), ShouldBeTrue)
		})
	})

	Convey("Given a valid Code: CodeBadRequest", t, func() {
		code := CodeBadRequest
		Convey("It should return true", func() {
			So(code.IsValid(), ShouldBeTrue)
		})
	})

	Convey("Given a valid Code: CodeUnauthorized", t, func() {
		code := CodeUnauthorized
		Convey("It should return true", func() {
			So(code.IsValid(), ShouldBeTrue)
		})
	})

	Convey("Given a valid Code: CodeForbidden", t, func() {
		code := CodeForbidden
		Convey("It should return true", func() {
			So(code.IsValid(), ShouldBeTrue)
		})
	})

	Convey("Given a valid Code: CodeConflict", t, func() {
		code := CodeConflict
		Convey("It should return true", func() {
			So(code.IsValid(), ShouldBeTrue)
		})
	})

	Convey("Given an invalid Code", t, func() {
		code := Code("invalid_code")
		Convey("It should return false", func() {
			So(code.IsValid(), ShouldBeFalse)
		})
	})
}

func TestCodeString(t *testing.T) {
	Convey("Given a valid Code: CodeInternalServerError", t, func() {
		code := CodeInternalServerError
		Convey("It should return 'internal_server_error'", func() {
			So(code.String(), ShouldEqual, "internal_server_error")
		})
	})

	Convey("Given a valid Code: CodeNotFound", t, func() {
		code := CodeNotFound
		Convey("It should return 'not_found'", func() {
			So(code.String(), ShouldEqual, "not_found")
		})
	})

	Convey("Given a valid Code: CodeBadRequest", t, func() {
		code := CodeBadRequest
		Convey("It should return 'bad_request'", func() {
			So(code.String(), ShouldEqual, "bad_request")
		})
	})

	Convey("Given a valid Code: CodeUnauthorized", t, func() {
		code := CodeUnauthorized
		Convey("It should return 'unauthorized'", func() {
			So(code.String(), ShouldEqual, "unauthorized")
		})
	})

	Convey("Given a valid Code: CodeForbidden", t, func() {
		code := CodeForbidden
		Convey("It should return 'forbidden'", func() {
			So(code.String(), ShouldEqual, "forbidden")
		})
	})

	Convey("Given a valid Code: CodeConflict", t, func() {
		code := CodeConflict
		Convey("It should return 'conflict'", func() {
			So(code.String(), ShouldEqual, "conflict")
		})
	})
}

func TestCodeMarshalJSON(t *testing.T) {
	Convey("Given a valid Code: CodeInternalServerError", t, func() {
		code := CodeInternalServerError
		Convey("It should marshal to '\"internal_server_error\"'", func() {
			data, err := code.MarshalJSON()
			So(err, ShouldBeNil)
			So(string(data), ShouldEqual, `"internal_server_error"`)
		})
	})

	Convey("Given a valid Code: CodeNotFound", t, func() {
		code := CodeNotFound
		Convey("It should marshal to '\"not_found\"'", func() {
			data, err := code.MarshalJSON()
			So(err, ShouldBeNil)
			So(string(data), ShouldEqual, `"not_found"`)
		})
	})

	Convey("Given an invalid Code", t, func() {
		code := Code("invalid_code")
		Convey("It should return an error", func() {
			data, err := code.MarshalJSON()
			So(err, ShouldNotBeNil)
			So(data, ShouldBeNil)
			So(err.Error(), ShouldEqual, "invalid Code: invalid_code")
		})
	})
}

func TestCodeUnmarshalJSON(t *testing.T) {
	Convey("Given a valid JSON string for Code: CodeInternalServerError", t, func() {
		var code Code
		data := []byte(`"internal_server_error"`)
		Convey("It should unmarshal successfully", func() {
			err := json.Unmarshal(data, &code)
			So(err, ShouldBeNil)
			So(code, ShouldEqual, CodeInternalServerError)
		})
	})

	Convey("Given a valid JSON string for Code: CodeNotFound", t, func() {
		var code Code
		data := []byte(`"not_found"`)
		Convey("It should unmarshal successfully", func() {
			err := json.Unmarshal(data, &code)
			So(err, ShouldBeNil)
			So(code, ShouldEqual, CodeNotFound)
		})
	})

	Convey("Given an invalid JSON string for Code", t, func() {
		var code Code
		data := []byte(`"invalid_code"`)
		Convey("It should return an error", func() {
			err := json.Unmarshal(data, &code)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "invalid Code: invalid_code")
		})
	})

	Convey("Given an invalid JSON for Code", t, func() {
		var code Code
		data := []byte(`123`) // Invalid JSON for a string
		Convey("It should return a JSON unmarshal error", func() {
			err := json.Unmarshal(data, &code)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestCodeOmitEmpty(t *testing.T) {
	Convey("Given an Error struct with an empty Code", t, func() {
		err := Error{
			Description: "Some description",
			Source:      &Source{Field: "field_name"},
		}

		Convey("When marshaled to JSON", func() {
			data, marshalErr := json.Marshal(err)

			Convey("The 'code' field should be omitted", func() {
				So(marshalErr, ShouldBeNil)
				So(string(data), ShouldNotContainSubstring, `"code"`)
			})
		})
	})

	Convey("Given an Error struct with a non-empty Code", t, func() {
		code := CodeInternalServerError
		err := Error{
			Code:        &code,
			Description: "Some description",
			Source:      &Source{Field: "field_name"},
		}

		Convey("When marshaled to JSON", func() {
			data, marshalErr := json.Marshal(err)

			Convey("The 'code' field should be present", func() {
				So(marshalErr, ShouldBeNil)
				So(string(data), ShouldContainSubstring, `"code":"internal_server_error"`)
			})
		})
	})
}

func TestDescriptionOmitEmpty(t *testing.T) {
	Convey("Given an Error struct with an empty Description", t, func() {
		code := CodeInternalServerError
		err := Error{
			Code:   &code,
			Source: &Source{Field: "field_name"},
		}

		Convey("When marshaled to JSON", func() {
			data, marshalErr := json.Marshal(err)

			Convey("The 'description' field should be omitted", func() {
				So(marshalErr, ShouldBeNil)
				So(string(data), ShouldNotContainSubstring, `"description"`)
			})
		})
	})

	Convey("Given an Error struct with a non-empty Description", t, func() {
		code := CodeInternalServerError
		err := Error{
			Code:        &code,
			Description: "Some description",
			Source:      &Source{Field: "field_name"},
		}

		Convey("When marshaled to JSON", func() {
			data, marshalErr := json.Marshal(err)

			Convey("The 'description' field should be present", func() {
				So(marshalErr, ShouldBeNil)
				So(string(data), ShouldContainSubstring, `"description":"Some description"`)
			})
		})
	})
}

func TestSourceOmitEmpty(t *testing.T) {
	Convey("Given an Error struct with an empty Source", t, func() {
		code := CodeInternalServerError
		err := Error{
			Code:        &code,
			Description: "Some description",
		}

		Convey("When marshaled to JSON", func() {
			data, marshalErr := json.Marshal(err)

			Convey("The 'source' field should be omitted", func() {
				So(marshalErr, ShouldBeNil)
				So(string(data), ShouldNotContainSubstring, `"source"`)
			})
		})
	})

	Convey("Given an Error struct with a non-empty Source", t, func() {
		code := CodeInternalServerError
		err := Error{
			Code:        &code,
			Description: "Some description",
			Source:      &Source{Field: "field_name"},
		}

		Convey("When marshaled to JSON", func() {
			data, marshalErr := json.Marshal(err)

			Convey("The 'source' field should be present", func() {
				So(marshalErr, ShouldBeNil)
				So(string(data), ShouldContainSubstring, `"source":{"field":"field_name"}`)
			})
		})
	})
}

func TestErrorListOmitEmpty(t *testing.T) {
	Convey("Given an ErrorList struct with an empty Errors slice", t, func() {
		errList := ErrorList{
			Errors: []Error{},
		}

		Convey("When marshaled to JSON", func() {
			data, marshalErr := json.Marshal(errList)

			Convey("The 'errors' field should be omitted", func() {
				So(marshalErr, ShouldBeNil)
				So(string(data), ShouldNotContainSubstring, `"errors"`)
			})
		})
	})

	Convey("Given an ErrorList struct with a non-empty Errors slice", t, func() {
		code := CodeInternalServerError
		errList := ErrorList{
			Errors: []Error{
				{
					Code:        &code,
					Description: "Some description",
					Source:      &Source{Field: "field_name"},
				},
			},
		}

		Convey("When marshaled to JSON", func() {
			data, marshalErr := json.Marshal(errList)

			Convey("The 'errors' field should be present", func() {
				So(marshalErr, ShouldBeNil)
				So(string(data), ShouldContainSubstring, `"errors":[{"code":"internal_server_error","description":"Some description","source":{"field":"field_name"}}]`)
			})
		})
	})
}

func TestCreateError(t *testing.T) {
	Convey("Given a valid Error struct", t, func() {
		code := CodeInternalServerError
		err := Error{
			Code:        &code,
			Description: "Some description",
			Source:      &Source{Field: "field_name"},
		}

		Convey("When CreateError is called", func() {
			data, err := json.Marshal(err)

			if err != nil {
				t.Logf("failed to marshal test data into bytes, error: %v", err)
				t.FailNow()
			}

			reader := bytes.NewReader(data)
			result, err := CreateError(reader)

			Convey("Then there should be no error", func() {
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.Code, ShouldEqual, &code)
				So(result.Description, ShouldEqual, "Some description")
				So(result.Source, ShouldNotBeNil)
				So(result.Source.Field, ShouldEqual, "field_name")
				So(result.Source.Parameter, ShouldEqual, "")
				So(result.Source.Header, ShouldEqual, "")
			})
		})
	})

	Convey("Return error when unable to read message", t, func() {
		Convey("when the reader returns an error", func() {
			reader := &ErrorReader{}
			_, err := CreateError(reader)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errs.ErrUnableToReadMessage.Error())
		})
	})

	Convey("Return error when unable to parse JSON", t, func() {
		Convey("when the JSON is invalid", func() {
			b := `{"code": "123}`
			reader := bytes.NewReader([]byte(b))
			_, err := CreateError(reader)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, errs.ErrUnableToParseJSON.Error())
		})
	})
}

func TestValidateError(t *testing.T) {
	Convey("Given a valid Error struct", t, func() {
		code := CodeInternalServerError
		e := Error{
			Code:        &code,
			Description: "Some description",
			Source:      &Source{Field: "field_name"},
		}

		Convey("When ValidateError is called", func() {
			err := ValidateError(&e)

			Convey("Then there should be no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a valid Error struct with two Source fields set", t, func() {
		code := CodeInternalServerError
		e := Error{
			Code:        &code,
			Description: "Some description",
			Source:      &Source{Parameter: "param_name", Header: "header_name"},
		}

		Convey("When ValidateError is called", func() {
			err := ValidateError(&e)

			Convey("Then there should be an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "only one of Source.Field, Source.Parameter, Source.Header can be set")
			})
		})
	})

	Convey("Given a valid Error struct with all Source fields set", t, func() {
		code := CodeInternalServerError
		e := Error{
			Code:        &code,
			Description: "Some description",
			Source:      &Source{Field: "field_name", Parameter: "param_name", Header: "header_name"},
		}

		Convey("When ValidateError is called", func() {
			err := ValidateError(&e)

			Convey("Then there should be an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "only one of Source.Field, Source.Parameter, Source.Header can be set")
			})
		})
	})

	Convey("Given a nil Error", t, func() {
		Convey("When ValidateError is called", func() {
			err := ValidateError(nil)

			Convey("Then there should be an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "error cannot be nil")
			})
		})
	})
}
