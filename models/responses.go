package models

type SuccessResponse struct {
	Body    []byte            `json:"-"`
	Status  int               `json:"-"`
	Headers map[string]string `json:"-"`
}

func NewSuccessResponse(jsonBody []byte, statusCode int, headers map[string]string) *SuccessResponse {
	return &SuccessResponse{
		Body:    jsonBody,
		Status:  statusCode,
		Headers: headers,
	}
}
