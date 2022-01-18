package daemons

import (
	"analitics/pkg/transport"
	"os"
)

func (d *Daemon) StatusRun(app map[string]interface{}) {
	server := transport.NewServer()
	d.Logger().Info().Msg("Http server is starting...")
	err := server.Start()
	if err != nil {
		d.Logger().Error().Err(err).Msg("Server hasn't been started!")
		os.Exit(1)
	}

}
