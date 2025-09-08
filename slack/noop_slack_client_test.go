package slack_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/slack"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNoopNotifier_SendError(t *testing.T) {
	Convey("Given a NoopNotifier", t, func() {
		notifier := &slack.NoopNotifier{}

		Convey("When SendError is called", func() {
			err := notifier.SendError(context.Background(), "Test Summary", errors.New("Test Error"), nil)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
