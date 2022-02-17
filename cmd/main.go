package main

import (
	"analitics/pkg/config"
	"analitics/pkg/daemons"
	"analitics/pkg/database"
	_ "go.uber.org/automaxprocs"
)

func main() {
	database.Migrate()
	if daemon := daemons.New(config.Application.Daemon); daemon != nil {
		if err := daemons.Start(daemon); err != nil {
			config.Logger.Fatal().Err(err).Msgf("Daemon '%s'", config.Application.Daemon)
		}
	}
}
