package main

import "github.com/rs/zerolog/log"

func main() {
	log.Info().Msg("dev start")
	panic("empty")
	log.Info().Msg("dev done")
	return
}
