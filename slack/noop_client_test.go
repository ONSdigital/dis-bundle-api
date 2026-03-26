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

		Convey("When SendAlarm is called", func() {
			ref, err := client.SendAlarm(context.Background(), "Test Summary", errors.New("Test Error"), nil)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(ref, ShouldBeNil)
			})
		})
	})
}

func TestNoopClient_SendWarning(t *testing.T) {
	Convey("Given a NoopClient", t, func() {
		client := &slack.NoopClient{}

		Convey("When SendWarning is called", func() {
			ref, err := client.SendWarning(context.Background(), "Test Summary", nil)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(ref, ShouldBeNil)
			})
		})
	})
}

func TestNoopClient_SendInfo(t *testing.T) {
	Convey("Given a NoopClient", t, func() {
		client := &slack.NoopClient{}

		Convey("When SendInfo is called", func() {
			ref, err := client.SendInfo(context.Background(), "Test Summary", nil)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(ref, ShouldBeNil)
			})
		})
	})
}

func TestNoopClient_SendPublishLog(t *testing.T) {
	Convey("Given a NoopClient", t, func() {
		client := &slack.NoopClient{}

		Convey("When SendPublishLog is called", func() {
			ref, err := client.SendPublishLog(context.Background(), "Test Summary", nil)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(ref, ShouldBeNil)
			})
		})
	})
}

func TestNoopClient_UpdatePublishLog(t *testing.T) {
	Convey("Given a NoopClient", t, func() {
		client := &slack.NoopClient{}

		Convey("When UpdatePublishLog is called", func() {
			ref, err := client.UpdatePublishLog(context.Background(), &slack.MessageRef{ChannelID: "test-channel", Timestamp: "1234567890.123456"}, "Test Summary", nil)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(ref, ShouldBeNil)
			})
		})
	})
}
