package main

import (
	"context"
	"github.com/f0xdl/unit-watch-bot/internal/app"
	"github.com/rs/zerolog/log"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	application, err := app.SetupApp()
	if err != nil {
		log.Fatal().Err(err).Msg("Error setting up app")
	}
	err = application.Run(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Error running app")
	}
	log.Info().Msg("app stopped")
	return
}
