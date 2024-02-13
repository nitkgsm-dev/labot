package main

import (
	"errors"
	"github.com/chzyer/readline"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"io"
	"log/slog"
	"strings"
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("stop"),
	readline.PcItem("ping"),
	readline.PcItem("guild",
		readline.PcItem("list"),
	),
)

func handleConsole(l *readline.Instance, client bot.Client) {
loop:
	for {
		line, err := l.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				if len(line) == 0 {
					slog.Info("exiting...")
					break loop
				} else {
					continue
				}
			} else if errors.Is(err, io.EOF) {
				break loop
			}
			slog.Error("failed reading line", slog.Any("err", err))
			continue
		}
		args := strings.Split(strings.TrimSpace(line), " ")
		if len(args) == 0 {
			continue
		}

		_ = l.SaveHistory(line)

		switch args[0] {
		case "stop":
			slog.Info("stopping bot...")
			break loop
		case "ping":
			slog.Info("gateway latency info", slog.Duration("latency", client.Gateway().Latency()))
		case "guild":
			if len(args) < 2 {
				slog.Info("missing args")
				continue
			}
			switch args[1] {
			case "list":
				slog.Info("guilds", slog.Int("total_guilds", client.Caches().GuildsLen()))
				client.Caches().GuildsForEach(func(guild discord.Guild) {
					member, ok := client.Caches().SelfMember(guild.ID)
					slog.Info("guild",
						slog.String("name", guild.Name),
						slog.String("id", guild.ID.String()),
						slog.String("owner_id", guild.OwnerID.String()),
						slog.Int("member_count", guild.MemberCount),
						slog.Bool("member_cached", ok),
						slog.Any("self_member_roles", member.RoleIDs),
					)
				})
			}
		}
	}
}
