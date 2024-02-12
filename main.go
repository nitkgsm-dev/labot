package main

import (
	"context"
	"github.com/chzyer/readline"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
	"log/slog"
	"os"
)

func main() {
	token := os.Getenv("discord_token")

	l, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     ".labot/.console_history",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	l.CaptureExitSignal()

	slog.SetDefault(slog.New(slog.NewTextHandler(l.Stderr(), nil)))

	mux := handler.New()
	mux.Use(middleware.Logger)

	mux.Command("/ping", func(e *handler.CommandEvent) error {
		messenger, err := e.Client().WebhookManager().GetMessenger(e.Channel())
		if err != nil {
			return err
		}
		if _, err := messenger.Send(discord.NewMessageBuilder().SetContent("pong"), e.Client()); err != nil {
			return err
		}
		return e.CreateMessage(discord.NewMessageBuilder().SetContent("pong").BuildCreate())
	})

	mux.NotFound(func(e *events.InteractionCreate) error {
		slog.Warn("not found", slog.Any("interaction", e.Interaction))
		return nil
	})

	client, err := disgo.New(token, bot.WithDefaultGateway())
	if err != nil {
		panic(err)
	}

	client.AddEventListeners(mux)

	if err := handler.SyncCommands(client, commands, nil); err != nil {
		panic(err)
	}

	if err := client.OpenGateway(context.Background()); err != nil {
		panic(err)
	}
	defer client.Close(context.Background())

	handleConsole(l, client)
}
