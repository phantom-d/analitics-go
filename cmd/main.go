package main

import (
	"analitics/pkg/config"
	"analitics/pkg/database"
	"fmt"

	"github.com/phantom-d/go-daemons"
)

func main() {
	if config.App().Status {
		result, err := daemons.DaemonsStatus(config.App().Daemon)
		if err != nil {
			config.Log().Error().Err(err).Msgf("Daemon status for '%s'", config.App().Daemon)
		} else {
			fmt.Println(string(result))
		}
		return
	}
	database.Migrate()
	if daemon := daemons.New(config.App().Daemon); daemon != nil {
		if config.App().Worker == "" {
			if err := daemons.Start(daemon); err != nil {
				config.Log().Fatal().Err(err).Msgf("Start daemon '%s'", config.App().Daemon)
			}
		} else {
			if err := daemon.Run(); err != nil {
				config.Log().Fatal().Err(err).Msgf("Run daemon '%s'", config.App().Daemon)
			}
		}
	}
}
