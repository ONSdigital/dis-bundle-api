package models

import "net/http"

// Struct to represent an API error, and the HTTP Status code to return
type ErrorResult[TError Error] struct {
	Error          *TError
	HTTPStatusCode int
}

func CreateErrorResult[TError Error](err *TError, httpStatusCode int) *ErrorResult[TError] {
	return &ErrorResult[TError]{
		Error:          err,
		HTTPStatusCode: httpStatusCode,
	}
}

// Create an error response with a 400 status code
func CreateBadRequestErrorResult[TError Error](err *TError) *ErrorResult[TError] {
	return CreateErrorResult(err, http.StatusBadRequest)
}

// Create an error response with a 500 status code
func CreateInternalServerErrorResult[TError Error](err *TError) *ErrorResult[TError] {
	return CreateErrorResult(err, http.StatusInternalServerError)
}

// Create an error response with a 404 status code
func CreateNotFoundResult[TError Error](err *TError) *ErrorResult[TError] {
	return CreateErrorResult(err, http.StatusNotFound)
}

// Struct to represent a successful API result, and the HTTP status code to return
type SuccessResult[TResult any] struct {
	Result         *TResult
	HTTPStatusCode int
}

func CreateSuccessResult[TResult any](result *TResult, httpStatusCode int) *SuccessResult[TResult] {
	return &SuccessResult[TResult]{
		Result:         result,
		HTTPStatusCode: httpStatusCode,
	}
}

// Create a success result with a 200 status
func CreateOkResult[TResult any](result *TResult) *SuccessResult[TResult] {
	return &SuccessResult[TResult]{
		Result:         result,
		HTTPStatusCode: http.StatusOK,
	}
}
