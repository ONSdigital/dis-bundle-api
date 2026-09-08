package slack

import (
	"context"
	"time"
)

//go:generate moq -out mocks/client.go -pkg mocks . Clienter

// Clienter represents an interface for a generic Client
type Clienter interface {
	SendAlarm(ctx context.Context, title string, err error, details []Detail, links []Link) (*MessageRef, error)
	SendWarning(ctx context.Context, title string, details []Detail, links []Link) (*MessageRef, error)
	SendInfo(ctx context.Context, title string, details []Detail, links []Link) (*MessageRef, error)
	SendPublishLog(ctx context.Context, title string, details []Detail, links []Link) (*MessageRef, error)
	UpdateMessage(ctx context.Context, ref *MessageRef, title string, err error, details []Detail, links []Link, color Colour, emoji Emoji) (*MessageRef, error)
	GetTimeout() time.Duration
}

// Ensure that Client and NoopClient implement the Clienter interface.
var _ Clienter = (*Client)(nil)
var _ Clienter = (*NoopClient)(nil)
