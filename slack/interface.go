package slack

import (
	"context"
)

//go:generate moq -out mocks/client.go -pkg mocks . Clienter

// Clienter represents an interface for a generic Client
type Clienter interface {
	SendAlarm(ctx context.Context, summary string, err error, fields []Field) (*MessageRef, error)
	SendWarning(ctx context.Context, summary string, fields []Field) (*MessageRef, error)
	SendInfo(ctx context.Context, summary string, fields []Field) (*MessageRef, error)
	SendPublishLog(ctx context.Context, summary string, fields []Field) (*MessageRef, error)
	UpdatePublishLog(ctx context.Context, ref *MessageRef, summary string, fields []Field) (*MessageRef, error)
	UpdatePublishLogAsAlarm(ctx context.Context, ref *MessageRef, summary string, fields []Field) (*MessageRef, error)
}
