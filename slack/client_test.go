package slack

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slack-go/slack"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	validSlackConfig = &SlackConfig{
		Channels: Channels{
			InfoChannel:       "info-channel",
			WarningChannel:    "warning-channel",
			AlarmChannel:      "alarm-channel",
			PublishLogChannel: "publish-log-channel",
		},
	}
	validAPIToken        = "valid-api-token"
	postMessageAPIPath   = "/api/chat.postMessage"
	updateMessageAPIPath = "/api/chat.update"
)

func getMockHTTPServer(expectedPath string) *httptest.Server {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expectedPath {
			http.Error(w, "unexpected path", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true, "channel": "test-channel", "ts": "1234.5678"}`))
	}))
	return testServer
}

func TestNew(t *testing.T) {
	Convey("Given a SlackConfig, apiToken and enabled set to true", t, func() {
		config := validSlackConfig

		Convey("When New is called", func() {
			client, err := New(config, validAPIToken, true)
			So(err, ShouldBeNil)

			Convey("Then a Client is returned", func() {
				_, ok := client.(*Client)
				So(ok, ShouldBeTrue)
			})
		})
	})

	Convey("Given a SlackConfig, apiToken and enabled set to false", t, func() {
		config := &SlackConfig{}

		Convey("When New is called", func() {
			client, err := New(config, validAPIToken, false)
			So(err, ShouldBeNil)

			Convey("Then a NoopClient is returned", func() {
				_, ok := client.(*NoopClient)
				So(ok, ShouldBeTrue)
			})
		})
	})

	Convey("Given a nil SlackConfig, apiToken and enabled set to true", t, func() {
		var config *SlackConfig = nil

		Convey("When New is called", func() {
			_, err := New(config, validAPIToken, true)

			Convey("Then an error is returned", func() {
				So(err, ShouldEqual, errNilSlackConfig)
			})
		})
	})

	Convey("Given an invalid SlackConfig, apiToken and enabled set to true", t, func() {
		config := &SlackConfig{
			Channels: Channels{
				InfoChannel:    "",
				WarningChannel: "warning-channel",
				AlarmChannel:   "alarm-channel",
			},
		}

		Convey("When New is called", func() {
			_, err := New(config, validAPIToken, true)

			Convey("Then an error is returned", func() {
				So(err.Error(), ShouldEqual, "slack info channel is missing")
			})
		})
	})
}

// This test covers the doSendMessage method indirectly through SendInfo, SendWarning, SendAlarm and SendPublishLog.
// It verifies that messages can be sent to Slack without errors and uses a mock HTTP server to simulate Slack's API.
func TestClient_DoSendMessage(t *testing.T) {
	Convey("Given a mock Slack Client and valid parameters", t, func() {
		testServer := getMockHTTPServer(postMessageAPIPath)
		defer testServer.Close()

		slackClient := slack.New(
			validAPIToken,
			slack.OptionHTTPClient(testServer.Client()),
			slack.OptionAPIURL(testServer.URL+"/api/"),
		)

		client := &Client{
			client:   slackClient,
			channels: validSlackConfig.Channels,
		}

		Convey("When doSendMessage is called through SendInfo", func() {
			ref, err := client.SendInfo(context.Background(), "Test Summary", []Field{{Title: "key", Value: "value"}})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(ref, ShouldNotBeNil)
				So(ref.ChannelID, ShouldEqual, "test-channel")
				So(ref.Timestamp, ShouldEqual, "1234.5678")
			})
		})

		Convey("When doSendMessage is called through SendWarning", func() {
			ref, err := client.SendWarning(context.Background(), "Test Summary", []Field{{Title: "key", Value: "value"}})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(ref, ShouldNotBeNil)
				So(ref.ChannelID, ShouldEqual, "test-channel")
				So(ref.Timestamp, ShouldEqual, "1234.5678")
			})
		})

		Convey("When doSendMessage is called through SendAlarm", func() {
			ref, err := client.SendAlarm(context.Background(), "Test Summary", errors.New("test error"), []Field{{Title: "key", Value: "value"}})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(ref, ShouldNotBeNil)
				So(ref.ChannelID, ShouldEqual, "test-channel")
				So(ref.Timestamp, ShouldEqual, "1234.5678")
			})
		})

		Convey("When doSendMessage is called through SendPublishLog", func() {
			ref, err := client.SendPublishLog(context.Background(), "Test Summary", []Field{{Title: "key", Value: "value"}})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(ref, ShouldNotBeNil)
				So(ref.ChannelID, ShouldEqual, "test-channel")
				So(ref.Timestamp, ShouldEqual, "1234.5678")
			})
		})
	})

	Convey("Given a Slack Client that returns an error", t, func() {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"ok": false, "error": "server_error"}`))
		}))
		defer testServer.Close()

		slackClient := slack.New(
			validAPIToken,
			slack.OptionHTTPClient(testServer.Client()),
			slack.OptionAPIURL(testServer.URL+"/api/"),
		)

		client := &Client{
			client:   slackClient,
			channels: validSlackConfig.Channels,
		}

		Convey("When SendInfo is called", func() {
			ref, err := client.SendInfo(context.Background(), "Test Summary", nil)

			Convey("Then a wrapped error is returned", func() {
				So(ref, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to send message to Slack channel")
			})
		})
	})
}

