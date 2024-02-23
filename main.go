package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
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
	err := realMain()
	if err != nil {
		slog.Error("unexpected error", slog.Any("err", err))
		os.Exit(1)
	}
}

func realMain() error {
	cmd, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     ".console_history",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		slog.Error("failed to initiate cmd", slog.Any("err", err))
		os.Exit(1)
	}

	defer cmd.Close()
	cmd.CaptureExitSignal()

	logFormat := logging.FormatText
	if *jsonLog {
		logFormat = logging.FormatJSON
	}

	logLevel := logging.LevelInfo
	if *debug {
		logLevel = logging.LevelDebug
	}

	logger := logging.DefaultBuilder().
		SetWriter(cmd.Stdout()).
		SetLogFormat(logFormat).
		SetDisplaySource(*debug).
		SetDateFormat(*dateFormat).
		SetLevel(logLevel).
		Build()
	slog.SetDefault(logger)
	if *token == "" {
		*token = os.Getenv("DISCORD_TOKEN")
	}

	mux := handler.New()
	mux.Use(middleware.Logger)

	registerCommands(mux)

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
		return fmt.Errorf("failed creating client: %w", err)
	}

	if err := client.OpenGateway(context.Background()); err != nil {
		return fmt.Errorf("failed opening gateway: %w", err)
	}

	select {
	case <-ready:
		slog.Info("ready")
	case <-time.After(10 * time.Second):
		return errors.New("ready event not received")
	}

	defer client.Close(context.Background())

	if err := handler.SyncCommands(client, commands, nil); err != nil {
		return fmt.Errorf("failed syncing commands: %w", err)
	}

	handleConsole(cmd, client)

	return nil
}
