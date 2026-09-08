package slack

import (
	"context"
	"fmt"
	"time"

	"github.com/slack-go/slack"
)

// Client is a wrapper around the go-slack client
type Client struct {
	client   *slack.Client
	channels Channels
	timeout  time.Duration
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
		timeout:  slackConfig.Timeout,
	}, nil
}

// SendAlarm sends an error notification to the configured Slack alarm channel.
func (c *Client) SendAlarm(ctx context.Context, title string, err error, details []Detail, links []Link) (*MessageRef, error) {
	return c.doSendMessage(ctx, c.channels.AlarmChannel, title, AlarmEmoji, buildAttachments(err, details, links, RedColour))
}

// SendWarning sends a warning notification to the configured Slack warning channel.
func (c *Client) SendWarning(ctx context.Context, title string, details []Detail, links []Link) (*MessageRef, error) {
	return c.doSendMessage(ctx, c.channels.WarningChannel, title, WarningEmoji, buildAttachments(nil, details, links, YellowColour))
}

// SendInfo sends an info notification to the configured Slack info channel.
func (c *Client) SendInfo(ctx context.Context, title string, details []Detail, links []Link) (*MessageRef, error) {
	return c.doSendMessage(ctx, c.channels.InfoChannel, title, InfoEmoji, buildAttachments(nil, details, links, GreenColour))
}

// SendPublishLog sends a publish log notification to the configured Slack publish log channel.
func (c *Client) SendPublishLog(ctx context.Context, title string, details []Detail, links []Link) (*MessageRef, error) {
	return c.doSendMessage(ctx, c.channels.PublishLogChannel, title, TimerEmoji, buildAttachments(nil, details, links, YellowColour))
}

// UpdateMessage updates a previously sent Slack message with the given title, error, details, links, color, and emoji.
func (c *Client) UpdateMessage(ctx context.Context, ref *MessageRef, title string, err error, details []Detail, links []Link, color Colour, emoji Emoji) (*MessageRef, error) {
	return c.doUpdateMessage(ctx, ref, title, emoji, buildAttachments(err, details, links, color))
}

// doSendMessage is a helper function to send a message to a specified Slack channel with given parameters
func (c *Client) doSendMessage(ctx context.Context, channel, title string, emoji Emoji, attachments []slack.Attachment) (*MessageRef, error) {
	channelID, timestamp, err := c.client.PostMessageContext(
		ctx,
		channel,
		slack.MsgOptionText(fmt.Sprintf("%s *%s*", emoji.String(), title), false),
		slack.MsgOptionAttachments(attachments...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send message to Slack channel %s: %w", channel, err)
	}

	return &MessageRef{ChannelID: channelID, Timestamp: timestamp}, nil
}

// doUpdateMessage updates a Slack message using its reference.
func (c *Client) doUpdateMessage(ctx context.Context, ref *MessageRef, title string, emoji Emoji, attachments []slack.Attachment) (*MessageRef, error) {
	if ref == nil {
		return nil, errMissingMessageRef
	}
	if ref.ChannelID == "" {
		return nil, errMissingMessageRefChannel
	}
	if ref.Timestamp == "" {
		return nil, errMissingMessageRefTimestamp
	}

	channelID, timestamp, _, err := c.client.UpdateMessageContext(
		ctx,
		ref.ChannelID,
		ref.Timestamp,
		slack.MsgOptionText(fmt.Sprintf("%s *%s*", emoji.String(), title), false),
		slack.MsgOptionAttachments(attachments...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update message in Slack channel %s: %w", ref.ChannelID, err)
	}

	return &MessageRef{ChannelID: channelID, Timestamp: timestamp}, nil
}

// buildAttachmentFieldsFromDetails constructs Slack attachment fields from the given error and details.
func buildAttachmentFieldsFromDetails(err error, details []Detail) []slack.AttachmentField {
	attachmentFields := make([]slack.AttachmentField, 0, len(details)+1)

	if err != nil {
		attachmentFields = append(attachmentFields, slack.AttachmentField{
			Title: "Error",
			Value: err.Error(),
		})
	}

	for _, detail := range details {
		attachmentFields = append(attachmentFields, slack.AttachmentField{
			Title: detail.Title,
			Value: detail.Value,
			Short: true,
		})
	}

	return attachmentFields
}

// buildAttachmentFieldsFromLinks constructs Slack attachment fields for the given links.
func buildAttachmentFieldsFromLinks(links []Link) []slack.AttachmentField {
	attachmentLinks := make([]slack.AttachmentField, 0, len(links))

	for _, link := range links {
		attachmentLinks = append(attachmentLinks, slack.AttachmentField{
			Value: fmt.Sprintf("<%s|%s>", link.URL, link.Title),
		})
	}

	return attachmentLinks
}

// buildAttachments constructs Slack attachments for a message, including the details, links, and error if present.
func buildAttachments(err error, details []Detail, links []Link, color Colour) []slack.Attachment {
	attachments := make([]slack.Attachment, 0, 2)

	// Message details
	attachments = append(attachments, slack.Attachment{
		Fields: buildAttachmentFieldsFromDetails(err, details),
		Color:  color.String(),
	})

	// Message links
	if len(links) > 0 {
		attachments = append(attachments, slack.Attachment{
			Title:  "Resources",
			Fields: buildAttachmentFieldsFromLinks(links),
		})
	}

	return attachments
}

// GetTimeout returns the timeout duration for Slack API requests.
func (c *Client) GetTimeout() time.Duration {
	return c.timeout
}
