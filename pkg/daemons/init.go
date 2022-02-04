package daemons

import (
	"analitics/pkg/config"
	"os"
	"os/exec"
)

type Daemon interface {
	Run()
	Start(Daemon, bool)
	Data() *DaemonData
	SetData(*DaemonData)
}

type DaemonData struct {
	Name                string                 `mapstructure:"Name"`
	MemoryLimit         uint64                 `mapstructure:"MemoryLimit"`
	Workers             []config.Worker        `mapstructure:"Workers"`
	Params              map[string]interface{} `mapstructure:"Params"`
	Sleep               int64                  `mapstructure:"Sleep"`
	binPath             string
	binPerms            os.FileMode
	binHash             []byte
	cmd                 *exec.Cmd
	descriptorsReleased chan bool
	stop                chan struct{}
	done                chan struct{}
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
