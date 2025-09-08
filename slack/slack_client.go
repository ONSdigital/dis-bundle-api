package slack

import (
	"context"
	"fmt"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/slack-go/slack"
)

// SlackNotifier is a Notifier that uses slack-go/slack.Client to send notifications
type SlackNotifier struct {
	Client   SlackClient
	Username string
	Channel  string
	Emoji    string
}

// NewSlackNotifier creates a new SlackNotifier if slack is enabled in the config, otherwise it returns a NoopNotifier
func NewSlackNotifier(slackConfig *SlackConfig) Notifier {
	if slackConfig.Enabled {
		return &SlackNotifier{
			Client:   slack.New(slackConfig.APIToken),
			Username: slackConfig.Username,
			Channel:  slackConfig.Channel,
			Emoji:    slackConfig.Emoji,
		}
	}
	return &NoopNotifier{}
}

// SendError formats an error message with a summary, error, and optional details, then sends it to the Slack channel
func (n *SlackNotifier) SendError(ctx context.Context, summary string, err error, details map[string]interface{}) error {
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

	_, _, sendErr := n.Client.PostMessage(
		n.Channel,
		slack.MsgOptionText(fmt.Sprintf(":x: %s", summary), false),
		slack.MsgOptionUsername(n.Username),
		slack.MsgOptionIconEmoji(n.Emoji),
		slack.MsgOptionAttachments(attachment),
	)

	if sendErr != nil {
		log.Error(ctx, "failed to send slack error message", sendErr, log.Data{"channel": n.Channel, "summary": summary, "error": err, "details": details})
		return sendErr
	}

	return nil
}
