package slack_test

import (
	"testing"

	"github.com/ONSdigital/dis-bundle-api/slack"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	validSlackConfig = &slack.SlackConfig{
		Channels: slack.Channels{
			InfoChannel:    "info-channel",
			WarningChannel: "warning-channel",
			AlarmChannel:   "alarm-channel",
		},
		Enabled: true,
	}
	validAPIToken = "valid-api-token"
)

func TestNew(t *testing.T) {
	Convey("Given a SlackConfig with Enabled set to true", t, func() {
		config := validSlackConfig

		Convey("When New is called", func() {
			client, err := slack.New(config, validAPIToken)
			So(err, ShouldBeNil)

			Convey("Then a Client is returned", func() {
				_, ok := client.(*slack.Client)
				So(ok, ShouldBeTrue)
			})
		})
	})

	Convey("Given a SlackConfig with Enabled set to false", t, func() {
		config := &slack.SlackConfig{
			Enabled: false,
		}

		Convey("When New is called", func() {
			client, err := slack.New(config, validAPIToken)
			So(err, ShouldBeNil)

			Convey("Then a NoopClient is returned", func() {
				_, ok := client.(*slack.NoopClient)
				So(ok, ShouldBeTrue)
			})
		})
	})

	Convey("Given a SlackConfig with a missing API token", t, func() {
		config := &slack.SlackConfig{
			Enabled: true,
		}

		Convey("When New is called", func() {
			_, err := slack.New(config, "")

			Convey("Then an error is returned", func() {
				So(err.Error(), ShouldEqual, "slack API token is missing")
			})
		})
	})

	Convey("Given a SlackConfig with invalid configuration", t, func() {
		config := &slack.SlackConfig{
			Enabled: true,
			Channels: slack.Channels{
				InfoChannel:    "",
				WarningChannel: "warning-channel",
				AlarmChannel:   "alarm-channel",
			},
		}

		Convey("When New is called", func() {
			_, err := slack.New(config, validAPIToken)

			Convey("Then an error is returned", func() {
				So(err.Error(), ShouldEqual, "slack info channel is missing")
			})
		})
	})
}

func TestClient_Channels(t *testing.T) {
	Convey("Given a Client", t, func() {
		client, err := slack.New(validSlackConfig, validAPIToken)
		So(err, ShouldBeNil)

		Convey("When Channels is called", func() {
			returnedChannels := client.Channels()
			Convey("Then the correct Channels are returned", func() {
				So(returnedChannels.InfoChannel, ShouldEqual, validSlackConfig.Channels.InfoChannel)
				So(returnedChannels.WarningChannel, ShouldEqual, validSlackConfig.Channels.WarningChannel)
				So(returnedChannels.AlarmChannel, ShouldEqual, validSlackConfig.Channels.AlarmChannel)
			})
		})
	})
}
