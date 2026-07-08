package application

import (
	"context"
	"time"
)

// newSlackContext creates a new context with a timeout.
// This is used to ensure that Slack notifications do not get cancelled if the main context is cancelled.
func newSlackContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(
		context.WithoutCancel(ctx),
		timeout,
	)
}
