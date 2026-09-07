package slack

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		Timeout: 30 * time.Second,
	}
	validAPIToken        = "valid-api-token"
	postMessageAPIPath   = "/api/chat.postMessage"
	updateMessageAPIPath = "/api/chat.update"

	testDetails = []Detail{
		{Title: "key 1", Value: "value 1"},
		{Title: "key 2", Value: "value 2"},
		{Title: "key 3", Value: "value 3"},
	}

	testLinks = []Link{
		{Title: "link 1", URL: "http://example.com/1"},
		{Title: "link 2", URL: "http://example.com/2"},
		{Title: "link 3", URL: "http://example.com/3"},
	}

	testTitle = "Test Title"

	testMessageRef = &MessageRef{
		ChannelID: "test-channel",
		Timestamp: "1234.5678",
	}

	testError = errors.New("test error")
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
			ref, err := client.SendInfo(context.Background(), testTitle, testDetails, testLinks)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(ref, ShouldNotBeNil)
				So(ref.ChannelID, ShouldEqual, "test-channel")
				So(ref.Timestamp, ShouldEqual, "1234.5678")
			})
		})

		Convey("When doSendMessage is called through SendWarning", func() {
			ref, err := client.SendWarning(context.Background(), testTitle, testDetails, testLinks)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(ref, ShouldNotBeNil)
				So(ref.ChannelID, ShouldEqual, "test-channel")
				So(ref.Timestamp, ShouldEqual, "1234.5678")
			})
		})

		Convey("When doSendMessage is called through SendAlarm", func() {
			ref, err := client.SendAlarm(context.Background(), testTitle, testError, testDetails, testLinks)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(ref, ShouldNotBeNil)
				So(ref.ChannelID, ShouldEqual, "test-channel")
				So(ref.Timestamp, ShouldEqual, "1234.5678")
			})
		})

		Convey("When doSendMessage is called through SendPublishLog", func() {
			ref, err := client.SendPublishLog(context.Background(), testTitle, testDetails, testLinks)

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
			ref, err := client.SendInfo(context.Background(), "Test Title", testDetails, testLinks)

			Convey("Then a wrapped error is returned", func() {
				So(ref, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to send message to Slack channel")
			})
		})
	})
}

// This test covers the doUpdateMessage method indirectly through UpdateMessage.
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

		Convey("When UpdateMessage is called", func() {
			updatedRef, err := client.UpdateMessage(context.Background(), testMessageRef, testTitle, testError, testDetails, testLinks, GreenColour, TickEmoji)

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

		Convey("When UpdateMessage is called", func() {
			updatedRef, err := client.UpdateMessage(context.Background(), testMessageRef, testTitle, testError, testDetails, testLinks, RedColour, AlarmEmoji)

			Convey("Then a wrapped error is returned", func() {
				So(updatedRef, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "failed to update message in Slack channel")
			})
		})
	})

	Convey("Given an empty Client", t, func() {
		client := &Client{}

		Convey("When UpdateMessage is called with a nil ref", func() {
			updatedRef, err := client.UpdateMessage(context.Background(), nil, testTitle, testError, testDetails, testLinks, GreenColour, TickEmoji)

			Convey("Then a missing ref error is returned", func() {
				So(err, ShouldEqual, errMissingMessageRef)
				So(updatedRef, ShouldBeNil)
			})
		})

		Convey("When UpdateMessage is called with an empty channel", func() {
			updatedRef, err := client.UpdateMessage(context.Background(), &MessageRef{Timestamp: "1234.5678"}, testTitle, testError, testDetails, testLinks, GreenColour, TickEmoji)

			Convey("Then a missing channel error is returned", func() {
				So(err, ShouldEqual, errMissingMessageRefChannel)
				So(updatedRef, ShouldBeNil)
			})
		})

		Convey("When UpdateMessage is called with an empty timestamp", func() {
			updatedRef, err := client.UpdateMessage(context.Background(), &MessageRef{ChannelID: "test-channel"}, testTitle, testError, testDetails, testLinks, GreenColour, TickEmoji)

			Convey("Then a missing timestamp error is returned", func() {
				So(err, ShouldEqual, errMissingMessageRefTimestamp)
				So(updatedRef, ShouldBeNil)
			})
		})
	})
}

