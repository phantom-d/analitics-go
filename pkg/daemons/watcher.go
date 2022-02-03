package daemons

import (
	"runtime"
	"time"
)

type Watcher struct {
	*DaemonData
}

func (watcher *Watcher) SetData(data *DaemonData) {
	watcher.DaemonData = data
}

func (watcher *Watcher) Data() *DaemonData {
	return watcher.DaemonData
}

func (watcher *Watcher) Run() {
	runtime.GC()
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	// TODO: Добавить запуск демонов с контролем сигналов
	for memStats.Alloc <= watcher.MemoryLimit {
		for _, cfg := range watcher.Workers {
			daemon := New(cfg.Name)
			if daemon != nil {
				daemon.Data().Start(daemon, true)
			}
		}
		runtime.GC()
		runtime.ReadMemStats(memStats)
		time.Sleep(time.Duration(watcher.Sleep) * time.Second)
	}
}
