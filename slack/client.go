package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

// Client is a wrapper around the go-slack client
type Client struct {
	client   SlackClient
	channels Channels
}

// New returns a new Client if Slack notifications are enabled
// If not enabled, it returns a NoopClient
func New(slackConfig *SlackConfig, apiToken string) (Clienter, error) {
	if slackConfig.Enabled {
		if apiToken == "" {
			return nil, errMissingAPIToken
		}

		if err := validateSlackConfig(slackConfig); err != nil {
			return nil, err
		}

		client := slack.New(apiToken)

		return &Client{
			client:   client,
			channels: slackConfig.Channels,
		}, nil
	}
	return &NoopClient{}, nil
}

// SendError formats an error message with a summary, error, and optional details, then sends it to the provided Slack channel
func (c *Client) SendError(ctx context.Context, channel, summary string, err error, details map[string]interface{}) error {
	fields := []slack.AttachmentField{
		{Title: "Error", Value: err.Error(), Short: false},
	}

	for key, value := range details {
		fields = append(fields, slack.AttachmentField{
			Title: key,
			Value: fmt.Sprintf("%v", value),
			Short: false,
		})
	}

	attachment := slack.Attachment{
		Color:  "danger",
		Fields: fields,
	}

	_, _, sendErr := c.client.PostMessage(
		channel,
		slack.MsgOptionText(fmt.Sprintf("%s %s", AlarmEmoji.String(), summary), false),
		slack.MsgOptionAttachments(attachment),
	)

	if sendErr != nil {
		return sendErr
	}

	return nil
}

// Channels returns the configured Slack channels for the client
func (c *Client) Channels() Channels {
	return c.channels
}
