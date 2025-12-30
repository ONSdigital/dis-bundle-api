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
			InfoChannel:    "info-channel",
			WarningChannel: "warning-channel",
			AlarmChannel:   "alarm-channel",
		},
	}
	validAPIToken      = "valid-api-token"
	postMessageAPIPath = "/api/chat.postMessage"
)

func getMockHTTPServer(expectedPath string) *httptest.Server {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expectedPath {
			http.Error(w, "unexpected path", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
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

// This test covers the doSendMessage method indirectly through SendInfo, SendWarning, and SendAlarm.
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
			err := client.SendInfo(context.Background(), "Test Summary", map[string]interface{}{"key": "value"})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When doSendMessage is called through SendWarning", func() {
			err := client.SendWarning(context.Background(), "Test Summary", map[string]interface{}{"key": "value"})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When doSendMessage is called through SendAlarm", func() {
			err := client.SendAlarm(context.Background(), "Test Summary", errors.New("test error"), map[string]interface{}{"key": "value"})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestBuildAttachmentFields(t *testing.T) {
	Convey("Given an error and details", t, func() {
		err := errors.New("example error")
		details := map[string]interface{}{
			"key1": "value1",
		}

		Convey("When buildAttachmentFields is called", func() {
			fields := buildAttachmentFields(err, details)

			Convey("Then the returned fields contain the error and details", func() {
				So(len(fields), ShouldEqual, 2)
				So(fields[0].Title, ShouldEqual, "Error")
				So(fields[0].Value, ShouldEqual, "example error")
				So(fields[1].Title, ShouldEqual, "key1")
				So(fields[1].Value, ShouldEqual, "value1")
			})
		})
	})

	Convey("Given no error and details", t, func() {
		details := map[string]interface{}{
			"key1": 1,
		}

		Convey("When buildAttachmentFields is called", func() {
			fields := buildAttachmentFields(nil, details)

			Convey("Then the returned fields contain only the details", func() {
				So(len(fields), ShouldEqual, 1)
				So(fields[0].Title, ShouldEqual, "key1")
				So(fields[0].Value, ShouldEqual, "1")
			})
		})
	})

	Convey("Given an error and no details", t, func() {
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

	Convey("Given no error and no details", t, func() {
		Convey("When buildAttachmentFields is called", func() {
			fields := buildAttachmentFields(nil, nil)

			Convey("Then the returned fields are empty", func() {
				So(len(fields), ShouldEqual, 0)
			})
		})
	})
}
