package apierrors

import (
	"errors"
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
	// Generic Error Descriptions
	ErrorDescriptionInternalError = "Failed to process the request due to an internal error."
	ErrorDescriptionNotFound      = "The requested resource does not exist."
	ErrorDescriptionConflict      = "Change rejected due to a conflict with the current resource state. A common cause is attempting to change a bundle that is already locked pending publication or has already been published."

	// Validation Error Descriptions
	ErrorDescriptionMalformedRequest  = "Unable to process request due to a malformed or invalid request body or query parameter."
	ErrorDescriptionMissingParameters = "Unable to process request due to missing required parameters in the request body or query parameters."
	ErrorDescriptionInvalidTimeFormat = "Invalid time format in request body."

	// Bundle Error Descriptions
	ErrorDescriptionBundleTitleAlreadyExist = "A bundle with the same title already exists."

	// Bundle Contents Error Descriptions
	ErrorDescriptionVersionAlreadyExists = "This edition/version of a series already exists in another bundle."

	// State Error Descriptions
	ErrorDescriptionInvalidStateTransition      = "Unable to process request due to invalid state transition."
	ErrorDescriptionStateNotAllowedToTransition = "state not allowed to transition."

	// Header Error Descriptions
	ErrorDescriptionMissingIfMatchHeader = "Unable to process request due to missing If-Match header."
	ErrorDescriptionInvalidIfMatchHeader = "Unable to process request invalid If-Match header."

	// Auth Error Descriptions
	ErrorDescriptionAccessDenied = "Access denied."

	// Scheduling Error Descriptions
	ErrorDescriptionScheduledAtIsInPast       = "scheduled_at cannot be in the past."
	ErrorDescriptionScheduledAtShouldNotBeSet = "scheduled_at should not be set for manual bundles."
	ErrorDescriptionScheduledAtIsRequired     = "scheduled_at is required for scheduled bundles."

	// Marshal/Unmarshal Error Descriptions
	ErrorDescriptionMarshalJSONObject = "Failed to Marshal bundle resource into bytes."
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
	ErrBundleHasNoContentItems  = errors.New("bundle has no content items")

	// Content-Specific
	ErrContentItemNotFound = errors.New("content item not found")

	// Validation
	ErrMissingParameters      = errors.New("missing required parameters in request")
	ErrInvalidQueryParameter  = errors.New("invalid query parameter")
	ErrTooManyQueryParameters = errors.New("too many query parameters provided")

	// State errors
	ErrExpectedStateOfCreated  = errors.New("expected bundle state to be 'CREATED'")
	ErrExpectedStateOfApproved = errors.New("expected bundle state to be 'APPROVED'")
	ErrInvalidTransition       = errors.New("state not allowed to transition")
	ErrVersionStateNotApproved = errors.New("version state expected to be APPROVED when transitioning bundle to PUBLISHED")

	// Parsing errors
	ErrUnableToParseTime = errors.New("failed to parse time from json body")
	ErrUnableToParseJSON = errors.New("failed to parse json body")

	// Header errors
	ErrMissingIfMatchHeader = errors.New("missing If-Match header")
	ErrInvalidIfMatchHeader = errors.New("etag does not match")

	// Scheduling errors
	ErrScheduledAtRequired = errors.New("scheduled_at is required for scheduled bundles")
	ErrScheduledAtSet      = errors.New("scheduled_at should not be set for manual bundles")
	ErrScheduledAtInPast   = errors.New("scheduled_at cannot be in the past")

	// Role errors
	ErrInvalidRole = errors.New("invalid role provided")

	// Other errors
	ErrUnableToReadMessage = errors.New("failed to read message body")
)

// Map errors to HTTP status codes
var ErrorToStatusCode = map[error]int{
	ErrInvalidBody:            400,
	ErrMissingParameters:      400,
	ErrInvalidQueryParameter:  400,
	ErrTooManyQueryParameters: 400,
	ErrMissingBundleID:        400,
	ErrInvalidBundleReference: 400,
	ErrInvalidBundleState:     400,
	ErrInvalidTransition:      400,
	ErrMissingIfMatchHeader:   400,

	ErrDeleteBundleForbidden:  403,
	ErrExpectedStateOfCreated: 403,

	ErrBundleNotFound:          404,
	ErrBundleEventNotFound:     404,
	ErrBundleHasNoContentItems: 404,
	ErrContentItemNotFound:     404,

	ErrBundleAlreadyExists:  409,
	ErrInvalidIfMatchHeader: 409,
}

func GetStatusCodeForErr(err error) int {
	if code, exists := ErrorToStatusCode[err]; exists {
		return code
	}
	return 500
}
