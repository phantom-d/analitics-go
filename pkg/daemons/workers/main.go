package workers

import (
	"analitics/pkg/config"
	"analitics/pkg/transport"
)

type Worker struct {
	Name    string
	Queue   string
	Enabled bool
	Sleep   int64
	Job     Job
}

type Job interface {
	Save() (result interface{}, err error)
	ExtractId([]map[string]interface{}) (result []string, err error)
}

func New(cfg config.Worker) *Worker {
	if cfg.Enabled {
		worker := &Worker{
			Name:    cfg.Name,
			Queue:   cfg.Queue,
			Enabled: cfg.Enabled,
			Sleep:   cfg.Sleep,
			Job:     factory.CreateInstance(cfg.Name),
		}
		return worker
	} else {
		config.Logger.Info().Msgf("Worker '%s' is disabled!", cfg.Name)
	}
	return nil
}

func (w *Worker) BeforeRun() (interface{}, error) {
	return config.DynamicCall(w.Job, "BeforeRun")
}

func (w *Worker) AfterRun() (interface{}, error) {
	return config.DynamicCall(w.Job, "AfterRun")
}

func (w *Worker) BeforeIteration(data []map[string]interface{}) (interface{}, error) {
	return config.DynamicCall(w.Job, "BeforeIteration", data)
}

func (w *Worker) AfterIteration(errorItems []map[string]interface{}) (interface{}, error) {
	return config.DynamicCall(w.Job, "AfterIteration", errorItems)
}

func (w *Worker) AddToQueue(params map[string]interface{}, errorItems []map[string]interface{}) bool {
	result := true
	items, _ := w.Job.ExtractId(errorItems)

	if items != nil {
		tr := transport.New(params)
		result = tr.Client.ResendErrorItems(w.Queue, items)
	}

	return result
}
