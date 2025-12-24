package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

// Client is a wrapper around the go-slack client
type Client struct {
	client   *slack.Client
	channels Channels
}

// New returns a new Client if Slack notifications are enabled.
// If not enabled, it returns a NoopClient
func New(slackConfig *SlackConfig, apiToken string) (Clienter, error) {
	if slackConfig == nil {
		return nil, errNilSlackConfig
	}

	if slackConfig.Enabled {
		if err := validateSlackConfig(slackConfig, apiToken); err != nil {
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

// SendAlarm sends an error notification to the configured Slack alarm channel.
func (c *Client) SendAlarm(ctx context.Context, summary string, err error, details map[string]interface{}) error {
	return c.doSendMessage(ctx, c.channels.AlarmChannel, RedColour, AlarmEmoji, summary, buildAttachmentFields(err, details))
}

// SendWarning sends a warning notification to the configured Slack warning channel.
func (c *Client) SendWarning(ctx context.Context, summary string, details map[string]interface{}) error {
	return c.doSendMessage(ctx, c.channels.WarningChannel, YellowColour, WarningEmoji, summary, buildAttachmentFields(nil, details))
}

// SendInfo sends an info notification to the configured Slack info channel.
func (c *Client) SendInfo(ctx context.Context, summary string, details map[string]interface{}) error {
	return c.doSendMessage(ctx, c.channels.InfoChannel, GreenColour, InfoEmoji, summary, buildAttachmentFields(nil, details))
}

// doSendMessage is a helper function to send a message to a specified Slack channel with given parameters
func (c *Client) doSendMessage(ctx context.Context, channel string, color Colour, emoji Emoji, summary string, fields []slack.AttachmentField) error {
	attachment := slack.Attachment{
		Color:  color.String(),
		Fields: fields,
	}

	_, _, err := c.client.PostMessageContext(
		ctx,
		channel,
		slack.MsgOptionText(fmt.Sprintf("%s %s", emoji.String(), summary), false),
		slack.MsgOptionAttachments(attachment),
	)

	return err
}

// buildAttachmentFields constructs Slack attachment fields from the given error and details.
// If err is not nil, it sets the first field to the error message
func buildAttachmentFields(err error, details map[string]interface{}) []slack.AttachmentField {
	fields := []slack.AttachmentField{}

	if err != nil {
		fields = append(fields, slack.AttachmentField{
			Title: "Error",
			Value: err.Error(),
		})
	}

	for key, value := range details {
		fields = append(fields, slack.AttachmentField{
			Title: key,
			Value: fmt.Sprintf("%v", value),
		})
	}

	return fields
}
