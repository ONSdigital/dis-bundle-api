package slack

import "errors"

// Predefined errors used within the slack package
var (
	// configuration errors
	errNilSlackConfig           = errors.New("slack configuration is nil")
	errMissingAPIToken          = errors.New("slack API token is missing")
	errMissingInfoChannel       = errors.New("slack info channel is missing")
	errMissingWarningChannel    = errors.New("slack warning channel is missing")
	errMissingAlarmChannel      = errors.New("slack alarm channel is missing")
	errMissingPublishLogChannel = errors.New("slack publish log channel is missing")

	// MessageRef errors
	errMissingMessageRef          = errors.New("slack message reference is missing")
	errMissingMessageRefChannel   = errors.New("slack message reference channel is missing")
	errMissingMessageRefTimestamp = errors.New("slack message reference timestamp is missing")
)
