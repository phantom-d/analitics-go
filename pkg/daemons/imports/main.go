package imports

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

func New(cfg config.Worker, parent string, params map[string]interface{}) WorkerInterface {
	if cfg.Enabled {
		w := factory.CreateInstance(cfg.Name)
		wd := &Worker{Params: params, Parent: parent}
		err := mapstructure.Decode(cfg, &wd)
		if err != nil {
			config.Log().Info().Msg("Worker load config")
			return nil
		}
		pidFileName, err := filepath.Abs(fmt.Sprintf("%s/%s_%s.pid", config.App().PidDir, parent, cfg.Name))
		if err != nil {
			config.Log().Fatal().Err(err).Msgf("Init daemon '%s'", cfg.Name)
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
		wd.Context = &config.Context{
			Name:        cfg.Name,
			Type:        `worker`,
			PidFileName: pidFileName,
			PidFilePerm: 0644,
			WorkDir:     "./",
			Args:        args,
		}
		w.SetData(wd)
		return w
	} else {
		config.Log().Info().Msgf("Worker '%s' is disabled!", cfg.Name)
	}
	return nil
}

func Run(w WorkerInterface) (err error) {
	var cancel context.CancelFunc
	wd := w.Data()
	err = wd.Context.CreatePidFile()
	if err != nil {
		config.Log().Fatal().Err(err).Msgf("Worker '%s' Process", wd.Name)
	}
	wd.ctx, cancel = context.WithCancel(context.Background())
	wd.signalChan = make(chan os.Signal, 1)
	signal.Notify(wd.signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	defer func() {
		signal.Stop(wd.signalChan)
		cancel()
	}()

	go func() {
		for {
			select {
			case s := <-wd.signalChan:
				switch s {
				case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
					config.Log().Info().Msgf("worker '%s' terminate", wd.Name)
					cancel()
					wd.Terminate(s)
					err := wd.Context.Release()
					if err != nil {
						config.Log().Error().Err(err).Msgf("Worker '%s' terminate", wd.Name)
					}
					os.Exit(1)
				}
			case <-wd.ctx.Done():
				config.Log().Info().Msgf("worker '%s' is done", wd.Name)
				os.Exit(1)
			}
		}
	}()

	config.Log().Info().Msgf("Start worker '%s'!", wd.Name)
	db := database.New(config.App().Database, false)
	for {
		select {
		case <-wd.ctx.Done():
			return
		case <-time.Tick(wd.Sleep):
			runtime.GC()
			memStats := &runtime.MemStats{}
			runtime.ReadMemStats(memStats)
			_, err = wd.BeforeRun()
			if err != nil {
				config.Log().Error().Err(err).Msgf("Worker '%s' processing BeforeRun", wd.Name)
			}
			if memStats.Alloc > wd.MemoryLimit {
				break
			}
			timeStart := time.Now()
			tr := transport.New("client", wd.Params)
			data, errorData := tr.GetEntities(wd.Queue)
			result := resultProcess{Queue: wd.Queue}
			runtime.ReadMemStats(memStats)
			for errorData == nil && len(data.Data) > 0 {
				if memStats.Alloc > wd.MemoryLimit {
					break
				}
				result.PackageID = data.PackageID
				_, err := wd.BeforeIteration(data.Data)
				if err != nil {
					config.Log().Error().Err(err).Msgf("Worker '%s' processing BeforeIteration", wd.Name)
				}
				if data != nil {
					result.Total = len(data.Data)
					for _, item := range data.Data {
						_, err := w.Save(db, item)
						if err != nil {
							result.ErrorItems = append(result.ErrorItems, item)
						}
					}
				}
				result.Duration = time.Now().Sub(timeStart)
				runtime.ReadMemStats(memStats)
				result.Memory = memStats.Alloc
				err = Confirm(w, result)
				if err != nil {
					config.Log().Error().Err(err).Msgf("Worker '%s' processing Confirm", wd.Name)
				}
				_, err = wd.AfterIteration(result.ErrorItems)
				if err != nil {
					config.Log().Error().Err(err).Msgf("Worker '%s' processing AfterIteration", wd.Name)
				}
				timeStart = time.Now()
				runtime.GC()
				runtime.ReadMemStats(memStats)
				data, errorData = tr.GetEntities(wd.Queue)
			}

			if errorData != nil || len(data.Data) == 0 {
				result = resultProcess{Queue: wd.Queue}
				result.Duration = time.Now().Sub(timeStart)
				runtime.ReadMemStats(memStats)
				result.Memory = memStats.Alloc
				err := Confirm(w, result)
				if err != nil {
					config.Log().Error().Err(err).Msgf("Worker '%s' processing zero Confirm", wd.Name)
				}
			}
			runtime.GC()
			runtime.ReadMemStats(memStats)
			_, err = wd.AfterRun()
			if err != nil {
				config.Log().Error().Err(err).Msgf("Worker '%s' processing AfterRun", wd.Name)
			}
		}
	}
}

func AddToQueue(w WorkerInterface, errorItems []map[string]interface{}) bool {
	result := true
	wd := w.Data()
	items, _ := w.ExtractId(errorItems)

	if items != nil {
		tr := transport.New("client", wd.Params)
		result = tr.ResendErrorItems(wd.Queue, items)
	}

	return result
}

func Confirm(w WorkerInterface, data resultProcess) (err error) {
	errorCount := len(data.ErrorItems)
	wd := w.Data()
	if data.PackageID > 0 {
		if errorCount == 0 || (data.Total > errorCount && AddToQueue(w, data.ErrorItems)) {
			tr := transport.New("client", wd.Params)
			tr.ConfirmPackage(data.Queue, data.PackageID)
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

	config.Log().Info().
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
func (w *Worker) Run() (err error) {
	_, err = w.Context.Run()
	return
}

func (w *Worker) Data() *Worker {
	return w
}

func (w *Worker) GetStatus() (result bool, err error) {
	return w.Context.GetStatus()
}

func (w *Worker) Terminate(s os.Signal) {
}

func (w *Worker) BeforeRun() (result interface{}, err error) {
	return
}

func (w *Worker) AfterRun() (result interface{}, err error) {
	return
}

func (w *Worker) BeforeIteration(data []map[string]interface{}) (result interface{}, err error) {
	return
}

func (w *Worker) AfterIteration(errorItems []map[string]interface{}) (result interface{}, err error) {
	return
}
