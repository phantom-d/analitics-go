package config

import (
	goflag "flag"
	"github.com/rs/zerolog"
	flag "github.com/spf13/pflag"
)

type Config struct {
	ConfigPath string
	PidDir     string
	Daemon     string
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
	Sleep       int64                  `yaml:"sleep"`
	Workers     []Worker               `yaml:"workers"`
	Params      map[string]interface{} `yaml:"params"`
}

type Worker struct {
	Name        string `yaml:"name"`
	MemoryLimit uint64 `yaml:"memory-limit"`
	Queue       string `yaml:"queue"`
	Enabled     bool   `yaml:"enabled"`
	Sleep       int64  `yaml:"sleep"`
}

var (
	Application *Config = &Config{}
	Logger      zerolog.Logger
)

func init() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.StringVarP(&Application.ConfigPath, "config", "c", "config.yaml", "Path to a config file")
	flag.StringVarP(&Application.PidDir, "pid-dir", "p", "pids", "Path to a save pid files")
	flag.StringVarP(&Application.Daemon, "daemon", "d", "watcher", "Daemon name to starting")
	flag.BoolVar(&Application.MigrateUp, "migrate", false, "Start with migrate up")
	flag.BoolVar(&Application.Debug, "debug", false, "Enable debug mode")
	flag.StringVarP(&Application.Signal, "signal", "s", "", `Send signal to the daemon:
  quit — graceful shutdown
  stop — fast shutdown`)
	flag.Parse()
	GetLogger()
	GetConfig()
}
