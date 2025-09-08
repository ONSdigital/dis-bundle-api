package slack

import (
	"context"

	"github.com/slack-go/slack"
)

//go:generate moq -out mocks/slack.go -pkg mocks . SlackClient
//go:generate moq -out mocks/notifier.go -pkg mocks . Notifier

// Notifier represents an interface for a generic notifier
type Notifier interface {
	SendError(ctx context.Context, summary string, err error, details map[string]interface{}) error
}

// SlackClient is an interface to enable mocking of the slack-go/slack.Client
type SlackClient interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
}
