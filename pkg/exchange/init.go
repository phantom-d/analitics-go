package exchange

import "analitics/pkg/application"

type Exchange struct {
	Config application.Daemon
}
type ExchangeHandler struct {
}

func New(name string, app *application.Application) *Exchange {
	cfg := app.GetConfig()
	exch := &Exchange{}

	if daemon, found := cfg.Daemons[name]; found {
		exch.Config = daemon
	}
	return exch
}

func (exch *Exchange) Start() {

}
