package main

import (
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/nitkgsm-dev/labot/pkg/console"
	"log/slog"
)

func registerConsoleCommands() {
	console.RegisterHandler("ping", func(ctx console.CommandEvent) error {
		return ctx.SendMessage("gateway latency info", slog.Duration("latency", ctx.Client().Gateway().Latency()))
	})
	console.RegisterHandler("guild", func(ctx console.CommandEvent) (returnErr error) {
		if len(ctx.Args()) < 2 {
			return ctx.SendMessage("missing args")
		}
		switch ctx.Args()[1] {
		case "list":
			if err := ctx.SendMessage("guilds", slog.Int("total_guilds", ctx.Client().Caches().GuildsLen())); err != nil {
				defer func() { returnErr = fmt.Errorf("%w: %w", returnErr, err) }()
			}
			ctx.Client().Caches().GuildsForEach(func(guild discord.Guild) {
				member, ok := ctx.Client().Caches().SelfMember(guild.ID)
				if err := ctx.SendMessage("guild",
					slog.String("name", guild.Name),
					slog.String("id", guild.ID.String()),
					slog.String("owner_id", guild.OwnerID.String()),
					slog.Int("member_count", guild.MemberCount),
					slog.Bool("member_cached", ok),
					slog.Any("self_member_roles", member.RoleIDs),
				); err != nil {
					defer func() { returnErr = fmt.Errorf("%w: %w", returnErr, err) }()
				}
			})
			return
		default:
			return ctx.SendMessage("unknown subcommand")
		}
	})
	console.RegisterCompleter("guild", func(s string) []string {
		slog.Debug(s)
		return []string{"list"}
	})
}
