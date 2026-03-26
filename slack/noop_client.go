package slack

import "context"

// NoopClient is a Client that does nothing, used when Slack notifications are disabled
type NoopClient struct{}

func (n *NoopClient) SendAlarm(ctx context.Context, summary string, err error, fields []Field) (*MessageRef, error) {
	return nil, nil
}

func (n *NoopClient) SendWarning(ctx context.Context, summary string, fields []Field) (*MessageRef, error) {
	return nil, nil
}

func (n *NoopClient) SendInfo(ctx context.Context, summary string, fields []Field) (*MessageRef, error) {
	return nil, nil
}

func (n *NoopClient) SendPublishLog(ctx context.Context, summary string, fields []Field) (*MessageRef, error) {
	return nil, nil
}

func (n *NoopClient) UpdatePublishLog(ctx context.Context, ref *MessageRef, summary string, fields []Field) (*MessageRef, error) {
	return nil, nil
}
