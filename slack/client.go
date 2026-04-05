package slackclient

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Client struct {
	api    *slack.Client
	sm     *socketmode.Client
	users  map[string]string
	userMu sync.RWMutex
}

func New(botToken, appToken string) *Client {
	api := slack.New(botToken, slack.OptionAppLevelToken(appToken))
	sm := socketmode.New(api)
	return &Client{
		api:   api,
		sm:    sm,
		users: make(map[string]string),
	}
}

func (c *Client) LoadChannels() ([]Channel, error) {
	var channels []Channel

	// Public + private channels
	params := &slack.GetConversationsParameters{
		Types:           []string{"public_channel", "private_channel"},
		ExcludeArchived: true,
		Limit:           200,
	}
	for {
		convs, cursor, err := c.api.GetConversations(params)
		if err != nil {
			return nil, fmt.Errorf("get conversations: %w", err)
		}
		for _, ch := range convs {
			if ch.IsMember {
				channels = append(channels, Channel{ID: ch.ID, Name: ch.Name})
			}
		}
		if cursor == "" {
			break
		}
		params.Cursor = cursor
	}

	// DMs
	dmParams := &slack.GetConversationsParameters{
		Types: []string{"im"},
		Limit: 100,
	}
	for {
		convs, cursor, err := c.api.GetConversations(dmParams)
		if err != nil {
			break // DMs optional; don't hard fail
		}
		for _, ch := range convs {
			name := c.resolveUser(ch.User)
			channels = append(channels, Channel{ID: ch.ID, Name: name, IsDM: true})
		}
		if cursor == "" {
			break
		}
		dmParams.Cursor = cursor
	}

	return channels, nil
}

func (c *Client) LoadMessages(channelID string) ([]Message, error) {
	params := slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     50,
	}
	resp, err := c.api.GetConversationHistory(&params)
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}

	msgs := make([]Message, 0, len(resp.Messages))
	// API returns newest first; reverse for chronological display
	for i := len(resp.Messages) - 1; i >= 0; i-- {
		m := resp.Messages[i]
		if m.SubType != "" {
			continue // skip join/leave/etc
		}
		msgs = append(msgs, Message{
			Timestamp: parseTS(m.Timestamp),
			UserID:    m.User,
			UserName:  c.resolveUser(m.User),
			Text:      m.Text,
		})
	}
	return msgs, nil
}

func (c *Client) SendMessage(channelID, text string) error {
	_, _, err := c.api.PostMessage(channelID, slack.MsgOptionText(text, false))
	return err
}

// ListenForEvents runs the Socket Mode client and sends incoming messages to ch.
// Call this in a goroutine. It blocks until the client disconnects.
func (c *Client) ListenForEvents(ch chan<- IncomingMessage) {
	go func() {
		_ = c.sm.Run()
	}()

	for evt := range c.sm.Events {
		switch evt.Type {
		case socketmode.EventTypeEventsAPI:
			eventsAPI, ok := evt.Data.(slackevents.EventsAPIEvent)
			if !ok {
				continue
			}
			c.sm.Ack(*evt.Request)

			if eventsAPI.Type == slackevents.CallbackEvent {
				inner, ok := eventsAPI.InnerEvent.Data.(*slackevents.MessageEvent)
				if !ok || inner.SubType != "" {
					continue
				}
				ch <- IncomingMessage{
					ChannelID: inner.Channel,
					Message: Message{
						Timestamp: parseTS(inner.TimeStamp),
						UserID:    inner.User,
						UserName:  c.resolveUser(inner.User),
						Text:      inner.Text,
					},
				}
			}
		case socketmode.EventTypeConnecting:
			// ignore
		case socketmode.EventTypeConnected:
			// ignore
		}
	}
}

func (c *Client) resolveUser(userID string) string {
	if userID == "" {
		return "unknown"
	}
	c.userMu.RLock()
	name, ok := c.users[userID]
	c.userMu.RUnlock()
	if ok {
		return name
	}

	info, err := c.api.GetUserInfo(userID)
	if err != nil {
		return userID
	}
	displayName := info.Profile.DisplayName
	if displayName == "" {
		displayName = info.Name
	}

	c.userMu.Lock()
	c.users[userID] = displayName
	c.userMu.Unlock()
	return displayName
}

func parseTS(ts string) time.Time {
	if ts == "" {
		return time.Now()
	}
	// Slack timestamps are "1234567890.123456"
	sec, _ := strconv.ParseFloat(ts, 64)
	return time.Unix(int64(sec), 0)
}
