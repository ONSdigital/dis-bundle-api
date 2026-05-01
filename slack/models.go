package slack

// List of custom emoji names used in Slack
const (
	InfoEmoji    Emoji = ":information_source:"
	WarningEmoji Emoji = ":warning:"
	AlarmEmoji   Emoji = ":rotating_light:"
	TimerEmoji   Emoji = ":hourglass_flowing_sand:"
	TickEmoji    Emoji = ":white_check_mark:"
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

// MessageRef represents a reference to a Slack message.
// It contains the channel ID and timestamp of the message, which can be used for updating the message later.
type MessageRef struct {
	ChannelID string
	Timestamp string
}

// Field represents a key-value pair to be included in the Slack message attachments.
type Field struct {
	Title string
	Value string
}
