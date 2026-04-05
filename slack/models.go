package slackclient

import "time"

type Channel struct {
	ID   string
	Name string
	IsDM bool
}

type Message struct {
	Timestamp time.Time
	UserID    string
	UserName  string
	Text      string
}

type IncomingMessage struct {
	ChannelID string
	Message   Message
}
