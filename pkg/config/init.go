package config

import (
	goflag "flag"
	"github.com/rs/zerolog"
	flag "github.com/spf13/pflag"
	"os"
	"path/filepath"
)

type Config struct {
	ConfigPath string
	PidDir     string
	Daemon     string
	Debug      bool
	MigrateUp  bool
	Database   map[string]interface{} `yaml:"database"`
	Daemons    map[string]Daemon
}

type Daemon struct {
	Name        string                 `yaml:"name"`
	Enabled     bool                   `yaml:"enabled"`
	MemoryLimit int64                  `yaml:"memory-limit"`
	Sleep       int64                  `yaml:"sleep"`
	Workers     []Worker               `yaml:"workers"`
	Params      map[string]interface{} `yaml:"params"`
}

type Worker struct {
	Name    string `yaml:"name"`
	Queue   string `yaml:"queue"`
	Enabled bool   `yaml:"enabled"`
	Sleep   int64  `yaml:"sleep"`
}

var (
	Application *Config = &Config{}
	Logger      zerolog.Logger
)

func init() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.StringVarP(&Application.ConfigPath, "config", "c", "config.yaml", "Path to a config file")
	flag.StringVarP(&Application.PidDir, "pids", "p", "./pids", "Path to a save pid files")
	flag.StringVarP(&Application.Daemon, "daemon", "d", "watcher", "Daemon name to starting")
	flag.BoolVar(&Application.MigrateUp, "migrate", false, "Start with migrate up")
	flag.BoolVar(&Application.Debug, "debug", false, "Enable debug mode")
	flag.Parse()

	if filepath.IsAbs(Application.PidDir) == false {
		PidPath, err := filepath.Abs(Application.PidDir)
		if err != nil {
			Logger.Fatal().Err(err).Msg("")
		}
		Application.PidDir = PidPath
	}

	if _, err := os.Stat(Application.PidDir); os.IsNotExist(err) {
		err := os.Mkdir(Application.PidDir, os.ModePerm)
		if err != nil {
			Logger.Fatal().Err(err).Msg("Not read config file!")
		}
	}

	GetLogger()
	GetConfig()
}
