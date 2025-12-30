package slack

// SlackConfig holds configuration for sending Slack notifications
type SlackConfig struct {
	Channels Channels
}

// Channels holds the Slack channel names for different notification levels
type Channels struct {
	InfoChannel    string `envconfig:"SLACK_INFO_CHANNEL"`
	WarningChannel string `envconfig:"SLACK_WARNING_CHANNEL"`
	AlarmChannel   string `envconfig:"SLACK_ALARM_CHANNEL"`
}

// validateSlackConfig checks that all required fields are set in the SlackConfig
func validateSlackConfig(cfg *SlackConfig, apiToken string) error {
	if apiToken == "" {
		return errMissingAPIToken
	}
	if cfg.Channels.InfoChannel == "" {
		return errMissingInfoChannel
	}
	if cfg.Channels.WarningChannel == "" {
		return errMissingWarningChannel
	}
	if cfg.Channels.AlarmChannel == "" {
		return errMissingAlarmChannel
	}
	return nil
}
