package daemons

import (
	"analitics/pkg/transport"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
)

func (d Daemon) WatcherRun() {
	server := transport.NewServer(d.db)
	fmt.Println("server is starting...")
	err := server.Start()
	if err != nil {
		log.Error().Err(err).Msg("Server hasn't been started.")
		os.Exit(1)
	}

}
