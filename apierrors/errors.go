package apierrors

import "errors"

// Custom error types for common HTTP cases
type ErrInvalidPatch struct {
	Msg string
}

func (e ErrInvalidPatch) Error() string {
	return e.Msg
}

var (
	ErrUnableToReadMessage = errors.New("failed to read message body")
	ErrUnableToParseJSON   = errors.New("failed to parse json body")
	ErrUnableToParseTime   = errors.New("failed to parse time from json body")
	ErrScheduledAtRequired = errors.New("scheduled_at is required for scheduled bundles")
	ErrScheduledAtSet      = errors.New("scheduled_at should not be set for manual bundles")
	ErrScheduledAtInPast   = errors.New("scheduled_at cannot be in the past")
)

var (
	ErrDescription              = "Unable to process request due to a malformed or invalid request body or query parameter."
	ErrInternalErrorDescription = "An internal error occurred."
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
	ErrMissingBundleID        = errors.New("missing bundle ID")
	ErrInvalidBundleReference = errors.New("invalid bundle reference")
	ErrBundleEventNotFound    = errors.New("bundle event not found")

	// Validation
	ErrMissingParameters      = errors.New("missing required parameters in request")
	ErrInvalidQueryParameter  = errors.New("invalid query parameter")
	ErrTooManyQueryParameters = errors.New("too many query parameters provided")

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
