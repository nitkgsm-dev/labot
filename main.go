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
	"github.com/nitkgsm-dev/labot/pkg/console"
	"github.com/nitkgsm-dev/labot/pkg/logging"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
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
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	registerConsoleCommands()

	cmd, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     ".console_history",
		AutoComplete:    console.Completer(),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		slog.Error("failed to initiate cmd", slog.Any("err", err))
		done()
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
	ctx = logging.WithLogger(ctx, logger)

	clientChan := make(chan bot.Client)
	go console.Listen(ctx, cmd, clientChan, done)

	err = realMain(ctx, clientChan)
	done()

	if err != nil {
		slog.Error("unexpected error", slog.Any("err", err))
		os.Exit(1)
	}
}

func realMain(ctx context.Context, clientChan chan<- bot.Client) error {
	logger := logging.FromContext(ctx)

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

	if err := client.OpenGateway(ctx); err != nil {
		return fmt.Errorf("failed opening gateway: %w", err)
	}

	select {
	case <-ready:
		slog.Info("ready")
	case <-time.After(10 * time.Second):
		return errors.New("ready event not received")
	}

	defer client.Close(ctx)
	clientChan <- client

	if err := handler.SyncCommands(client, commands, nil); err != nil {
		return fmt.Errorf("failed syncing commands: %w", err)
	}

	<-ctx.Done()

	return nil
}
