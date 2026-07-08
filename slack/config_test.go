package slack

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateSlackConfig(t *testing.T) {
	Convey("Given a SlackConfig", t, func() {
		test := []struct {
			name      string
			apiToken  string
			config    *SlackConfig
			expectErr error
		}{
			{
				name:      "missing API token",
				apiToken:  "",
				config:    &SlackConfig{},
				expectErr: errMissingAPIToken,
			},
			{
				name:     "missing an InfoChannel",
				apiToken: validAPIToken,
				config: &SlackConfig{
					Channels: Channels{},
				},

				expectErr: errMissingInfoChannel,
			},
			{
				name:     "missing a WarningChannel",
				apiToken: validAPIToken,
				config: &SlackConfig{
					Channels: Channels{
						InfoChannel: "info-channel",
					},
				},
				expectErr: errMissingWarningChannel,
			},
			{
				name:     "missing an AlarmChannel",
				apiToken: validAPIToken,
				config: &SlackConfig{
					Channels: Channels{
						InfoChannel:    "info-channel",
						WarningChannel: "warning-channel",
					},
				},
				expectErr: errMissingAlarmChannel,
			},
			{
				name:     "missing a PublishLogChannel",
				apiToken: validAPIToken,
				config: &SlackConfig{
					Channels: Channels{
						InfoChannel:    "info-channel",
						WarningChannel: "warning-channel",
						AlarmChannel:   "alarm-channel",
					},
				},
				expectErr: errMissingPublishLogChannel,
			},
			{
				name:     "invalid timeout",
				apiToken: validAPIToken,
				config: &SlackConfig{
					Channels: Channels{
						InfoChannel:       "info-channel",
						WarningChannel:    "warning-channel",
						AlarmChannel:      "alarm-channel",
						PublishLogChannel: "publish-log-channel",
					},
					Timeout: 0,
				},
				expectErr: errInvalidTimeout,
			},
			{
				name:     "a valid config",
				apiToken: validAPIToken,
				config: &SlackConfig{
					Channels: Channels{
						InfoChannel:       "info-channel",
						WarningChannel:    "warning-channel",
						AlarmChannel:      "alarm-channel",
						PublishLogChannel: "publish-log-channel",
					},
					Timeout: 30 * time.Second,
				},
				expectErr: nil,
			},
		}

		for _, tc := range test {
			Convey("When the config is "+tc.name, func() {
				err := validateSlackConfig(tc.config, tc.apiToken)
				So(err, ShouldEqual, tc.expectErr)
			})
		}
	})
}
