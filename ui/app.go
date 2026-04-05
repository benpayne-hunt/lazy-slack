package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	slackclient "github.com/bensomething/lazy-slack/slack"
)

type focus int

const (
	focusChannels focus = iota
	focusMessages
	focusInput
)

// Messages used as tea.Msg
type channelsLoadedMsg struct{ channels []slackclient.Channel }
type channelsErrMsg struct{ err error }
type messagesLoadedMsg struct{ messages []slackclient.Message }
type messagesErrMsg struct{ err error }
type IncomingMessageMsg slackclient.IncomingMessage
type sendErrMsg struct{ err error }

type channelItem struct {
	channel slackclient.Channel
}

func (c channelItem) Title() string {
	if c.channel.IsDM {
		return "@" + c.channel.Name
	}
	return "#" + c.channel.Name
}
func (c channelItem) Description() string { return "" }
func (c channelItem) FilterValue() string  { return c.channel.Name }

type Model struct {
	client          slackclient.SlackClient
	channels        []slackclient.Channel
	selectedChannel *slackclient.Channel
	messages        []slackclient.Message
	focus           focus

	channelList list.Model
	msgViewport viewport.Model
	input       textarea.Model

	width  int
	height int

	statusMsg string
}

func New(client slackclient.SlackClient) Model {
	// Channel list
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetHeight(1)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("205")).
		BorderLeftForeground(lipgloss.Color("205"))

	cl := list.New(nil, delegate, 0, 0)
	cl.Title = "Channels"
	cl.SetShowStatusBar(false)
	cl.SetFilteringEnabled(true)
	cl.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

	// Viewport
	vp := viewport.New(0, 0)

	// Textarea
	ta := textarea.New()
	ta.Placeholder = "Type a message... (ctrl+s to send, esc to cancel)"
	ta.CharLimit = 4000
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	return Model{
		client:      client,
		channelList: cl,
		msgViewport: vp,
		input:       ta,
		focus:       focusChannels,
	}
}

