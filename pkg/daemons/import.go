package daemons

import (
	"analitics/pkg/config"
	"analitics/pkg/daemons/workers"
	"analitics/pkg/database"
	"analitics/pkg/transport"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"time"
)

type resultProcess struct {
	PackageID  int64
	Queue      string
	Duration   time.Duration
	Total      int
	ErrorItems []map[string]interface{}
}

type resultLog struct {
	Queue             string  `json:"queue"`
	Duration          float64 `json:"duration"`
	DurationFormatted string  `json:"duration_formatted"`
	Total             int     `json:"total"`
	Imported          int     `json:"imported"`
	Errors            int     `json:"errors"`
}

type Import struct {
	*DaemonData
}

func (imp *Import) SetData(data *DaemonData) {
	imp.DaemonData = data
}

func (imp *Import) Run() {
	for {
		for _, cfg := range imp.Workers {
			worker := workers.New(cfg)
			if worker != nil {
				_, err := worker.BeforeRun()
				if err != nil {
					config.Logger.Error().Err(err).Msg("")
				}
				imp.Process(worker)
				_, err = worker.AfterRun()
				if err != nil {
					config.Logger.Error().Err(err).Msg("")
				}
			}
		}
		time.Sleep(time.Duration(imp.Sleep) * time.Second)
	}
}

func (imp *Import) Process(w *workers.Worker) {
	config.Logger.Info().Msgf("Start worker '%s'!", w.Name)
	db := database.New(config.Application.Database, false)
	// TODO: Добавить запуск демонов с контролем сигналов и превышения памяти
	for {
		timeStart := time.Now()
		tr := transport.New(imp.Params)
		data, errorData := tr.Client.GetEntities(w.Queue)
		result := resultProcess{Queue: w.Queue}
		for errorData == nil && len(data.Data) > 0 {
			result.PackageID = data.PackageID
			_, err := w.BeforeIteration(data.Data)
			if err != nil {
				config.Logger.Error().Err(err).Msg("")
			}
			if data != nil {
				result.Total = len(data.Data)
				for _, item := range data.Data {
					err = mapstructure.Decode(item, &w.Job)
					if err != nil {
						config.Logger.Error().Err(err).Msg("")
						return
					}
					_, err := w.Job.Save(db)
					if err != nil {
						result.ErrorItems = append(result.ErrorItems, item)
					}
				}
			}
			result.Duration = time.Now().Sub(timeStart)
			err = imp.Confirm(w, result)
			if err != nil {
				config.Logger.Error().Err(err).Msg("")
			}
			_, err = w.AfterIteration(result.ErrorItems)
			if err != nil {
				config.Logger.Error().Err(err).Msg("")
			}
			timeStart = time.Now()
			data, errorData = tr.Client.GetEntities(w.Queue)
		}

		if errorData != nil || len(data.Data) == 0 {
			result = resultProcess{Queue: w.Queue}
			result.Duration = time.Now().Sub(timeStart)
			err := imp.Confirm(w, result)
			if err != nil {
				config.Logger.Error().Err(err).Msg("")
			}
		}
		time.Sleep(time.Duration(w.Sleep) * time.Second)
	}
}

func (imp *Import) Confirm(w *workers.Worker, data resultProcess) (err error) {
	errorCount := len(data.ErrorItems)
	if data.PackageID > 0 {
		if errorCount == 0 || (data.Total > errorCount && w.AddToQueue(imp.Params, data.ErrorItems)) {
			tr := transport.New(imp.Params)
			tr.Client.ConfirmPackage(data.Queue, data.PackageID)
		}
	}
	logData := resultLog{}
	logData.Queue = data.Queue
	logData.Total = data.Total
	logData.Duration = data.Duration.Round(time.Second).Seconds()
	logData.DurationFormatted = config.FmtDuration(data.Duration)
	logData.Errors = errorCount
	logData.Imported = logData.Total - logData.Errors
	message, err := json.MarshalIndent(logData, "", "    ")
	if err != nil {
		return
	}
	config.Logger.Info().
		Dict("message_json", zerolog.Dict().
			Str("queue", logData.Queue).
			Float64("duration", logData.Duration).
			Str("duration_formatted", logData.DurationFormatted).
			Int("total", logData.Total).
			Int("imported", logData.Imported).
			Int("errors", logData.Errors),
		).
		Msg(fmt.Sprintf("%s", string(message)))
	return
}
