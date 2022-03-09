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
		if config.Application.Worker == "" {
			if err := daemons.Start(daemon); err != nil {
				config.Logger.Fatal().Err(err).Msgf("Start daemon '%s'", config.Application.Daemon)
			}
		} else {
			if err := daemon.Run(); err != nil {
				config.Logger.Fatal().Err(err).Msgf("Run daemon '%s'", config.Application.Daemon)
			}
		}
	}
}
