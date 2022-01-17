package main

import (
	"analitics/pkg/application"
	goflag "flag"
	flag "github.com/spf13/pflag"
)

func main() {
	cfg := &application.Application{}
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.StringVarP(&cfg.ConfigPath, "config", "c", "config.yaml", "Path to a config file")
	flag.StringVarP(&cfg.Daemon, "daemon", "d", "watcher", "Daemon name to starting")
	flag.BoolVar(&cfg.MigrateUp, "migrate", false, "Start with migrate up")
	flag.BoolVar(&cfg.Debug, "debug", false, "Enable debug mode")
	flag.Parse()

	app := application.New(cfg)
	app.Run()
}
