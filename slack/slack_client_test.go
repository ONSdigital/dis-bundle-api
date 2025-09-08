package slack_test

import (
	"context"
	"errors"
	"testing"

	goslack "github.com/slack-go/slack"

	"github.com/ONSdigital/dis-bundle-api/slack"
	"github.com/ONSdigital/dis-bundle-api/slack/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewSlackNotifier(t *testing.T) {
	Convey("Given a SlackConfig with Enabled set to true", t, func() {
		config := &slack.SlackConfig{
			APIToken: "test-token",
			Username: "test-username",
			Channel:  "test-channel",
			Emoji:    ":test-emoji:",
			Enabled:  true,
		}

		Convey("When NewSlackNotifier is called", func() {
			notifier := slack.NewSlackNotifier(config)

			Convey("Then a SlackNotifier is returned", func() {
				_, ok := notifier.(*slack.SlackNotifier)
				So(ok, ShouldBeTrue)
			})
		})
	})

	Convey("Given a SlackConfig with Enabled set to false", t, func() {
		config := &slack.SlackConfig{
			APIToken: "test-token",
			Username: "test-username",
			Channel:  "test-channel",
			Emoji:    ":test-emoji:",
			Enabled:  false,
		}

		Convey("When NewSlackNotifier is called", func() {
			notifier := slack.NewSlackNotifier(config)

			Convey("Then a NoopNotifier is returned", func() {
				_, ok := notifier.(*slack.NoopNotifier)
				So(ok, ShouldBeTrue)
			})
		})
	})
}

func TestSlackNotifier_SendError(t *testing.T) {
	Convey("Given a SlackNotifier with a mock SlackClient", t, func() {
		mockSlackClient := mocks.SlackClientMock{
			PostMessageFunc: func(channelID string, options ...goslack.MsgOption) (string, string, error) {
				return "", "", nil
			},
		}
		notifier := slack.SlackNotifier{
			Client:   &mockSlackClient,
			Username: "test-username",
			Channel:  "test-channel",
			Emoji:    ":test-emoji:",
		}

		Convey("When SendError is called with a summary, error, and details", func() {
			err := notifier.SendError(context.Background(), "Test Summary", errors.New("Test Error"), map[string]interface{}{"key1": "value1", "key2": 2})

			Convey("Then no error is returned and PostMessage is called on the SlackClient", func() {
				So(err, ShouldBeNil)
				So(mockSlackClient.PostMessageCalls(), ShouldHaveLength, 1)

				call := mockSlackClient.PostMessageCalls()[0]
				So(call.ChannelID, ShouldEqual, "test-channel")
				So(len(call.Options), ShouldEqual, 4)
			})
		})

		Convey("When SendError is called with nil details", func() {
			err := notifier.SendError(context.Background(), "Test Summary", errors.New("Test Error"), nil)

			Convey("Then no error is returned and PostMessage is called on the SlackClient", func() {
				So(err, ShouldBeNil)
				So(mockSlackClient.PostMessageCalls(), ShouldHaveLength, 1)

				call := mockSlackClient.PostMessageCalls()[0]
				So(call.ChannelID, ShouldEqual, "test-channel")
				So(len(call.Options), ShouldEqual, 4)
			})
		})

		Convey("When SendError is called and PostMessage returns an error", func() {
			mockSlackClient.PostMessageFunc = func(channelID string, options ...goslack.MsgOption) (string, string, error) {
				return "", "", errors.New("PostMessage Error")
			}

			err := notifier.SendError(context.Background(), "Test Summary", errors.New("Test Error"), map[string]interface{}{"key1": "value1", "key2": 2})

			Convey("Then the error from PostMessage is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "PostMessage Error")
				So(mockSlackClient.PostMessageCalls(), ShouldHaveLength, 1)
			})
		})
	})
}
