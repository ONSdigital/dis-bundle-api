package slack

import "errors"

// Predefined errors used within the slack package
var (
	// configuration errors
	errMissingAPIToken       = errors.New("slack API token is missing")
	errMissingInfoChannel    = errors.New("slack info channel is missing")
	errMissingWarningChannel = errors.New("slack warning channel is missing")
	errMissingAlarmChannel   = errors.New("slack alarm channel is missing")
)
