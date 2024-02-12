package main

import (
	"github.com/chzyer/readline"
	"github.com/disgoorg/disgo/bot"
	"log/slog"
	"strings"
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("stop"),
)

func handleConsole(l *readline.Instance, client bot.Client) {
loop:
	for {
		line, err := l.Readline()
		if err != nil {
			slog.Error("failed reading line", slog.Any("err", err))
		}
		args := strings.Split(strings.TrimSpace(line), " ")
		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "stop":
			break loop
		}
	}
}