func (m Model) Init() tea.Cmd {
	return loadChannels(m.client)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout()

	case channelsLoadedMsg:
		m.channels = msg.channels
		items := make([]list.Item, len(msg.channels))
		for i, ch := range msg.channels {
			items[i] = channelItem{ch}
		}
		m.channelList.SetItems(items)

	case channelsErrMsg:
		m.statusMsg = "Error loading channels: " + msg.err.Error()

	case messagesLoadedMsg:
		m.messages = msg.messages
		m.msgViewport.SetContent(m.renderMessages())
		m.msgViewport.GotoBottom()

	case messagesErrMsg:
		m.statusMsg = "Error loading messages: " + msg.err.Error()

	case IncomingMessageMsg:
		if m.selectedChannel != nil && m.selectedChannel.ID == msg.ChannelID {
			m.messages = append(m.messages, msg.Message)
			m.msgViewport.SetContent(m.renderMessages())
			m.msgViewport.GotoBottom()
		}

	case sendErrMsg:
		m.statusMsg = "Send error: " + msg.err.Error()

	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c", "q":
			if m.focus != focusInput {
				return m, tea.Quit
			}
		case "tab":
			if m.focus == focusChannels {
				m.focus = focusMessages
				m.input.Blur()
			} else if m.focus == focusMessages {
				m.focus = focusChannels
			}
			return m, nil
		case "i", "enter":
			if m.focus == focusMessages {
				m.focus = focusInput
				m.input.Focus()
				return m, textarea.Blink
			}
		case "esc":
			if m.focus == focusInput {
				m.focus = focusMessages
				m.input.Blur()
				return m, nil
			}
		case "ctrl+s":
			if m.focus == focusInput && m.selectedChannel != nil {
				text := strings.TrimSpace(m.input.Value())
				if text != "" {
					m.input.Reset()
					m.focus = focusMessages
					m.input.Blur()
					return m, sendMessage(m.client, m.selectedChannel.ID, text)
				}
			}
		}

		// Route keys to focused component
		switch m.focus {
		case focusChannels:
			var cmd tea.Cmd
			m.channelList, cmd = m.channelList.Update(msg)
			cmds = append(cmds, cmd)

			if msg.String() == "enter" {
				if item, ok := m.channelList.SelectedItem().(channelItem); ok {
					ch := item.channel
					m.selectedChannel = &ch
					m.messages = nil
					m.msgViewport.SetContent("")
					m.statusMsg = ""
					cmds = append(cmds, loadMessages(m.client, ch.ID))
				}
			}

		case focusMessages:
			var cmd tea.Cmd
			m.msgViewport, cmd = m.msgViewport.Update(msg)
			cmds = append(cmds, cmd)

		case focusInput:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}

	default:
		// Pass to all sub-components for timer/tick events
		var cmd tea.Cmd
		m.channelList, cmd = m.channelList.Update(msg)
		cmds = append(cmds, cmd)
		m.msgViewport, cmd = m.msgViewport.Update(msg)
		cmds = append(cmds, cmd)
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	sideW := 24
	mainW := m.width - sideW - 3 // 3 for borders/gap

	// Left panel
	m.channelList.SetSize(sideW, m.height-2)
	sideStyle := lipgloss.NewStyle().
		Width(sideW).
		Height(m.height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.borderColor(focusChannels))
	leftPane := sideStyle.Render(m.channelList.View())

	// Right panel: viewport + input
	inputH := 0
	inputView := ""
	if m.focus == focusInput || m.selectedChannel != nil {
		inputH = 5
		inputStyle := lipgloss.NewStyle().
			Width(mainW).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(m.borderColor(focusInput))
		m.input.SetWidth(mainW - 2)
		inputView = inputStyle.Render(m.input.View())
	}

	vpH := m.height - 2 - inputH
	if vpH < 1 {
		vpH = 1
	}
	m.msgViewport.Width = mainW
	m.msgViewport.Height = vpH - 2

	title := "No channel selected"
	if m.selectedChannel != nil {
		if m.selectedChannel.IsDM {
			title = "@" + m.selectedChannel.Name
		} else {
			title = "#" + m.selectedChannel.Name
		}
	}

	vpStyle := lipgloss.NewStyle().
		Width(mainW).
		Height(vpH).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.borderColor(focusMessages))

	vpContent := lipgloss.NewStyle().
		Bold(true).Foreground(lipgloss.Color("205")).
		Render(title) + "\n\n" + m.msgViewport.View()

	rightPane := lipgloss.JoinVertical(lipgloss.Left,
		vpStyle.Render(vpContent),
		inputView,
	)

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", rightPane)

	// Status bar
	statusStyle := lipgloss.NewStyle().
		Width(m.width).
		Foreground(lipgloss.Color("241")).
		Padding(0, 1)
	help := "tab: switch panel  •  enter/i: compose  •  ctrl+s: send  •  esc: cancel  •  q: quit"
	if m.statusMsg != "" {
		help = m.statusMsg
	}
	status := statusStyle.Render(help)

	return lipgloss.JoinVertical(lipgloss.Left, body, status)
}

func (m *Model) layout() {
	// Trigger resize on sub-components via View() when needed
}

func (m Model) borderColor(f focus) lipgloss.Color {
	if m.focus == f {
		return lipgloss.Color("205")
	}
	return lipgloss.Color("238")
}

func (m Model) renderMessages() string {
	if len(m.messages) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("No messages")
	}

	var sb strings.Builder
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))

	for _, msg := range m.messages {
		ts := timeStyle.Render(msg.Timestamp.Format("15:04"))
		name := nameStyle.Render(msg.UserName)
		sb.WriteString(fmt.Sprintf("%s  %s: %s\n", ts, name, msg.Text))
	}
	return sb.String()
}

// Commands

func loadChannels(client slackclient.SlackClient) tea.Cmd {
	return func() tea.Msg {
		channels, err := client.LoadChannels()
		if err != nil {
			return channelsErrMsg{err}
		}
		return channelsLoadedMsg{channels}
	}
}

func loadMessages(client slackclient.SlackClient, channelID string) tea.Cmd {
	return func() tea.Msg {
		msgs, err := client.LoadMessages(channelID)
		if err != nil {
			return messagesErrMsg{err}
		}
		return messagesLoadedMsg{msgs}
	}
}

func sendMessage(client slackclient.SlackClient, channelID, text string) tea.Cmd {
	return func() tea.Msg {
		if err := client.SendMessage(channelID, text); err != nil {
			return sendErrMsg{err}
		}
		// Reload messages after send to show our own message
		msgs, err := client.LoadMessages(channelID)
		if err != nil {
			return sendErrMsg{err}
		}
		return messagesLoadedMsg{msgs}
	}
}