// This test covers the doUpdateMessage method indirectly through UpdatePublishLog and UpdatePublishLogAsAlarm.
// It verifies that message updates can be sent to Slack without errors and uses a mock HTTP server to simulate Slack's API.
func TestClient_DoUpdateMessage(t *testing.T) {
	Convey("Given a mock Slack Client and valid parameters", t, func() {
		testServer := getMockHTTPServer(updateMessageAPIPath)
		defer testServer.Close()

		slackClient := slack.New(
			validAPIToken,
			slack.OptionHTTPClient(testServer.Client()),
			slack.OptionAPIURL(testServer.URL+"/api/"),
		)

		client := &Client{
			client:   slackClient,
			channels: validSlackConfig.Channels,
		}

		ref := &MessageRef{ChannelID: "test-channel", Timestamp: "1234.5678"}

		Convey("When doUpdateMessage is called through UpdatePublishLog", func() {
			updatedRef, err := client.UpdatePublishLog(context.Background(), ref, "Updated Summary", []Field{{Title: "key", Value: "value"}})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(updatedRef, ShouldNotBeNil)
				So(updatedRef.ChannelID, ShouldEqual, "test-channel")
				So(updatedRef.Timestamp, ShouldEqual, "1234.5678")
			})
		})

		Convey("When doUpdateMessage is called through UpdatePublishLogAsAlarm", func() {
			updatedRef, err := client.UpdatePublishLogAsAlarm(context.Background(), ref, "Updated Summary", []Field{{Title: "key", Value: "value"}})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(updatedRef, ShouldNotBeNil)
				So(updatedRef.ChannelID, ShouldEqual, "test-channel")
				So(updatedRef.Timestamp, ShouldEqual, "1234.5678")
			})
		})
	})

	Convey("Given a Slack Client that returns an error", t, func() {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"ok": false, "error": "server_error"}`))
		}))
		defer testServer.Close()

		slackClient := slack.New(
			validAPIToken,
			slack.OptionHTTPClient(testServer.Client()),
			slack.OptionAPIURL(testServer.URL+"/api/"),
		)

		client := &Client{
			client:   slackClient,
			channels: validSlackConfig.Channels,
		}

		ref := &MessageRef{ChannelID: "test-channel", Timestamp: "1234.5678"}

		Convey("When UpdatePublishLog is called", func() {
			updatedRef, err := client.UpdatePublishLog(context.Background(), ref, "Updated Summary", nil)

			Convey("Then a wrapped error is returned", func() {
				So(updatedRef, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to update message in Slack channel")
			})
		})

		Convey("When UpdatePublishLogAsAlarm is called", func() {
			updatedRef, err := client.UpdatePublishLogAsAlarm(context.Background(), ref, "Updated Summary", nil)

			Convey("Then a wrapped error is returned", func() {
				So(updatedRef, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to update message in Slack channel")
			})
		})
	})

	Convey("Given invalid message references", t, func() {
		client := &Client{}

		Convey("When UpdatePublishLog is called with a nil ref", func() {
			updatedRef, err := client.UpdatePublishLog(context.Background(), nil, "Updated Summary", nil)

			Convey("Then a missing ref error is returned", func() {
				So(err, ShouldEqual, errMissingMessageRef)
				So(updatedRef, ShouldBeNil)
			})
		})

		Convey("When UpdatePublishLog is called with an empty channel", func() {
			updatedRef, err := client.UpdatePublishLog(context.Background(), &MessageRef{Timestamp: "1234.5678"}, "Updated Summary", nil)

			Convey("Then a missing channel error is returned", func() {
				So(err, ShouldEqual, errMissingMessageRefChannel)
				So(updatedRef, ShouldBeNil)
			})
		})

		Convey("When UpdatePublishLog is called with an empty timestamp", func() {
			updatedRef, err := client.UpdatePublishLog(context.Background(), &MessageRef{ChannelID: "test-channel"}, "Updated Summary", nil)

			Convey("Then a missing timestamp error is returned", func() {
				So(err, ShouldEqual, errMissingMessageRefTimestamp)
				So(updatedRef, ShouldBeNil)
			})
		})

		Convey("When UpdatePublishLogAsAlarm is called with a nil ref", func() {
			updatedRef, err := client.UpdatePublishLogAsAlarm(context.Background(), nil, "Updated Summary", nil)

			Convey("Then a missing ref error is returned", func() {
				So(err, ShouldEqual, errMissingMessageRef)
				So(updatedRef, ShouldBeNil)
			})
		})

		Convey("When UpdatePublishLogAsAlarm is called with an empty channel", func() {
			updatedRef, err := client.UpdatePublishLogAsAlarm(context.Background(), &MessageRef{Timestamp: "1234.5678"}, "Updated Summary", nil)

			Convey("Then a missing channel error is returned", func() {
				So(err, ShouldEqual, errMissingMessageRefChannel)
				So(updatedRef, ShouldBeNil)
			})
		})

		Convey("When UpdatePublishLogAsAlarm is called with an empty timestamp", func() {
			updatedRef, err := client.UpdatePublishLogAsAlarm(context.Background(), &MessageRef{ChannelID: "test-channel"}, "Updated Summary", nil)

			Convey("Then a missing timestamp error is returned", func() {
				So(err, ShouldEqual, errMissingMessageRefTimestamp)
				So(updatedRef, ShouldBeNil)
			})
		})
	})
}

func TestBuildAttachmentFields(t *testing.T) {
	Convey("Given an error and fields", t, func() {
		err := errors.New("example error")
		fields := []Field{
			{Title: "key1", Value: "value1"},
		}

		Convey("When buildAttachmentFields is called", func() {
			fields := buildAttachmentFields(err, fields)

			Convey("Then the returned fields contain the error and details", func() {
				So(len(fields), ShouldEqual, 2)
				So(fields[0].Title, ShouldEqual, "Error")
				So(fields[0].Value, ShouldEqual, "example error")
				So(fields[1].Title, ShouldEqual, "key1")
				So(fields[1].Value, ShouldEqual, "value1")
			})
		})
	})

	Convey("Given no error and fields", t, func() {
		fields := []Field{
			{Title: "key1", Value: "1"},
		}

		Convey("When buildAttachmentFields is called", func() {
			fields := buildAttachmentFields(nil, fields)

			Convey("Then the returned fields contain only the details", func() {
				So(len(fields), ShouldEqual, 1)
				So(fields[0].Title, ShouldEqual, "key1")
				So(fields[0].Value, ShouldEqual, "1")
			})
		})
	})

	Convey("Given an error and no fields", t, func() {
		err := errors.New("example error")

		Convey("When buildAttachmentFields is called", func() {
			fields := buildAttachmentFields(err, nil)

			Convey("Then the returned fields contain only the error", func() {
				So(len(fields), ShouldEqual, 1)
				So(fields[0].Title, ShouldEqual, "Error")
				So(fields[0].Value, ShouldEqual, "example error")
			})
		})
	})

	Convey("Given no error and no fields", t, func() {
		Convey("When buildAttachmentFields is called", func() {
			fields := buildAttachmentFields(nil, nil)

			Convey("Then the returned fields are empty", func() {
				So(len(fields), ShouldEqual, 0)
			})
		})
	})
}
