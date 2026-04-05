package config

import (
	"fmt"
	"os"
)

type Config struct {
	BotToken string
	AppToken string
}

func Load() (*Config, error) {
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("SLACK_BOT_TOKEN environment variable not set")
	}
	appToken := os.Getenv("SLACK_APP_TOKEN")
	if appToken == "" {
		return nil, fmt.Errorf("SLACK_APP_TOKEN environment variable not set")
	}
	return &Config{BotToken: botToken, AppToken: appToken}, nil
}
