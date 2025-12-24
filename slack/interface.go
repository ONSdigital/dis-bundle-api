package slack

import (
	"context"
)

//go:generate moq -out mocks/client.go -pkg mocks . Clienter

// Clienter represents an interface for a generic Client
type Clienter interface {
	SendAlarm(ctx context.Context, summary string, err error, details map[string]interface{}) error
	SendWarning(ctx context.Context, summary string, details map[string]interface{}) error
	SendInfo(ctx context.Context, summary string, details map[string]interface{}) error
}
