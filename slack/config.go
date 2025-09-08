package slack

// SlackConfig holds configuration for sending Slack notifications
type SlackConfig struct {
	APIToken string `envconfig:"SLACK_API_TOKEN"`
	Username string `envconfig:"SLACK_USERNAME"`
	Channel  string `envconfig:"SLACK_CHANNEL"`
	Emoji    string `envconfig:"SLACK_EMOJI"`
	Enabled  bool   `envconfig:"SLACK_ENABLED"`
}
