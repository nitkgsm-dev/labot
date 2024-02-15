package main

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func commandPing(e *handler.CommandEvent) error {
	messenger, err := e.Client().WebhookManager().GetMessenger(e.Channel())
	if err != nil {
		return err
	}
	if _, err := messenger.Send(discord.NewMessageBuilder().SetContent("Pong!")); err != nil {
		return err
	}
	return e.CreateMessage(discord.NewMessageBuilder().SetContent("Pong!").BuildCreate())
}
