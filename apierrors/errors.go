package apierrors

import "errors"

// Custom error types for common HTTP cases
type ErrInvalidPatch struct {
	Msg string
}

func (e ErrInvalidPatch) Error() string {
	return e.Msg
}

// Response error descriptions
var (
	ErrorDescriptionMalformedRequest            = "Unable to process request due to a malformed or invalid request body or query parameter"
	ErrorDescriptionMissingParameters           = "Unable to process request due to missing required parameters in the request body or query parameters"
	ErrorDescriptionNotFound                    = "The requested resource does not exist"
	ErrorDescriptionInternalError               = "Failed to process the request due to an internal error"
	ErrorDescriptionAlreadyPublished            = "Change rejected due to a conflict with the current resource state. A common cause is attempting to change a bundle that is already locked pending publication or has already been published."
	ErrorDescriptionInvalidTimeFormat           = "Invalid time format in request body"
	ErrorDescriptionScheduledAtIsInPast         = "scheduled_at cannot be in the past"
	ErrorDescriptionScheduledAtShouldNotBeSet   = "scheduled_at should not be set for manual bundles"
	ErrorDescriptionScheduledAtIsRequired       = "scheduled_at is required for scheduled bundles"
	ErrorDescriptionBundleTitleAlreadyExist     = "A bundle with the same title already exists"
	ErrorDescriptionStateNotAllowedToTransition = "state not allowed to transition"
)

var (
	ErrUnableToReadMessage = errors.New("failed to read message body")
	ErrUnableToParseJSON   = errors.New("failed to parse json body")
	ErrUnableToParseTime   = errors.New("failed to parse time from json body")
	ErrScheduledAtRequired = errors.New("scheduled_at is required for scheduled bundles")
	ErrScheduledAtSet      = errors.New("scheduled_at should not be set for manual bundles")
	ErrScheduledAtInPast   = errors.New("scheduled_at cannot be in the past")
)

var (
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
	ErrBundleNotFound           = errors.New("bundle not found")
	ErrDeleteBundleForbidden    = errors.New("cannot delete a published bundle")
	ErrBundleAlreadyExists      = errors.New("bundle already exists")
	ErrBundleTitleAlreadyExists = errors.New("bundle with the same title already exists")
	ErrInvalidBundleState       = errors.New("invalid bundle state")
	ErrMissingBundleID          = errors.New("missing bundle ID")
	ErrInvalidBundleReference   = errors.New("invalid bundle reference")
	ErrBundleEventNotFound      = errors.New("bundle event not found")

	// Content-Specific
	ErrContentItemNotFound = errors.New("content item not found")

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
	ErrContentItemNotFound: true,
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
