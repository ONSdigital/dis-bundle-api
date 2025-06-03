package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	dpresponse "github.com/ONSdigital/dp-net/v3/handlers/response"
)

type HTTPResponseBuilder interface {
	WithHeader(key string, value string) HTTPResponseBuilder
	WithETag(value string) HTTPResponseBuilder
	WithCacheControl(value CacheControl) HTTPResponseBuilder
	WithJsonBody(body any) HTTPResponseBuilder
	WithBody(bodyType ResponseBodyType, body any) HTTPResponseBuilder
	WithStatusCode(statusCode int) HTTPResponseBuilder
	Build(responseWriter http.ResponseWriter) error
}

type ResponseWriter interface {
	Write(responseWriter http.ResponseWriter) error
}

type ResponseBuilder struct {
	headers    map[string]string
	etag       *string
	body       *ResponseBody
	statusCode int
}

var _ HTTPResponseBuilder = (*ResponseBuilder)(nil)
var _ ResponseWriter = (*ResponseBody)(nil)

type ResponseBodyType string

const (
	ResponseBodyTypeJson ResponseBodyType = "application/json"
)

type ResponseBody struct {
	responseType ResponseBodyType
	body         interface{}
}

func createResponseBody(responseType ResponseBodyType, body any) *ResponseBody {
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

func (r *ResponseBuilder) WithHeader(key string, value string) HTTPResponseBuilder {
	r.headers[key] = value
	return r
}

func (r *ResponseBuilder) WithETag(value string) HTTPResponseBuilder {
	r.etag = &value
	return r
}

func (r *ResponseBuilder) WithCacheControl(value CacheControl) HTTPResponseBuilder {
	r.headers[HeaderCacheControl] = value.String()
	return r
}

func (r *ResponseBuilder) WithJsonBody(body any) HTTPResponseBuilder {
	return r.WithBody(ResponseBodyTypeJson, body)
}

func (r *ResponseBuilder) WithBody(bodyType ResponseBodyType, body any) HTTPResponseBuilder {
	r.body = createResponseBody(bodyType, body)

	return r
}

func (r *ResponseBuilder) WithStatusCode(statusCode int) HTTPResponseBuilder {
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

	if r.body != nil {
		err := r.body.Write(responseWriter)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rb *ResponseBody) Write(r http.ResponseWriter) error {
	var body *[]byte
	var err error

	r.Header().Set(HeaderContentType, string(rb.responseType))

	switch rb.responseType {
	case ResponseBodyTypeJson:
		body, err = rb.marshalJsonBody()
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

func (r *ResponseBody) marshalJsonBody() (*[]byte, error) {
	body, err := json.Marshal(r.body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}
