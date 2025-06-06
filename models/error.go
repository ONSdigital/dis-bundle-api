package models

import (
	"encoding/json"
	"fmt"
	"io"

	errs "github.com/ONSdigital/dis-bundle-api/apierrors"
)

// Error represents the details of a specific error
type Error struct {
	Code        *Code   `json:"code,omitempty"`
	Description string  `json:"description,omitempty"`
	Source      *Source `json:"source,omitempty"`
}

// ErrorList represents a list of errors
type ErrorList struct {
	Errors []*Error `json:"errors"`
}

// Source represents the details of which field or parameter the error relates to. Used to return validation errors to 4xx requests. Only one of the properties below can be returned in any single error.
type Source struct {
	Field     string `json:"field,omitempty"`
	Parameter string `json:"parameter,omitempty"`
	Header    string `json:"header,omitempty"`
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
	CodeMissingParameters   Code = "missing_parameters"
	CodeInvalidParameters   Code = "invalid_parameters"
	JSONMarshalError        Code = "JSONMarshalError"
	JSONUnmarshalError      Code = "JSONUnmarshalError"
	WriteResponseError      Code = "WriteResponseError"
)

// IsValid validates that the Code is a valid enum value
func (c Code) IsValid() bool {
	switch c {
	case CodeInternalServerError, CodeNotFound, CodeBadRequest, CodeUnauthorized, CodeForbidden, CodeConflict, CodeMissingParameters, CodeInvalidParameters, JSONMarshalError, JSONUnmarshalError, WriteResponseError:
		return true
	default:
		return false
	}
}

// String returns the string value of the Code
func (c Code) String() string {
	return string(c)
}
