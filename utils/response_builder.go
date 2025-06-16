package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
)

// HTTPResponseBuilder provides a fluent interface for building HTTP responses.
// It allows chaining method calls to configure headers, body, ETag, cache control,
// and status code before writing the response.
type HTTPResponseBuilder interface {
	// WithHeader adds a custom header to the HTTP response
	WithHeader(key string, value string) HTTPResponseBuilder

	// WithETag sets the ETag header for the HTTP response.
	WithETag(value string) HTTPResponseBuilder

	// WithCacheControl sets the Cache-Control header for the HTTP response.
	WithCacheControl(value CacheControl) HTTPResponseBuilder

	// WithJSONBody sets the response body as JSON content.
	WithJSONBody(body any) HTTPResponseBuilder

	// WithBody sets the response body with a specific content type.
	WithBody(bodyType ContentType, body any) HTTPResponseBuilder

	// WithStatusCode sets the status code of the HTTP response
	WithStatusCode(statusCode int) HTTPResponseBuilder

	// Build constructs and writes the HTTP response with the values that were set in the various With... methods
	Build(responseWriter http.ResponseWriter) error
}

type ResponseWriter interface {
	Write(responseWriter http.ResponseWriter) error
}

// Concrete implementation of HTTPResponseBuilder
type ResponseBuilder struct {
	headers    map[string]string // HTTP headers to write in the response
	etag       *string           // ETag value to write in the response headers
	body       *ResponseBody     // Response body to write
	statusCode int               // Response status code
}

// Interface checks
var _ HTTPResponseBuilder = (*ResponseBuilder)(nil)
var _ ResponseWriter = (*ResponseBody)(nil)

// ContentType is the content type of the response body to write
type ContentType string

const (
	ContentTypeJSON ContentType = "application/json"
)

// ResponseBody encapsulates the response body content and its content type
type ResponseBody struct {
	responseType ContentType
	body         interface{}
}

func createResponseBody(responseType ContentType, body any) *ResponseBody {
	return &ResponseBody{
		responseType: responseType,
		body:         body,
	}
}

func CreateHTTPResponseBuilder() HTTPResponseBuilder {
	return &ResponseBuilder{
		headers: make(map[string]string),
	}
}

// WithHeader adds a custom header for writing to the HTTP response
func (r *ResponseBuilder) WithHeader(key, value string) HTTPResponseBuilder {
	r.headers[key] = value
	return r
}

// WithETag sets the ETag header for writing to the HTTP response
func (r *ResponseBuilder) WithETag(value string) HTTPResponseBuilder {
	r.etag = &value
	return r
}

// WithCacheControl sets the Cache-Control header for writing to the HTTP response
func (r *ResponseBuilder) WithCacheControl(value CacheControl) HTTPResponseBuilder {
	r.headers[HeaderCacheControl] = value.String()
	return r
}

// WithJSONBody sets the body for the HTTP response, to be marshalled as a JSON response
func (r *ResponseBuilder) WithJSONBody(body any) HTTPResponseBuilder {
	return r.WithBody(ContentTypeJSON, body)
}

// WithBody sets the body for the HTTP response, with the specified body content type
func (r *ResponseBuilder) WithBody(contentType ContentType, body any) HTTPResponseBuilder {
	r.body = createResponseBody(contentType, body)

	return r
}

// WithStatusCode sets the HTTP status code for the response
func (r *ResponseBuilder) WithStatusCode(statusCode int) HTTPResponseBuilder {
	r.statusCode = statusCode
	return r
}

// Build constructs and writes the HTTP response to the provided ResponseWriter.
func (r *ResponseBuilder) Build(responseWriter http.ResponseWriter) error {
	if r.etag != nil {
		dpresponse.SetETag(responseWriter, *r.etag)
	}

	for key, value := range r.headers {
		responseWriter.Header().Set(key, value)
	}

	if r.statusCode > 0 {
		responseWriter.WriteHeader(r.statusCode)
	}

	if r.body != nil {
		err := r.body.Write(responseWriter)
		if err != nil {
			return err
		}
	}

	return nil
}

// Write writes the ResponseBody's body to the ResponseWriter
func (rb *ResponseBody) Write(r http.ResponseWriter) error {
	var body *[]byte
	var err error

	r.Header().Set(HeaderContentType, string(rb.responseType))

	switch rb.responseType {
	case ContentTypeJSON:
		body, err = rb.marshalJSONBody()
	default:
		err = fmt.Errorf("response body type %s is not implemented", rb.responseType)
	}

	if err != nil {
		return err
	}

	if _, err := r.Write(*body); err != nil {
		return err
	}

	return nil
}

// marshalJSONBody marshals the ResponseBody's body object to a *[]byte
func (rb *ResponseBody) marshalJSONBody() (*[]byte, error) {
	if rb.body == nil {
		return nil, errors.New("body is nil")
	}

	body, err := json.Marshal(rb.body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}
