package slack

import "context"

// NoopClient is a Client that does nothing, used when Slack notifications are disabled
type NoopClient struct{}

func (n *NoopClient) SendAlarm(ctx context.Context, summary string, err error, details map[string]interface{}) error {
	return nil
}

func (n *NoopClient) SendWarning(ctx context.Context, summary string, details map[string]interface{}) error {
	return nil
}

func (n *NoopClient) SendInfo(ctx context.Context, summary string, details map[string]interface{}) error {
	return nil
}
