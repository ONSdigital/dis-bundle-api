package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
	"github.com/ONSdigital/log.go/v2/log"
)

// Error represents the details of a specific error
type Error struct {
	Code        *Code   `bson:"code,omitempty"        json:"code,omitempty"`
	Description string  `bson:"description,omitempty" json:"description,omitempty"`
	Source      *Source `bson:"source,omitempty"      json:"source,omitempty"`
}

// ErrorList represents a list of errors
type ErrorList struct {
	Errors *[]Error `bson:"errors,omitempty" json:"errors,omitempty"`
}

// Source represents the details of which field or parameter the error relates to. Used to return validation errors to 4xx requests. Only one of the properties below can be returned in any single error.
type Source struct {
	Field     string `bson:"field,omitempty"     json:"field,omitempty"`
	Parameter string `bson:"parameter,omitempty" json:"parameter,omitempty"`
	Header    string `bson:"header,omitempty"    json:"header,omitempty"`
}

func CreateError(reader io.Reader) (*Error, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, errs.ErrUnableToReadMessage
	}

	var errorObj Error

	err = json.Unmarshal(b, &errorObj)
	if err != nil {
		return nil, errs.ErrUnableToParseJSON
	}

	return &errorObj, nil
}

func ValidateError(e *Error) error {
	if e == nil {
		return fmt.Errorf("error cannot be nil")
	}

	if e.Code != nil && !e.Code.IsValid() {
		return fmt.Errorf("invalid error code: %s", e.Code.String())
	}

	err := fmt.Errorf("only one of Source.Field, Source.Parameter, Source.Header can be set")
	if e.Source != nil {
		count := 0
		if e.Source.Field != "" {
			count++
		}
		if e.Source.Parameter != "" {
			count++
		}
		if e.Source.Header != "" {
			count++
		}
		if count > 1 {
			return err
		}
	}
	return nil
}

func HandleErr(w http.ResponseWriter, r *http.Request, code Code, description string, httpStatusCode int, source *Source) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	errBytes, errCheck := json.Marshal(Error{Code: &code, Description: description, Source: source})
	if errCheck != nil {
		log.Error(r.Context(), "api endpoint error writing error body", errCheck)
	}
	http.Error(w, string(errBytes), httpStatusCode)
}

// Code enum representing the error code
type Code string

// Define possible values for the Code enum
const (
	CodeInternalServerError Code = "internal_server_error"
	CodeNotFound            Code = "not_found"
	CodeBadRequest          Code = "bad_request"
	CodeUnauthorized        Code = "unauthorized"
	CodeForbidden           Code = "forbidden"
	CodeConflict            Code = "conflict"
	JSONMarshalError        Code = "JSONMarshalError"
	JSONUnmarshalError      Code = "JSONUnmarshalError"
	WriteResponseError      Code = "WriteResponseError"
)

// IsValid validates that the Code is a valid enum value
func (c Code) IsValid() bool {
	switch c {
	case CodeInternalServerError, CodeNotFound, CodeBadRequest, CodeUnauthorized, CodeForbidden, CodeConflict, JSONMarshalError, JSONUnmarshalError, WriteResponseError:
		return true
	default:
		return false
	}
}

// String returns the string value of the Code
func (c Code) String() string {
	return string(c)
}
