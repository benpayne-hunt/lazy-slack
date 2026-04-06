# LazySlack

A simple terminal UI for Slack, written in Go.

![demo gif placeholder](assets/demo.gif)

[![Release](https://img.shields.io/github/v/release/yourusername/lazy-slack)](https://github.com/yourusername/lazy-slack/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/lazy-slack)](https://goreportcard.com/report/github.com/yourusername/lazy-slack)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

Are you tired of getting Slack-brained? Tired of opening a 200MB Electron app just to say "on it" to your team? Tired of your terminal workflow grinding to a halt because you need to check one message?

**LazySlack** lets you send and receive Slack messages without ever leaving your terminal. Browse channels, read history, compose messages, and watch new messages arrive in real-time — all from a fast, keyboard-driven TUI.

---

## Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Slack App Setup](#slack-app-setup)
- [Usage](#usage)
- [Keybindings](#keybindings)
- [Features](#features)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [FAQ](#faq)

---

## Requirements

- Go 1.21+
- A Slack workspace where you can install apps

---

## Installation

### Homebrew (macOS/Linux)

```sh
brew install yourusername/tap/lazy-slack
```

### Go install

```sh
go install github.com/yourusername/lazy-slack@latest
```

### Binary releases

Download the latest binary for your platform from the [releases page](https://github.com/yourusername/lazy-slack/releases).

### From source

```sh
git clone https://github.com/yourusername/lazy-slack.git
cd lazy-slack
go build -o lazy-slack .
```

---

## Slack App Setup

LazySlack connects to Slack via a personal Slack App using Socket Mode — no public HTTP endpoint required, works entirely from your machine.

**This takes about 5 minutes.**

1. Go to [api.slack.com/apps](https://api.slack.com/apps) and click **Create New App → From scratch**
2. Give it a name (e.g. "LazySlack") and select your workspace
3. Under **Settings → Socket Mode**, enable Socket Mode and generate an **App-Level Token** with the `connections:write` scope — save this as `SLACK_APP_TOKEN`
4. Under **OAuth & Permissions**, add these Bot Token Scopes:

   | Scope | Purpose |
   |-------|---------|
   | `channels:read` | List public channels |
   | `channels:history` | Read public channel messages |
   | `groups:read` | List private channels |
   | `groups:history` | Read private channel messages |
   | `im:read` | List direct messages |
   | `im:history` | Read direct messages |
   | `chat:write` | Send messages |
   | `users:read` | Resolve usernames |

5. Click **Install to Workspace** and copy the **Bot User OAuth Token** — save this as `SLACK_BOT_TOKEN`
6. Under **Event Subscriptions**, enable events and subscribe to the `message.channels` and `message.im` bot events

Then export your tokens:

```sh
export SLACK_BOT_TOKEN=xoxb-...
export SLACK_APP_TOKEN=xapp-...
```

You may want to add these to your shell profile (`.zshrc`, `.bashrc`, etc).

---

## Usage

```sh
lazy-slack
```

**Demo mode** (no Slack account needed):

```sh
lazy-slack --demo
```

### Alias tip

Add this to your shell profile for quick access:

```sh
alias lsl='lazy-slack'
```

---

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down in channel list |
| `k` / `↑` | Move up in channel list |
| `enter` | Select channel / open compose |
| `tab` | Switch focus between panels |
| `/` | Filter channel list |
| `esc` | Cancel / close compose |

### Messages

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll messages down |
| `k` / `↑` | Scroll messages up |
| `i` | Open compose bar |
| `g` | Jump to top |
| `G` | Jump to bottom |

### Compose

| Key | Action |
|-----|--------|
| `ctrl+s` | Send message |
| `esc` | Discard and close |

### Global

| Key | Action |
|-----|--------|
| `q` | Quit (when not composing) |
| `ctrl+c` | Quit |

---

## Features

- **Real-time messages** via Slack's Socket Mode — new messages appear instantly without polling
- **Channels and DMs** — browse all channels you're a member of and your direct messages
- **Message history** — load recent messages when you open a channel
- **Send messages** — compose and send from the terminal
- **Username resolution** — user IDs are resolved to display names automatically, with caching
- **Fuzzy channel filter** — press `/` to filter channels by name
- **Demo mode** — try the UI without any Slack credentials (`--demo`)

---

## Configuration

LazySlack is configured via environment variables:

| Variable | Description |
|----------|-------------|
| `SLACK_BOT_TOKEN` | Bot User OAuth Token (`xoxb-...`) |
| `SLACK_APP_TOKEN` | App-Level Token for Socket Mode (`xapp-...`) |

A config file and more options (themes, default channel, message limit) are planned for a future release.

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a PR.

To run the project locally without a real Slack workspace, use demo mode:

```sh
go run . --demo
```

To run tests:

```sh
go test ./...
```

If you'd like to work on a new feature, open an issue first so we can discuss the approach.

---

## FAQ

**Does this work with Slack's free tier?**

Yes — Socket Mode and the scopes used by LazySlack are available on all Slack plans.

**Why do I need to create a Slack App? Can't it just use my account?**

Slack's user token API is heavily restricted and being deprecated in some areas. App-based access via Bot tokens is the supported, stable path. The setup is a one-time 5-minute process.

**My channels aren't showing up.**

LazySlack only shows channels your bot has been added to. In Slack, go to a channel and type `/invite @YourAppName` to add the bot.

**Will this work with multiple workspaces?**

Not yet — multi-workspace support is on the roadmap.

**Can I read threads?**

Thread replies aren't shown yet. This is on the roadmap.

---

## Alternatives

- [wee-slack](https://github.com/wee-slack/wee-slack) — Slack plugin for WeeChat
- [slack-term](https://github.com/erroneousboat/slack-term) — older terminal Slack client (uses deprecated RTM API)
