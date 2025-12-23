package slack

import "context"

// NoopClient is a Client that does nothing, used when Slack notifications are disabled
type NoopClient struct{}

func (n *NoopClient) SendError(ctx context.Context, channel, summary string, err error, details map[string]interface{}) error {
	return nil
}

func (n *NoopClient) Channels() Channels {
	return Channels{}
}
