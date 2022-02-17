package daemons

import (
	"analitics/pkg/config"
	"context"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"os"
	"os/signal"
	"runtime"
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
				config.Logger.Error().Err(err).Msg("")
				return nil
			}
			dd.Context = &Context{
				PidFileName: fmt.Sprintf("%s/%s.pid", config.Application.PidDir, dd.Name),
				PidFilePerm: 0644,
				WorkDir:     "./",
				Args:        []string{"[--daemon=" + dd.Name + "]"},
			}
			d.SetData(dd)
			return d
		} else {
			config.Logger.Debug().Msgf("Daemon '%s' is disabled!", name)
		}
	} else {
		config.Logger.Info().Msgf("Daemon '%s' not found!", name)
	}
	return nil
}

// Start daemon
func Start(d Daemon) (err error) {
	var memStats *runtime.MemStats
	dd := d.Data()
	config.Logger.Info().Msgf("Start daemon '%s'!", dd.Name)
	d.MakeDaemon()
	for {
		select {
		case <-dd.ctx.Done():
			return
		case <-time.Tick(dd.Sleep):
			runtime.GC()
			runtime.ReadMemStats(memStats)
			for memStats.Alloc <= dd.MemoryLimit {
				if err = d.Run(); err != nil {
					return
				}
			}
		}
	}
}

// Execute daemon as a new system process
func Exec(d Daemon) (err error) {
	_, err = d.Data().Context.Run()
	return
}

func (dd *DaemonData) Data() *DaemonData {
	return dd
}

func (dd *DaemonData) MakeDaemon() {
	var cancel context.CancelFunc
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
					cancel()
					dd.Terminate(s)
					os.Exit(1)
				}
			case <-dd.ctx.Done():
				config.Logger.Info().Msgf("daemon '%s' is done", dd.Name)
				os.Exit(1)
			}
		}
	}()
}

func (dd *DaemonData) Terminate(s os.Signal) {
	for _, cfg := range dd.Workers {
		if daemon := New(cfg.Name); daemon != nil {
			if dm, _ := daemon.Data().Context.Search(); dm != nil {
				if err := dm.Signal(s); err != nil {
					config.Logger.Error().Err(err).Msgf("Terminate daemon '%s'", dd.Name)
				}
			}
		}
	}
}
