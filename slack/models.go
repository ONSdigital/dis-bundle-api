package slack

// List of custom emoji names used in Slack
const (
	InfoEmoji    Emoji = ":information_source:"
	WarningEmoji Emoji = ":warning:"
	AlarmEmoji   Emoji = ":rotating_light:"
)

type Emoji string

func (e Emoji) String() string {
	return string(e)
}

// List of colours used in Slack attachments
const (
	RedColour    Colour = "danger"
	YellowColour Colour = "warning"
	GreenColour  Colour = "good"
)

type Colour string

func (c Colour) String() string {
	return string(c)
}
