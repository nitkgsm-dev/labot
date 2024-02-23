package console

import (
	"context"
	"errors"
	"github.com/chzyer/readline"
	"github.com/disgoorg/disgo/bot"
	"io"
	"log/slog"
	"strings"
)

type CommandEvent interface {
	context.Context

	Args() []string
	Command() string
	Client() bot.Client

	SendMessage(content string, args ...any) error
}

func NewCommandEvent(ctx context.Context, args []string, command string, client bot.Client) CommandEvent {
	return commandEventImpl{
		Context: ctx,
		args:    args,
		command: command,
		client:  client,
	}
}

var _ CommandEvent = (*commandEventImpl)(nil)

type commandEventImpl struct {
	context.Context
	args    []string
	command string
	client  bot.Client
}

func (c commandEventImpl) Client() bot.Client {
	return c.client
}

func (c commandEventImpl) Args() []string {
	return c.args
}

func (c commandEventImpl) Command() string {
	return c.command
}

func (c commandEventImpl) SendMessage(content string, args ...any) error {
	slog.Info(content, args...)
	return nil
}

var handlers = make(map[string]consoleHandler)
var completer = make(map[string]readline.DynamicCompleteFunc)

type consoleHandler func(ctx CommandEvent) error

func RegisterHandler(command string, handler consoleHandler) {
	handlers[command] = handler
	completer[command] = func(line string) []string {
		return nil
	}
}

func RegisterCompleter(command string, cmp readline.DynamicCompleteFunc) {
	completer[command] = cmp
}

func Completer() *readline.PrefixCompleter {
	var cs []readline.PrefixCompleterInterface
	for s := range handlers {
		cs = append(cs, readline.PcItem(s, readline.PcItemDynamic(completer[s])))
	}
	cs = append(cs, readline.PcItem("stop"))
	c := readline.NewPrefixCompleter(cs...)
	return c
}

func Listen(ctx context.Context, l *readline.Instance, clientChan <-chan bot.Client, cancelFunc context.CancelFunc) {
	defer cancelFunc()
	client, ok := <-clientChan
	if !ok {
		slog.Error("failed to get client")
		return
	}
loop:
	for {
		line, err := l.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				if len(line) == 0 {
					slog.Info("exiting...")
					return
				} else {
					continue loop
				}
			} else if errors.Is(err, io.EOF) {
				return
			}
			slog.Error("failed reading line", slog.Any("err", err))
			continue loop
		}
		args := strings.Split(strings.TrimSpace(line), " ")
		if len(args) == 0 {
			continue loop
		}

		_ = l.SaveHistory(line)

		for s, handler := range handlers {
			if strings.HasPrefix(args[0], s) {
				ctx := NewCommandEvent(ctx, args, s, client)
				if err := handler(ctx); err != nil {
					slog.Error("failed handling command", slog.Any("err", err))
				}
				continue loop
			}
		}

		if args[0] == "stop" {
			slog.Info("stopping bot...")
			return
		}
	}
}
