package daemons

import (
	"analitics/pkg/config"
	"github.com/mitchellh/mapstructure"
)

func New(name string) Daemon {
	if cfg, ok := config.Application.Daemons[name]; ok {
		if cfg.Enabled {
			cfg.Name = name
			d := factory.CreateInstance(name)
			dd := &DaemonData{}
			err := mapstructure.Decode(cfg, &dd)
			if err != nil {
				config.Logger.Error().Err(err).Msg("")
				return nil
			}
			d.SetData(dd)
			return d
		} else {
			config.Logger.Info().Msgf("Daemon '%s' is disabled!", name)
		}
	} else {
		config.Logger.Info().Msgf("Daemon '%s' not found!", name)
	}
	return nil
}

func (d *DaemonData) Start(daemon Daemon, daemonize bool) {
	config.Logger.Info().Msgf("Start daemon '%s'!", d.Name)
	if daemonize {

	}
	daemon.Run()
}
