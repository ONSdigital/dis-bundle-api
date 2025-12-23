package slack

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateSlackConfig(t *testing.T) {
	Convey("Given a SlackConfig", t, func() {
		test := []struct {
			name      string
			config    SlackConfig
			expectErr error
		}{
			{
				name: "missing an InfoChannel",
				config: SlackConfig{
					Channels: Channels{},
				},
				expectErr: errMissingInfoChannel,
			},
			{
				name: "missing a WarningChannel",
				config: SlackConfig{
					Channels: Channels{
						InfoChannel: "info-channel",
					},
				},
				expectErr: errMissingWarningChannel,
			},
			{
				name: "missing an AlarmChannel",
				config: SlackConfig{
					Channels: Channels{
						InfoChannel:    "info-channel",
						WarningChannel: "warning-channel",
					},
				},
				expectErr: errMissingAlarmChannel,
			},
			{
				name: "a valid config",
				config: SlackConfig{
					Channels: Channels{
						InfoChannel:    "info-channel",
						WarningChannel: "warning-channel",
						AlarmChannel:   "alarm-channel",
					},
				},
				expectErr: nil,
			},
		}

		for _, tc := range test {
			Convey("When the config is "+tc.name, func() {
				err := validateSlackConfig(&tc.config)
				So(err, ShouldEqual, tc.expectErr)
			})
		}
	})
}
