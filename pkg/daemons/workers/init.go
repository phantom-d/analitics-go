package workers

import (
	"analitics/pkg/config"
	"analitics/pkg/database"
	"analitics/pkg/transport"
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

func (w *Worker) BeforeRun() (interface{}, error) {
	return config.RequestFunc(w, strings.Title(w.Name), 3)
}

func (w *Worker) AfterRun() (interface{}, error) {
	return config.RequestFunc(w, strings.Title(w.Name), 3)
}

func (w *Worker) BeforeIteration(data []map[string]interface{}) (interface{}, error) {
	return config.RequestFunc(w, strings.Title(w.Name), 3, data)
}

func (w *Worker) AfterIteration(errorItems []map[string]interface{}) (interface{}, error) {
	return config.RequestFunc(w, strings.Title(w.Name), 3, errorItems)
}

func (w *Worker) ExtractId(errorItems []map[string]interface{}) (result []string, err error) {
	resp, err := config.RequestFunc(w, strings.Title(w.Name), 3, errorItems)
	if resp != nil {
		result = resp.([]string)
	}
	return
}

func (w *Worker) Save(importData map[string]interface{}) (interface{}, error) {
	database.Reconnect()
	return config.RequestFunc(w, strings.Title(w.Name), 3, importData)
}

func (w *Worker) AddToQueue(params map[string]interface{}, errorItems []map[string]interface{}) bool {
	result := true
	items, _ := w.ExtractId(errorItems)

	if items != nil {
		tr := transport.New(params)
		result = tr.Client.ResendErrorItems(w.Queue, items)
	}

	return result
}
