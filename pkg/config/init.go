package config

import (
	goflag "flag"
	"github.com/rs/zerolog"
	flag "github.com/spf13/pflag"
	"time"
)

type Config struct {
	ConfigPath string
	PidDir     string
	Daemon     string
	Worker     string
	Status     bool
	Debug      bool
	MigrateUp  bool
	Database   map[string]interface{} `yaml:"database"`
	Daemons    map[string]Daemon
	Signal     string
}

type Daemon struct {
	Name        string                 `yaml:"name"`
	Enabled     bool                   `yaml:"enabled"`
	MemoryLimit uint64                 `yaml:"memory-limit"`
	Sleep       time.Duration          `yaml:"sleep"`
	Workers     []Worker               `yaml:"workers"`
	Params      map[string]interface{} `yaml:"params"`
}

type Worker struct {
	Name        string        `yaml:"name"`
	MemoryLimit uint64        `yaml:"memory-limit"`
	Queue       string        `yaml:"queue"`
	Enabled     bool          `yaml:"enabled"`
	Sleep       time.Duration `yaml:"sleep"`
}

var (
	application *Config = &Config{}
	logger      zerolog.Logger
)

func init() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.StringVarP(&application.ConfigPath, "config", "c", "config.yaml", "Path to a config file")
	flag.StringVarP(&application.PidDir, "pid-dir", "p", "pids", "Path to a save pid files")
	flag.StringVarP(&application.Daemon, "daemon", "d", "watcher", "Daemon name to starting")
	flag.StringVarP(&application.Worker, "worker", "w", "", "Warker name to starting")
	flag.BoolVar(&application.MigrateUp, "migrate", false, "Start with migrate up")
	flag.BoolVar(&application.Status, "status", false, "Get daemons status")
	flag.BoolVar(&application.Debug, "debug", false, "Enable debug mode")
	flag.StringVarP(&application.Signal, "signal", "s", "", `Send signal to the daemon:
  quit — graceful shutdown
  stop — fast shutdown`)
	flag.Parse()
	GetLogger()
	GetConfig()
}
