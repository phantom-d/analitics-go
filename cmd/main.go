package main

import (
	"analitics/pkg/config"
	"analitics/pkg/daemons"
	"analitics/pkg/database"
)

func main() {
	database.Migrate()
	daemon := daemons.New(config.Application.Daemon)
	if daemon != nil {
		daemon.Data().Start(daemon, false)
	}
}
