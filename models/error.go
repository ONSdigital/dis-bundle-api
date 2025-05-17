package models

import (
	"encoding/json"
	"fmt"
)

type Error struct {
	Code        Code   `bson:"code" json:"code"`
	Description string `bson:"description" json:"description"`
	Source      Source `bson:"source" json:"source"`
}

type ErrorList struct {
	Errors []Error `bson:"errors" json:"errors"`
}

type Source struct {
	Field     string `bson:"field" json:"field"`
	Parameter string `bson:"parameter" json:"parameter"`
	Header    string `bson:"header" json:"header"`
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
)

// IsValid validates that the Code is a valid enum value
func (c Code) IsValid() bool {
	switch c {
	case CodeInternalServerError, CodeNotFound, CodeBadRequest, CodeUnauthorized, CodeForbidden, CodeConflict:
		return true
	default:
		return false
	}
}

// String returns the string value of the Code
func (c Code) String() string {
	return string(c)
}

// MarshalJSON marshals the Code to JSON
func (c Code) MarshalJSON() ([]byte, error) {
	if !c.IsValid() {
		return nil, fmt.Errorf("invalid Code: %s", c)
	}
	return json.Marshal(string(c))
}

// UnmarshalJSON unmarshals a string to Code
func (c *Code) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	converted := Code(str)
	if !converted.IsValid() {
		return fmt.Errorf("invalid Code: %s", str)
	}
	*c = converted
	return nil
}
