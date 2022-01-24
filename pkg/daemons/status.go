package daemons

import (
	"analitics/pkg/config"
	"analitics/pkg/transport"
	"os"
)

func (d *Daemon) StatusRun() {
	server := transport.NewServer()
	config.Logger.Info().Msg("Http server is starting...")
	err := server.Start()
	if err != nil {
		config.Logger.Error().Err(err).Msg("Server hasn't been started!")
		os.Exit(1)
	}
}
