package slack

import "context"

// NoopNotifier is a Notifier that does nothing, used when Slack notifications are disabled
type NoopNotifier struct{}

func (n *NoopNotifier) SendError(ctx context.Context, summary string, err error, details map[string]interface{}) error {
	return nil
}
