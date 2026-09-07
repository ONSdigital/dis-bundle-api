package slack

import (
	"context"
	"time"
)

// NoopClient is a Client that does nothing, used when Slack notifications are disabled
type NoopClient struct{}

func (n *NoopClient) SendAlarm(ctx context.Context, title string, err error, details []Detail, links []Link) (*MessageRef, error) {
	return nil, nil
}

func (n *NoopClient) SendWarning(ctx context.Context, title string, details []Detail, links []Link) (*MessageRef, error) {
	return nil, nil
}

func (n *NoopClient) SendInfo(ctx context.Context, title string, details []Detail, links []Link) (*MessageRef, error) {
	return nil, nil
}

func (n *NoopClient) SendPublishLog(ctx context.Context, title string, details []Detail, links []Link) (*MessageRef, error) {
	return nil, nil
}

func (n *NoopClient) UpdateMessage(ctx context.Context, ref *MessageRef, title string, err error, details []Detail, links []Link, color Colour, emoji Emoji) (*MessageRef, error) {
	return nil, nil
}

func (n *NoopClient) GetTimeout() time.Duration {
	return 0
}
