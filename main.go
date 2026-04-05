package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bensomething/lazy-slack/config"
	slackclient "github.com/bensomething/lazy-slack/slack"
	"github.com/bensomething/lazy-slack/ui"
)

func main() {
	demo := len(os.Args) > 1 && os.Args[1] == "--demo"

	var client slackclient.SlackClient
	if demo {
		client = slackclient.NewMock()
	} else {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			fmt.Fprintln(os.Stderr, "Set SLACK_BOT_TOKEN and SLACK_APP_TOKEN, or run with --demo.")
			os.Exit(1)
		}
		client = slackclient.New(cfg.BotToken, cfg.AppToken)
	}

	incoming := make(chan slackclient.IncomingMessage, 64)
	model := ui.New(client)
	p := tea.NewProgram(model, tea.WithAltScreen())

	go func() { client.ListenForEvents(incoming) }()
	go func() {
		for msg := range incoming {
			p.Send(ui.IncomingMessageMsg(msg))
		}
	}()

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
