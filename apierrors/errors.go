package apierrors

import (
	"errors"
	"fmt"
)

// Custom error types for common HTTP cases
type ErrInvalidPatch struct {
	Msg string
}

func (e ErrInvalidPatch) Error() string {
	return e.Msg
}

// Response error descriptions
const (
	ErrorDescriptionMalformedRequest  = "Unable to process request due to a malformed or invalid request body or query parameter"
	ErrorDescriptionMissingParameters = "Unable to process request due to missing required parameters in the request body or query parameters"
	ErrorDescriptionNotFound          = "The requested resource does not exist"
	ErrorDescriptionInternalError     = "Failed to process the request due to an internal error"

	// Requests
	ErrorDescriptionNoParametersFound = "No URL parameters found"
	ErrorDescriptionETagMissing       = "ETag header is required"
	ErrorDescriptionETagMismatch      = "Supplied ETag does not match"

	// Bundle specific errors
	ErrorDescriptionNoContentItemsFound = "No content items found"
	ErrorDescriptionMissingBundleID     = "missing bundle ID"
	ErrorDescriptionEmptyBundleID       = "Bundle ID cannot be empty"

	// Validation
	ErrorDescriptionValidationEventFailed = "Failed to validate event"

	// Auth
	ErrorDescriptionUserIdentityParseFailed = "Failed to get user identity from JWT"
	ErrorDescriptionNoTokenFound            = "No Authorisation token was found for the request"
)

// CreateMissingRouteVariableError creates a standard error message string for if a route variable is missing
func CreateMissingRouteVariableError(variable string) string {
	return fmt.Sprintf("Missing %s", variable)
}

// CreateMissingRouteVariableError creates a standard error message string for if a route variable is provided but empty
func CreateEmptyRouteVariableError(variable string) string {
	return fmt.Sprintf("%s cannot be empty", variable)
}

var (
	ErrUnableToReadMessage = errors.New("failed to read message body")
	ErrUnableToParseJSON   = errors.New("failed to parse json body")
)

const (
	ErrUnmarshalJSONObject    = "Failed to unmarshal bundle resource into bytes"
	ErrMarshalJSONObject      = "Failed to Marshal bundle resource into bytes"
	ErrWritingBytesToResponse = "Failed writing bytes to response"
)

// Core errors for dis-bundle-api
var (
	// Generic Errors
	ErrInternalServer = errors.New("internal error")
	ErrInvalidBody    = errors.New("invalid request body")
	ErrNotFound       = errors.New("not found")
	ErrUnauthorised   = errors.New("unauthorised access to API")

	// Bundle-Specific
	ErrBundleNotFound         = errors.New("bundle not found")
	ErrDeleteBundleForbidden  = errors.New("cannot delete a published bundle")
	ErrBundleAlreadyExists    = errors.New("bundle already exists")
	ErrInvalidBundleState     = errors.New("invalid bundle state")
	ErrMissingBundleID        = errors.New(ErrorDescriptionMissingBundleID)
	ErrInvalidBundleReference = errors.New("invalid bundle reference")
	ErrBundleEventNotFound    = errors.New("bundle event not found")

	// Validation
	ErrMissingParameters      = errors.New("missing required parameters in request")
	ErrInvalidQueryParameter  = errors.New("invalid query parameter")
	ErrTooManyQueryParameters = errors.New("too many query parameters provided")
	ErrMismatchedETag         = errors.New("etag does not match")

	// State errors
	ErrExpectedStateOfCreated  = errors.New("expected bundle state to be 'CREATED'")
	ErrExpectedStateOfApproved = errors.New("expected bundle state to be 'APPROVED'")
)

// 404 Not Found
var NotFoundMap = map[error]bool{
	ErrBundleNotFound:      true,
	ErrBundleEventNotFound: true,
}

// 400 Bad Request
var BadRequestMap = map[error]bool{
	ErrInvalidBody:            true,
	ErrMissingParameters:      true,
	ErrInvalidQueryParameter:  true,
	ErrTooManyQueryParameters: true,
	ErrMissingBundleID:        true,
	ErrInvalidBundleReference: true,
	ErrInvalidBundleState:     true,
}

// 409 Conflict
var ConflictMap = map[error]bool{
	ErrBundleAlreadyExists: true,
}

// 403 Forbidden
var ForbiddenMap = map[error]bool{
	ErrDeleteBundleForbidden:  true,
	ErrExpectedStateOfCreated: true,
}
