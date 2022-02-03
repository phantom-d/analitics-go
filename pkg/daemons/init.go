package daemons

import (
	"analitics/pkg/config"
	"github.com/mitchellh/mapstructure"
)

type Daemon interface {
	Run()
	SetData(data *DaemonData)
}

type DaemonData struct {
	Name        string                 `mapstructure:"Name"`
	MemoryLimit uint64                 `mapstructure:"MemoryLimit"`
	Workers     []config.Worker        `mapstructure:"Workers"`
	Params      map[string]interface{} `mapstructure:"Params"`
	Sleep       int64                  `mapstructure:"Sleep"`
}

type Factory map[string]func() Daemon

var factory = make(Factory)

func init() {
	factory.Register("watcher", func() Daemon { return &Watcher{} })
	factory.Register("import", func() Daemon { return &Import{} })
	factory.Register("status", func() Daemon { return &Status{} })
}

func (factory *Factory) Register(name string, factoryFunc func() Daemon) {
	(*factory)[name] = factoryFunc
}

func (factory *Factory) CreateInstance(name string) Daemon {
	return (*factory)[name]()
}

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
