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
func New(slackConfig *SlackConfig, apiToken string, enabled bool) (Clienter, error) {
	if !enabled {
		return &NoopClient{}, nil
	}

	if slackConfig == nil {
		return nil, errNilSlackConfig
	}

	if err := validateSlackConfig(slackConfig, apiToken); err != nil {
		return nil, err
	}

	return &Client{
		client:   slack.New(apiToken),
		channels: slackConfig.Channels,
	}, nil
}

// SendAlarm sends an error notification to the configured Slack alarm channel.
func (c *Client) SendAlarm(ctx context.Context, summary string, err error, fields []Field) (*MessageRef, error) {
	return c.doSendMessage(ctx, c.channels.AlarmChannel, RedColour, AlarmEmoji, summary, buildAttachmentFields(err, fields))
}

// SendWarning sends a warning notification to the configured Slack warning channel.
func (c *Client) SendWarning(ctx context.Context, summary string, fields []Field) (*MessageRef, error) {
	return c.doSendMessage(ctx, c.channels.WarningChannel, YellowColour, WarningEmoji, summary, buildAttachmentFields(nil, fields))
}

// SendInfo sends an info notification to the configured Slack info channel.
func (c *Client) SendInfo(ctx context.Context, summary string, fields []Field) (*MessageRef, error) {
	return c.doSendMessage(ctx, c.channels.InfoChannel, GreenColour, InfoEmoji, summary, buildAttachmentFields(nil, fields))
}

// SendPublishLog sends a publish log notification to the configured Slack publish log channel.
func (c *Client) SendPublishLog(ctx context.Context, summary string, fields []Field) (*MessageRef, error) {
	return c.doSendMessage(ctx, c.channels.PublishLogChannel, YellowColour, TimerEmoji, summary, buildAttachmentFields(nil, fields))
}

// UpdatePublishLog updates a previously sent publish log message.
// The colour is set to green as we assume that a publish log is only updated when the publish is successful.
func (c *Client) UpdatePublishLog(ctx context.Context, ref *MessageRef, summary string, fields []Field) (*MessageRef, error) {
	return c.doUpdateMessage(ctx, ref, GreenColour, TickEmoji, summary, buildAttachmentFields(nil, fields))
}

// UpdatePublishLogAsAlarm updates a previously sent publish log message as a red alarm,
// used when one or more content items failed to publish.
func (c *Client) UpdatePublishLogAsAlarm(ctx context.Context, ref *MessageRef, summary string, fields []Field) (*MessageRef, error) {
	return c.doUpdateMessage(ctx, ref, RedColour, AlarmEmoji, summary, buildAttachmentFields(nil, fields))
}

// doSendMessage is a helper function to send a message to a specified Slack channel with given parameters
func (c *Client) doSendMessage(ctx context.Context, channel string, color Colour, emoji Emoji, summary string, fields []slack.AttachmentField) (*MessageRef, error) {
	attachment := slack.Attachment{
		Color:  color.String(),
		Fields: fields,
	}

	channelID, timestamp, err := c.client.PostMessageContext(
		ctx,
		channel,
		slack.MsgOptionText(fmt.Sprintf("%s %s", emoji.String(), summary), false),
		slack.MsgOptionAttachments(attachment),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send message to Slack channel %s: %w", channel, err)
	}

	return &MessageRef{ChannelID: channelID, Timestamp: timestamp}, nil
}

// doUpdateMessage updates a Slack message using its reference.
func (c *Client) doUpdateMessage(ctx context.Context, ref *MessageRef, color Colour, emoji Emoji, summary string, fields []slack.AttachmentField) (*MessageRef, error) {
	if ref == nil {
		return nil, errMissingMessageRef
	}
	if ref.ChannelID == "" {
		return nil, errMissingMessageRefChannel
	}
	if ref.Timestamp == "" {
		return nil, errMissingMessageRefTimestamp
	}

	attachment := slack.Attachment{
		Color:  color.String(),
		Fields: fields,
	}

	channelID, timestamp, _, err := c.client.UpdateMessageContext(
		ctx,
		ref.ChannelID,
		ref.Timestamp,
		slack.MsgOptionText(fmt.Sprintf("%s %s", emoji.String(), summary), false),
		slack.MsgOptionAttachments(attachment),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update message in Slack channel %s: %w", ref.ChannelID, err)
	}

	return &MessageRef{ChannelID: channelID, Timestamp: timestamp}, nil
}

// buildAttachmentFields constructs Slack attachment fields from the given error and details.
// If err is not nil, it sets the first field to the error message
func buildAttachmentFields(err error, fields []Field) []slack.AttachmentField {
	attachmentFields := []slack.AttachmentField{}

	if err != nil {
		attachmentFields = append(attachmentFields, slack.AttachmentField{
			Title: "Error",
			Value: err.Error(),
		})
	}

	for _, field := range fields {
		attachmentFields = append(attachmentFields, slack.AttachmentField{
			Title: field.Title,
			Value: field.Value,
			Short: true,
		})
	}

	return attachmentFields
}
