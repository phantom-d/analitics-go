package daemons

import (
	"analitics/pkg/config"
	"analitics/pkg/transport"
)

type Status struct {
	*DaemonData
}

func (st *Status) SetData(data *DaemonData) {
	st.DaemonData = data
}

func (st *Status) Run() (err error) {
	server := transport.NewServer()
	config.Logger.Info().Msg("Http server is starting...")
	err = server.Start()
	return
}
