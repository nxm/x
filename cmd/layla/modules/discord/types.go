package discord

import (
	"fmt"
	"git.jakub.app/jakub/X/internal/env"

	"github.com/bwmarrin/discordgo"
)

var (
	DISCORD_TOKEN = env.GetEnv("DISCORD_TOKEN", "")
)

type Discord struct {
	dg *discordgo.Session
}

func New() (*Discord, error) {
	dg, err := discordgo.New("Bot " + DISCORD_TOKEN)
	if err != nil {
		return nil, fmt.Errorf("discord: error creating discord session: %w", err)
	}

	return &Discord{
		dg: dg,
	}, nil
}

func (d *Discord) Session() *discordgo.Session {
	return d.dg
}
