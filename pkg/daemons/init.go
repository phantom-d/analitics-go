package daemons

import (
	"analitics/pkg/config"
)

type Daemon interface {
	Run()
	Data() *DaemonData
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
