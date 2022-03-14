package daemons

import (
	"analitics/pkg/config"
	"analitics/pkg/daemons/workers"
	"os"
	"syscall"
)

type Import struct {
	*DaemonData
}

func (imp *Import) SetData(data *DaemonData) {
	imp.DaemonData = data
}

func (imp *Import) Run() (err error) {
	for _, cfg := range imp.Workers {
		if worker := workers.New(cfg, imp.Name, imp.Params); worker != nil {
			if config.Application.Worker == "" || config.Application.Worker == worker.Name {
				var dm *os.Process
				dm, err = worker.Context.Search()
				if err != nil {
					config.Logger.Error().Err(err).Msgf("Exec worker '%s'", cfg.Name)
				} else if dm != nil {
					err = dm.Signal(syscall.Signal(0))
					if err == os.ErrProcessDone {
						dm = nil
					}
				}
				if dm == nil {
					if config.Application.Worker == worker.Name {
						if err = worker.Run(); err != nil {
							config.Logger.Error().Err(err).Msgf("Start worker '%s'", cfg.Name)
							err = nil
						}
						break
					} else {
						if err = workers.Exec(worker); err != nil {
							config.Logger.Error().Err(err).Msgf("Exec worker '%s'", cfg.Name)
							err = nil
						}
					}
				}
			}
		}
	}
	return
}

func (imp *Import) Terminate(s os.Signal) {
	for _, cfg := range imp.Workers {
		if worker := workers.New(cfg, imp.Name, imp.Params); worker != nil {
			dm, err := worker.Context.Search()
			config.Logger.Debug().Msgf("Terminate worker dm: '%+v'", dm)
			config.Logger.Debug().Msgf("Terminate worker Context: '%+v'", worker.Context)
			if err != nil {
				config.Logger.Error().Err(err).Msgf("Terminate worker '%s'", cfg.Name)
			} else {
				if err := dm.Signal(s); err != nil {
					config.Logger.Error().Err(err).Msgf("Terminate worker '%s'", cfg.Name)
				}
				if _, err = dm.Wait(); err != nil {
					config.Logger.Error().Err(err).Msgf("Wait process worker '%s'", cfg.Name)
				}
			}
		}
	}
	err := imp.Context.Release()
	if err != nil {
		config.Logger.Error().Err(err).Msgf("Worker '%s' terminate", imp.Name)
	}
}
