package main

import (
	"context"
	"flag"
	"github.com/chzyer/readline"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
	"github.com/nitkgsm-dev/labot/pkg/logging"
	"log/slog"
	"os"
	"time"
)

var (
	jsonLog    = flag.Bool("json-log", false, "enable json log format")
	token      = flag.String("token", "", "discord bot token")
	debug      = flag.Bool("debug", false, "enable debug mode")
	dateFormat = flag.String("date-format", time.TimeOnly, "date format")
)

func init() {
	flag.Parse()
}

func main() {
	if *token == "" {
		*token = os.Getenv("DISCORD_TOKEN")
	}

	l, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     ".console_history",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		slog.Error("failed creating readline", slog.Any("err", err))
		panic(err)
	}
	defer l.Close()
	l.CaptureExitSignal()

	logFormat := logging.FormatText
	if *jsonLog {
		logFormat = logging.FormatJSON
	}

	logLevel := logging.LevelInfo
	if *debug {
		logLevel = logging.LevelDebug
	}

	logger := logging.DefaultBuilder().
		SetWriter(l.Stdout()).
		SetLogFormat(logFormat).
		SetDisplaySource(*debug).
		SetDateFormat(*dateFormat).
		SetLevel(logLevel).
		Build()
	slog.SetDefault(logger)

	mux := handler.New()
	mux.Use(middleware.Logger)

	mux.Command("/ping", func(e *handler.CommandEvent) error {
		messenger, err := e.Client().WebhookManager().GetMessenger(e.Channel())
		if err != nil {
			return err
		}
		if _, err := messenger.Send(discord.NewMessageBuilder().SetContent("pong")); err != nil {
			return err
		}
		return e.CreateMessage(discord.NewMessageBuilder().SetContent("pong").BuildCreate())
	})

	mux.NotFound(func(e *events.InteractionCreate) error {
		slog.Warn("not found", slog.Any("interaction", e.Interaction))
		return nil
	})

	ready := make(chan *events.Ready)

	client, err := disgo.New(*token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentsAll,
			),
		),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagsAll),
		),
		bot.WithEventManagerConfigOpts(
			bot.WithListeners(
				mux,
				bot.NewListenerChan(ready),
			),
			bot.WithAsyncEventsEnabled(),
		),
		bot.WithLogger(logger.WithGroup("disgo")),
	)
	if err != nil {
		slog.Error("failed creating client", slog.Any("err", err))
		panic(err)
	}

	if err := client.OpenGateway(context.Background()); err != nil {
		slog.Error("failed opening gateway", slog.Any("err", err))
		panic(err)
	}
	defer client.Close(context.Background())

	select {
	case <-ready:
		slog.Info("ready")
	case <-time.After(10 * time.Second):
		slog.Error("ready event not received")
		os.Exit(1)
	}

	if err := handler.SyncCommands(client, commands, nil); err != nil {
		slog.Error("failed syncing commands", slog.Any("err", err))
		panic(err)
	}

	handleConsole(l, client)
}
