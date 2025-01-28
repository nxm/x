package main

import (
	"git.jakub.app/jakub/X/cmd/layla/modules/discord"
	"github.com/rs/zerolog/log"
)

type svc struct {
	discordModule *discord.Discord
}

func run() error {
	var err error
	svc := svc{}

	svc.discordModule, err = discord.New()
	if err != nil {
		log.Error().Err(err).Msg("can't load discord module")
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