func TestBuildAttachmentFieldsFromDetails(t *testing.T) {
	Convey("Given an error and details", t, func() {
		Convey("When buildAttachmentFieldsFromDetails is called", func() {
			attachmentFields := buildAttachmentFieldsFromDetails(testError, testDetails)

			Convey("Then the returned attachmentFields contain the error and details", func() {
				So(len(attachmentFields), ShouldEqual, 4)
				So(attachmentFields[0].Title, ShouldEqual, "Error")
				So(attachmentFields[0].Value, ShouldEqual, "test error")
				So(attachmentFields[1].Title, ShouldEqual, "key 1")
				So(attachmentFields[1].Value, ShouldEqual, "value 1")
				So(attachmentFields[2].Title, ShouldEqual, "key 2")
				So(attachmentFields[2].Value, ShouldEqual, "value 2")
				So(attachmentFields[3].Title, ShouldEqual, "key 3")
				So(attachmentFields[3].Value, ShouldEqual, "value 3")
			})
		})
	})

	Convey("Given no error and details", t, func() {
		Convey("When buildAttachmentFieldsFromDetails is called", func() {
			attachmentFields := buildAttachmentFieldsFromDetails(nil, testDetails)

			Convey("Then the returned attachmentFields contain only the details", func() {
				So(len(attachmentFields), ShouldEqual, 3)
				So(attachmentFields[0].Title, ShouldEqual, "key 1")
				So(attachmentFields[0].Value, ShouldEqual, "value 1")
				So(attachmentFields[1].Title, ShouldEqual, "key 2")
				So(attachmentFields[1].Value, ShouldEqual, "value 2")
				So(attachmentFields[2].Title, ShouldEqual, "key 3")
				So(attachmentFields[2].Value, ShouldEqual, "value 3")
			})
		})
	})

	Convey("Given an error and no details", t, func() {
		Convey("When buildAttachmentFieldsFromDetails is called", func() {
			attachmentFields := buildAttachmentFieldsFromDetails(testError, nil)

			Convey("Then the returned attachmentFields contain only the error", func() {
				So(len(attachmentFields), ShouldEqual, 1)
				So(attachmentFields[0].Title, ShouldEqual, "Error")
				So(attachmentFields[0].Value, ShouldEqual, "test error")
			})
		})
	})

	Convey("Given no error and no details", t, func() {
		Convey("When buildAttachmentFieldsFromDetails is called", func() {
			attachmentFields := buildAttachmentFieldsFromDetails(nil, nil)

			Convey("Then the returned attachmentFields are empty", func() {
				So(len(attachmentFields), ShouldEqual, 0)
			})
		})
	})
}

func TestBuildAttachmentFieldsFromLinks(t *testing.T) {
	Convey("Given some links", t, func() {
		Convey("When buildAttachmentFieldsFromLinks is called", func() {
			attachmentLinks := buildAttachmentFieldsFromLinks(testLinks)

			Convey("Then the returned attachmentLinks match the input links", func() {
				So(len(attachmentLinks), ShouldEqual, 3)
				So(attachmentLinks[0].Value, ShouldEqual, "<http://example.com/1|link 1>")
				So(attachmentLinks[1].Value, ShouldEqual, "<http://example.com/2|link 2>")
				So(attachmentLinks[2].Value, ShouldEqual, "<http://example.com/3|link 3>")
			})
		})
	})

	Convey("Given no links", t, func() {
		Convey("When buildAttachmentFieldsFromLinks is called", func() {
			attachmentLinks := buildAttachmentFieldsFromLinks(nil)

			Convey("Then the returned links are empty", func() {
				So(len(attachmentLinks), ShouldEqual, 0)
			})
		})
	})
}

func TestBuildAttachments(t *testing.T) {
	Convey("Given valid parameters", t, func() {
		Convey("When buildAttachments is called", func() {
			attachments := buildAttachments(testError, testDetails, testLinks, RedColour)

			Convey("Then 2 attachments are returned", func() {
				So(len(attachments), ShouldEqual, 2)
				So(attachments[0].Color, ShouldEqual, RedColour.String())
				So(attachments[1].Title, ShouldEqual, "Resources")
			})
		})

		Convey("When buildAttachments is called without links", func() {
			attachments := buildAttachments(testError, testDetails, nil, RedColour)

			Convey("Then 1 attachment is returned", func() {
				So(len(attachments), ShouldEqual, 1)
				So(attachments[0].Color, ShouldEqual, RedColour.String())
			})
		})
	})
}

func TestClient_GetTimeout(t *testing.T) {
	Convey("Given a Client with a specific timeout", t, func() {
		client := &Client{
			timeout: 30 * time.Second,
		}

		Convey("When GetTimeout is called", func() {
			timeout := client.GetTimeout()

			Convey("Then the returned timeout matches the expected value", func() {
				So(timeout, ShouldEqual, 30*time.Second)
			})
		})
	})
}
