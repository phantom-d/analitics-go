package main

import (
	"analitics/pkg/config"
	"analitics/pkg/daemons"
	"analitics/pkg/database"

	_ "go.uber.org/automaxprocs"
)

func main() {
	database.Migrate()
	daemon := daemons.New(config.Application.Daemon)
	if daemon != nil {
		daemon.Start(daemon, false)
	}
}
