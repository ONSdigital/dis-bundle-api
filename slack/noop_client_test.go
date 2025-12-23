package slack_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/slack"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNoopClient_SendError(t *testing.T) {
	Convey("Given a NoopClient", t, func() {
		client := &slack.NoopClient{}

		Convey("When SendError is called", func() {
			err := client.SendError(context.Background(), "Test Channel", "Test Summary", errors.New("Test Error"), nil)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestNoopClient_Channels(t *testing.T) {
	Convey("Given a NoopClient", t, func() {
		client := &slack.NoopClient{}

		Convey("When Channels is called", func() {
			channels := client.Channels()

			Convey("Then empty channels are returned", func() {
				So(channels.InfoChannel, ShouldBeEmpty)
				So(channels.WarningChannel, ShouldBeEmpty)
				So(channels.AlarmChannel, ShouldBeEmpty)
			})
		})
	})
}
