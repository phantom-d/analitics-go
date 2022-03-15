package main

import (
	"analitics/pkg/config"
	"analitics/pkg/daemons"
	"analitics/pkg/database"
	"fmt"
)

func main() {
	if config.Application.Status {
		result, err := daemons.DaemonsStatus(config.Application.Daemon)
		if err != nil {
			config.Logger.Error().Err(err).Msgf("Daemon status for '%s'", config.Application.Daemon)
		} else {
			fmt.Println(string(result))
		}
		return
	}
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
