package daemons

import (
	"analitics/pkg/config"
	"analitics/pkg/daemons/workers"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"
)

func New(name string) Daemon {
	if cfg, ok := config.Application.Daemons[name]; ok {
		if cfg.Enabled {
			cfg.Name = name
			d := factory.CreateInstance(name)
			dd := &DaemonData{}
			err := mapstructure.Decode(cfg, &dd)
			if err != nil {
				config.Logger.Error().Err(err).Msgf("Init daemon '%s'", name)
				return nil
			}
			pidFileName, err := filepath.Abs(fmt.Sprintf("%s/%s.pid", config.Application.PidDir, dd.Name))
			if err != nil {
				config.Logger.Fatal().Err(err).Msgf("Init daemon '%s'", name)
			}
			var args []string
			notExists := true
			daemonArg := "--daemon=" + dd.Name

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
			dd.Context = &config.Context{
				Name:        name,
				Type:        `daemon`,
				PidFileName: pidFileName,
				PidFilePerm: 0644,
				WorkDir:     "./",
				Args:        args,
			}
			d.SetData(dd)
			return d
		} else {
			//config.Logger.Debug().Msgf("Daemon '%s' is disabled!", name)
		}
	} else {
		config.Logger.Info().Msgf("Daemon '%s' not found!", name)
	}
	return nil
}

// Start daemon
func Start(d Daemon) (err error) {
	var (
		cancel context.CancelFunc
	)
	dd := *d.Data()
	config.Logger.Info().Msgf("Start daemon '%s'!", dd.Name)
	err = dd.Context.CreatePidFile()
	if err != nil {
		return
	}
	dd.ctx, cancel = context.WithCancel(context.Background())
	dd.signalChan = make(chan os.Signal, 1)
	signal.Notify(dd.signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	defer func() {
		signal.Stop(dd.signalChan)
		cancel()
	}()

	go func() {
		for {
			select {
			case s := <-dd.signalChan:
				switch s {
				case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
					config.Logger.Info().Msgf("daemon '%s' terminated", dd.Name)
					d.Terminate(s)
					cancel()
					os.Exit(1)
				}
			case <-dd.ctx.Done():
				config.Logger.Info().Msgf("daemon '%s' is done", dd.Name)
				os.Exit(1)
			}
		}
	}()
	for {
		select {
		case <-dd.ctx.Done():
			return
		case <-time.Tick(dd.Sleep):
			if err = d.Run(); err != nil {
				return
			}
		}
	}
}

// Execute daemon as a new system process
func Exec(d Daemon) (err error) {
	_, err = d.Data().Context.Run()
	return
}

func DaemonsStatus(name string) (result []byte, err error) {
	daemonsStatus := make(map[string]DaemonStatus)
	if name == `` {
		name = `watcher`
	}
	if name != "" {
		if daemon := New(name); daemon != nil {
			if daemon.Data().Name == `watcher` {
				for _, cfg := range daemon.Data().Workers {
					if worker := New(cfg.Name); worker != nil {
						if daemonStatus, err := worker.Data().getWorkersStatus(); err == nil {
							daemonsStatus[cfg.Name] = daemonStatus
						}
					}
				}
			} else {
				if daemonStatus, err := daemon.Data().getWorkersStatus(); err == nil {
					daemonsStatus[name] = daemonStatus
				}
			}
		}
	}
	result, err = json.Marshal(daemonsStatus)
	return
}

func (dd *DaemonData) Data() *DaemonData {
	return dd
}

func (dd *DaemonData) Terminate(s os.Signal) {
	for _, cfg := range dd.Workers {
		if daemon := New(cfg.Name); daemon != nil {
			dm, err := daemon.Data().Context.Search()
			config.Logger.Debug().Msgf("Terminate daemon dm: '%+v'", dm)
			config.Logger.Debug().Msgf("Terminate daemon Context: '%+v'", daemon.Data().Context)
			if err != nil {
				config.Logger.Error().Err(err).Msgf("Terminate daemon '%s'", cfg.Name)
			} else {
				if err := dm.Signal(s); err != nil {
					config.Logger.Error().Err(err).Msgf("Terminate daemon '%s'", cfg.Name)
				}
				if _, err = dm.Wait(); err != nil {
					config.Logger.Error().Err(err).Msgf("Wait process worker '%s'", cfg.Name)
				}
			}
		}
	}
	err := dd.Context.Release()
	if err != nil {
		config.Logger.Error().Err(err).Msgf("Daemon '%s' terminate", dd.Name)
	}
}

func (dd *DaemonData) getWorkersStatus() (result DaemonStatus, err error) {
	result = DaemonStatus{}
	for _, cfg := range dd.Workers {
		if worker := workers.New(cfg, dd.Name, dd.Params); worker != nil {
			result.Count.Total += 1
			if status, err := worker.Context.GetStatus(); err == nil && status {
				result.Count.Current += 1
			}
		}
	}
	return
}
