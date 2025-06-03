package utils

import (
	"encoding/json"
	"net/http"

	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
)

type ResponseBuilder struct {
	headers    map[string]string
	etag       *string
	body       *interface{}
	statusCode int
}

func CreateResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		headers: make(map[string]string),
	}
}

func (r *ResponseBuilder) WithHeader(key string, value string) *ResponseBuilder {
	r.headers[key] = value
	return r
}

func (r *ResponseBuilder) WithETag(value string) *ResponseBuilder {
	r.etag = &value
	return r
}

func (r *ResponseBuilder) WithCacheControl(value CacheControl) *ResponseBuilder {
	r.headers[HeaderCacheControl] = value.String()
	return r
}

func (r *ResponseBuilder) WithJsonBody(body *interface{}) *ResponseBuilder {
	r.body = body
	return r
}

func (r *ResponseBuilder) WithStatusCode(statusCode int) *ResponseBuilder {
	r.statusCode = statusCode
	return r
}

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

	if r.body == nil {
		return nil
	}

	err := r.writeResponseBody(responseWriter)
	if err != nil {
		return err
	}

	return nil
}

func (r *ResponseBuilder) writeResponseBody(responseWriter http.ResponseWriter) error {
	body, err := r.marshalBody()

	if err != nil {
		return err
	}

	if _, err := responseWriter.Write(*body); err != nil {
		return err
	}

	return nil
}

func (r *ResponseBuilder) marshalBody() (*[]byte, error) {
	body, err := json.Marshal(r.body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}
