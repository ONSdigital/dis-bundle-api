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

	// Validation
	ErrMissingParameters      = errors.New("missing required parameters in request")
	ErrInvalidQueryParameter  = errors.New("invalid query parameter")
	ErrTooManyQueryParameters = errors.New("too many query parameters provided")

	// State errors
	ErrExpectedStateOfCreated  = errors.New("expected bundle state to be 'CREATED'")
	ErrExpectedStateOfApproved = errors.New("expected bundle state to be 'APPROVED'")
)

// Grouping for error response handling
var NotFoundMap = map[error]bool{
	ErrBundleNotFound: true,
}

var BadRequestMap = map[error]bool{
	ErrInvalidBody:            true,
	ErrMissingParameters:      true,
	ErrInvalidQueryParameter:  true,
	ErrTooManyQueryParameters: true,
	ErrMissingBundleID:        true,
	ErrInvalidBundleReference: true,
	ErrInvalidBundleState:     true,
}

var ConflictRequestMap = map[error]bool{
	ErrBundleAlreadyExists: true,
}

var ForbiddenMap = map[error]bool{
	ErrDeleteBundleForbidden:  true,
	ErrExpectedStateOfCreated: true,
}
