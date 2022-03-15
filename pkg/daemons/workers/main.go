package workers

import (
	"analitics/pkg/config"
	"analitics/pkg/database"
	"analitics/pkg/transport"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"syscall"
	"time"
)

func New(cfg config.Worker, parent string, params map[string]interface{}) *Worker {
	if cfg.Enabled {
		worker := &Worker{Job: factory.CreateInstance(cfg.Name), Params: params, Parent: parent}
		err := mapstructure.Decode(cfg, &worker)
		if err != nil {
			config.Logger.Info().Msg("Worker load config")
			return nil
		}
		pidFileName, err := filepath.Abs(fmt.Sprintf("%s/%s_%s.pid", config.Application.PidDir, parent, cfg.Name))
		if err != nil {
			config.Logger.Fatal().Err(err).Msgf("Init daemon '%s'", cfg.Name)
		}
		var args []string
		notExists := true
		daemonArg := "--daemon=" + parent

		for _, arg := range os.Args {
			if matched, _ := regexp.MatchString(`--migrate`, arg); matched {
				continue
			}
			if matched, _ := regexp.MatchString(`--daemon=`, arg); matched {
				arg = daemonArg
				notExists = false
			}
			args = append(args, arg)
		}
		if notExists {
			args = append(args, daemonArg)
		}
		args = append(args, "--worker="+cfg.Name)
		worker.Context = &config.Context{
			Name:        cfg.Name,
			Type:        `worker`,
			PidFileName: pidFileName,
			PidFilePerm: 0644,
			WorkDir:     "./",
			Args:        args,
		}
		return worker
	} else {
		config.Logger.Info().Msgf("Worker '%s' is disabled!", cfg.Name)
	}
	return nil
}

func (w *Worker) Run() (err error) {
	var (
		cancel context.CancelFunc
	)
	err = w.Context.CreatePidFile()
	if err != nil {
		config.Logger.Fatal().Err(err).Msgf("Worker '%s' Process", w.Name)
	}
	w.ctx, cancel = context.WithCancel(context.Background())
	w.signalChan = make(chan os.Signal, 1)
	signal.Notify(w.signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	defer func() {
		signal.Stop(w.signalChan)
		cancel()
	}()

	go func() {
		for {
			select {
			case s := <-w.signalChan:
				switch s {
				case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
					config.Logger.Info().Msgf("worker '%s' terminated", w.Name)
					cancel()
					w.Terminate(s)
					err := w.Context.Release()
					if err != nil {
						config.Logger.Error().Err(err).Msgf("Worker '%s' terminate", w.Name)
					}
					os.Exit(1)
				}
			case <-w.ctx.Done():
				config.Logger.Info().Msgf("worker '%s' is done", w.Name)
				os.Exit(1)
			}
		}
	}()

	config.Logger.Info().Msgf("Start worker '%s'!", w.Name)
	db := database.New(config.Application.Database, false)
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-time.Tick(w.Sleep):
			runtime.GC()
			memStats := &runtime.MemStats{}
			runtime.ReadMemStats(memStats)
			_, err = w.BeforeRun()
			if err != nil {
				config.Logger.Error().Err(err).Msgf("Worker '%s' processing BeforeRun", w.Name)
			}
			if memStats.Alloc > w.MemoryLimit {
				break
			}
			timeStart := time.Now()
			tr := transport.New(w.Params)
			data, errorData := tr.Client.GetEntities(w.Queue)
			result := resultProcess{Queue: w.Queue}
			runtime.ReadMemStats(memStats)
			for errorData == nil && len(data.Data) > 0 {
				if memStats.Alloc > w.MemoryLimit {
					break
				}
				result.PackageID = data.PackageID
				_, err := w.BeforeIteration(data.Data)
				if err != nil {
					config.Logger.Error().Err(err).Msgf("Worker '%s' processing BeforeIteration", w.Name)
				}
				if data != nil {
					result.Total = len(data.Data)
					for _, item := range data.Data {
						err = mapstructure.Decode(item, &w.Job)
						if err != nil {
							config.Logger.Error().Err(err).Msgf("Worker '%s' processing", w.Name)
							continue
						}
						_, err := w.Job.Save(db)
						if err != nil {
							result.ErrorItems = append(result.ErrorItems, item)
						}
					}
				}
				result.Duration = time.Now().Sub(timeStart)
				runtime.ReadMemStats(memStats)
				result.Memory = memStats.Alloc
				err = w.Confirm(result)
				if err != nil {
					config.Logger.Error().Err(err).Msgf("Worker '%s' processing Confirm", w.Name)
				}
				_, err = w.AfterIteration(result.ErrorItems)
				if err != nil {
					config.Logger.Error().Err(err).Msgf("Worker '%s' processing AfterIteration", w.Name)
				}
				timeStart = time.Now()
				runtime.GC()
				runtime.ReadMemStats(memStats)
				data, errorData = tr.Client.GetEntities(w.Queue)
			}

			if errorData != nil || len(data.Data) == 0 {
				result = resultProcess{Queue: w.Queue}
				result.Duration = time.Now().Sub(timeStart)
				runtime.ReadMemStats(memStats)
				result.Memory = memStats.Alloc
				err := w.Confirm(result)
				if err != nil {
					config.Logger.Error().Err(err).Msgf("Worker '%s' processing zero Confirm", w.Name)
				}
			}
			runtime.GC()
			runtime.ReadMemStats(memStats)
			_, err = w.AfterRun()
			if err != nil {
				config.Logger.Error().Err(err).Msgf("Worker '%s' processing AfterRun", w.Name)
			}
		}
	}
}

func (w *Worker) Confirm(data resultProcess) (err error) {
	errorCount := len(data.ErrorItems)
	if data.PackageID > 0 {
		if errorCount == 0 || (data.Total > errorCount && w.AddToQueue(data.ErrorItems)) {
			tr := transport.New(w.Params)
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
		Dict("context", zerolog.Dict().
			Uint64("memory", data.Memory).
			Str("category", "exchange_import"),
		).
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

// Execute daemon as a new system process
func Exec(w *Worker) (err error) {
	_, err = w.Context.Run()
	return
}

func (w *Worker) Terminate(s os.Signal) {
	err := w.Context.Release()
	if err != nil {
		config.Logger.Error().Err(err).Msgf("Worker '%s' terminate", w.Name)
	}
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

func (w *Worker) AddToQueue(errorItems []map[string]interface{}) bool {
	result := true
	items, _ := w.Job.ExtractId(errorItems)

	if items != nil {
		tr := transport.New(w.Params)
		result = tr.Client.ResendErrorItems(w.Queue, items)
	}

	return result
}
