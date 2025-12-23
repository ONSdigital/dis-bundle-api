package slack

import (
	"context"

	"github.com/slack-go/slack"
)

//go:generate moq -out mocks/slack.go -pkg mocks . SlackClient
//go:generate moq -out mocks/client.go -pkg mocks . Clienter

// Clienter represents an interface for a generic Client
type Clienter interface {
	Channels() Channels
	SendError(ctx context.Context, channel string, summary string, err error, details map[string]interface{}) error
}

// SlackClient is an interface to enable mocking of the slack-go/slack.Client
type SlackClient interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
}
