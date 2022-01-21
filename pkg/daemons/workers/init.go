package workers

import (
	"analitics/pkg/config"
	"analitics/pkg/database"
	"strings"
)

type Worker struct {
	Name    string
	Queue   string
	Enabled bool
	Sleep   int64
}

func New(cfg config.Worker) *Worker {
	if cfg.Enabled {
		worker := &Worker{cfg.Name, cfg.Queue, cfg.Enabled, cfg.Sleep}
		return worker
	} else {
		config.Logger.Info().Msgf("Worker '%s' is disabled!", cfg.Name)
	}
	return nil
}

func (w *Worker) Run(data []map[string]interface{}) {
	database.Reconnect()
	config.Logger.Info().Msgf("Start worker '%s'!", w.Name)
	funcName := strings.Title(w.Name) + "Run"
	args := make(map[string]interface{}, 0)
	args["arg0"] = data
	config.DynamicCall(w, funcName, args)
}
