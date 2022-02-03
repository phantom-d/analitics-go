package daemons

import (
	"analitics/pkg/config"
	"runtime"
	"time"
)

type Watcher struct {
	*DaemonData
}

func (watcher *Watcher) SetData(data *DaemonData) {
	watcher.DaemonData = data
}

func (watcher *Watcher) Run() {
	runtime.GC()
	config.Logger.Info().Msgf("Start daemon '%s'!", watcher.Name)
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	// TODO: Добавить запуск демонов с контролем сигналов
	for memStats.Alloc <= watcher.MemoryLimit {
		for _, cfg := range watcher.Workers {
			daemon := New(cfg.Name)
			if daemon != nil {
				config.Logger.Info().Msgf("Start daemon '%s'!", cfg.Name)
				daemon.Run()
			}
		}
		runtime.GC()
		runtime.ReadMemStats(memStats)
		time.Sleep(time.Duration(watcher.Sleep) * time.Second)
	}
}
