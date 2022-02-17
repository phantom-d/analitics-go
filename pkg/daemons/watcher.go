package daemons

import (
	"analitics/pkg/config"
	"os"
)

type Watcher struct {
	*DaemonData
}

func (watcher *Watcher) SetData(data *DaemonData) {
	watcher.DaemonData = data
}

func (watcher *Watcher) Run() (err error) {
	for _, cfg := range watcher.Workers {
		if daemon := New(cfg.Name); daemon != nil {
			var dm *os.Process
			if dm, err = daemon.Data().Context.Search(); dm == nil {
				if err = Exec(daemon); err != nil {
					config.Logger.Error().Err(err).Msgf("Exec daemon '%s'", cfg.Name)
					err = nil
				}
			}
		}
	}
	return
}
