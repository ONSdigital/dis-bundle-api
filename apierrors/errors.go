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
	ErrorDescriptionMalformedRequest            = "Unable to process request due to a malformed or invalid request body or query parameter"
	ErrorDescriptionMissingParameters           = "Unable to process request due to missing required parameters in the request body or query parameters"
	ErrorDescriptionNotFound                    = "The requested resource does not exist"
	ErrorDescriptionInternalError               = "Failed to process the request due to an internal error"
	ErrorDescriptionContentItemAlreadyPublished = "Change rejected due to a conflict with the current resource state. A common cause is attempting to change a bundle that is already locked pending publication or has already been published."

	// Invalid etag
	ErrorDescriptionMissingIfMatchHeader = "Unable to process request due to missing If-Match header"
	ErrorDescriptionInvalidIfMatchHeader = "Unable to process request invalid If-Match header"

	// Invalid state
	ErrorDescriptionInvalidStateTransition = "Unable to process request due to invalid state transition"

	// Auth
	ErrorDescriptionAccessDenied                = "Access denied."
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

	ErrMissingIfMatchHeader = errors.New("missing If-Match header")
	ErrInvalidIfMatchHeader = errors.New("etag does not match")
	ErrUnableToParseTime    = errors.New("failed to parse time from json body")
	ErrScheduledAtRequired  = errors.New("scheduled_at is required for scheduled bundles")
	ErrScheduledAtSet       = errors.New("scheduled_at should not be set for manual bundles")
	ErrScheduledAtInPast    = errors.New("scheduled_at cannot be in the past")
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
	ErrVersionStateMismatched  = errors.New("version state does not match content item state")
)

// 404 Not Found
var NotFoundMap = map[error]bool{
	ErrBundleNotFound:          true,
	ErrBundleEventNotFound:     true,
	ErrBundleHasNoContentItems: true,
	ErrContentItemNotFound:     true,
	ErrContentItemNotFound:     true,
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
	ErrInvalidTransition:      true,
	ErrMissingIfMatchHeader:   true,
}

// 409 Conflict
var ConflictMap = map[error]bool{
	ErrBundleAlreadyExists:  true,
	ErrInvalidIfMatchHeader: true,
}

// 403 Forbidden
var ForbiddenMap = map[error]bool{
	ErrDeleteBundleForbidden:  true,
	ErrExpectedStateOfCreated: true,
}

var ErrMapToStatusCodeMap = map[*map[error]bool]int{
	&NotFoundMap:   404,
	&BadRequestMap: 400,
	&ConflictMap:   409,
	&ForbiddenMap:  403,
}

func GetStatusCodeForErr(err error) int {
	for errMap := range ErrMapToStatusCodeMap {
		_, exists := (*errMap)[err]

		if exists {
			return ErrMapToStatusCodeMap[errMap]
		}
	}

	return 500
}
