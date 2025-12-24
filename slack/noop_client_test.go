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
			err := client.SendAlarm(context.Background(), "Test Summary", errors.New("Test Error"), nil)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestNoopClient_SendWarning(t *testing.T) {
	Convey("Given a NoopClient", t, func() {
		client := &slack.NoopClient{}

		Convey("When SendWarning is called", func() {
			err := client.SendWarning(context.Background(), "Test Summary", nil)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestNoopClient_SendInfo(t *testing.T) {
	Convey("Given a NoopClient", t, func() {
		client := &slack.NoopClient{}

		Convey("When SendInfo is called", func() {
			err := client.SendInfo(context.Background(), "Test Summary", nil)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
