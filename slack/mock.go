package slackclient

import (
	"fmt"
	"strings"
	"time"
)

// MockClient returns fake data for UI development without real Slack credentials.
type MockClient struct {
	messages map[string][]Message
}

func NewMock() *MockClient {
	now := time.Now()
	ago := func(m int) time.Time { return now.Add(-time.Duration(m) * time.Minute) }

	msgs := map[string][]Message{
		"C001": {
			{ago(42), "U1", "alice", "morning everyone"},
			{ago(40), "U2", "bob", "hey! anyone looked at the deploy logs?"},
			{ago(38), "U1", "alice", "yeah there's a weird spike around 9am but it recovered"},
			{ago(35), "U3", "carol", "I saw that too — think it was the cache warming up after the restart"},
			{ago(30), "U2", "bob", "makes sense, I'll add a note to the runbook"},
			{ago(20), "U1", "alice", "standup in 10 btw"},
			{ago(10), "U3", "carol", ":+1:"},
			{ago(2), "U2", "bob", "on my way"},
		},
		"C002": {
			{ago(120), "U2", "bob", "anyone catch the game last night?"},
			{ago(118), "U1", "alice", "yes!! that third quarter comeback was insane"},
			{ago(115), "U3", "carol", "I fell asleep during halftime lol"},
			{ago(60), "U4", "dave", "https://twitter.com/sports — highlights here if anyone missed it"},
			{ago(15), "U2", "bob", "dave clutch as always"},
		},
		"C003": {
			{ago(200), "U1", "alice", "PR is up for the new auth flow — would love a second pair of eyes"},
			{ago(195), "U3", "carol", "on it"},
			{ago(180), "U3", "carol", "left some comments, mostly nits but one thing about the token expiry logic"},
			{ago(170), "U1", "alice", "good catch on the expiry, fixing now"},
			{ago(90), "U1", "alice", "pushed a fix, lmk if it looks good"},
			{ago(45), "U3", "carol", "LGTM, approved"},
			{ago(40), "U1", "alice", ":tada: merging"},
		},
		"D001": {
			{ago(300), "U2", "bob", "hey got a sec?"},
			{ago(298), "U0", "you", "sure what's up"},
			{ago(295), "U2", "bob", "can you review my PR when you get a chance? no rush"},
			{ago(290), "U0", "you", "yeah I'll take a look this afternoon"},
			{ago(30), "U2", "bob", "thanks!"},
		},
		"D002": {
			{ago(500), "U3", "carol", "heads up — I'll be OOO Thursday and Friday"},
			{ago(498), "U0", "you", "noted, thanks for the heads up!"},
			{ago(496), "U3", "carol", "and I'll make sure to hand off the on-call rotation"},
			{ago(494), "U0", "you", "perfect"},
		},
	}

	return &MockClient{messages: msgs}
}

func (m *MockClient) LoadChannels() ([]Channel, error) {
	return []Channel{
		{ID: "C001", Name: "general"},
		{ID: "C002", Name: "random"},
		{ID: "C003", Name: "dev"},
		{ID: "D001", Name: "bob", IsDM: true},
		{ID: "D002", Name: "carol", IsDM: true},
	}, nil
}

func (m *MockClient) LoadMessages(channelID string) ([]Message, error) {
	msgs, ok := m.messages[channelID]
	if !ok {
		return nil, nil
	}
	return msgs, nil
}

func (m *MockClient) SendMessage(channelID, text string) error {
	msg := Message{
		Timestamp: time.Now(),
		UserID:    "U0",
		UserName:  "you",
		Text:      text,
	}
	m.messages[channelID] = append(m.messages[channelID], msg)
	return nil
}

func (m *MockClient) ListenForEvents(ch chan<- IncomingMessage) {
	// Simulate a few incoming messages with delays
	go func() {
		time.Sleep(4 * time.Second)
		ch <- IncomingMessage{
			ChannelID: "C001",
			Message: Message{
				Timestamp: time.Now(),
				UserID:    "U1",
				UserName:  "alice",
				Text:      fmt.Sprintf("(demo) new message at %s", time.Now().Format("15:04:05")),
			},
		}

		ticker := time.NewTicker(8 * time.Second)
		defer ticker.Stop()
		names := []string{"bob", "carol", "dave"}
		texts := []string{
			"just pushed a fix",
			"can someone review my PR?",
			"deploy looks good :white_check_mark:",
			"heading out for lunch, back in an hour",
			"anyone free for a quick call?",
		}
		i := 0
		for range ticker.C {
			ch <- IncomingMessage{
				ChannelID: "C001",
				Message: Message{
					Timestamp: time.Now(),
					UserID:    "U" + fmt.Sprint(i%3+1),
					UserName:  names[i%len(names)],
					Text:      texts[i%len(texts)] + " " + strings.Repeat(".", i%3+1),
				},
			}
			i++
		}
	}()
}
