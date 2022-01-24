package daemons

import (
	"analitics/pkg/config"
	"strings"
)

type Daemon struct {
	Name        string
	MemoryLimit int64
	Workers     []config.Worker
	Params      map[string]interface{}
	Sleep       int64
}

func New(name string) *Daemon {
	d := &Daemon{}
	if cfg, ok := config.Application.Daemons[name]; ok {
		if cfg.Enabled {
			d.Name = name
			d.MemoryLimit = cfg.MemoryLimit
			d.Workers = cfg.Workers
			d.Params = cfg.Params
			d.Sleep = cfg.Sleep
			return d
		} else {
			config.Logger.Info().Msgf("Daemon '%s' is disabled!", name)
		}
	} else {
		config.Logger.Info().Msgf("Daemon '%s' not found!", name)
	}
	return nil
}

func (d *Daemon) Run() {
	config.Logger.Info().Msgf("Start daemon '%s'!", d.Name)
	_, _ = config.RequestFunc(d, strings.Title(d.Name), 3)
}
