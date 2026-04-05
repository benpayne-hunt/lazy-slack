package slackclient

// SlackClient is the interface the UI depends on.
type SlackClient interface {
	LoadChannels() ([]Channel, error)
	LoadMessages(channelID string) ([]Message, error)
	SendMessage(channelID, text string) error
	ListenForEvents(ch chan<- IncomingMessage)
}
