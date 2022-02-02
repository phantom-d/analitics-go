package daemons

import (
	"analitics/pkg/config"
	"time"
)

type Watcher struct {
	*DaemonData
}

func (watcher *Watcher) SetData(data *DaemonData) {
	watcher.DaemonData = data
}

func (watcher *Watcher) Run() {
	config.Logger.Info().Msgf("Start daemon '%s'!", watcher.Name)
	// TODO: Добавить запуск демонов с контролем сигналов и превышения памяти
	for {
		for _, cfg := range watcher.Workers {
			daemon := New(cfg.Name)
			if daemon != nil {
				config.Logger.Info().Msgf("Start daemon '%s'!", cfg.Name)
				daemon.Run()
			}
		}
		time.Sleep(time.Duration(watcher.Sleep) * time.Second)
	}
}
