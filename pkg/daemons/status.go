package daemons

import (
	"analitics/pkg/config"
	"analitics/pkg/transport"
	"os"
)

type Status struct {
	*DaemonData
}

func (st *Status) SetData(data *DaemonData) {
	st.DaemonData = data
}

func (st *Status) Data() *DaemonData {
	return st.DaemonData
}

func (st *Status) Run() {
	server := transport.NewServer()
	config.Logger.Info().Msg("Http server is starting...")
	err := server.Start()
	if err != nil {
		config.Logger.Error().Err(err).Msg("Server hasn't been started!")
		os.Exit(1)
	}
}
