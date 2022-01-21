package daemons

import (
	"analitics/pkg/daemons/workers"
	"analitics/pkg/transport"
	"time"
)

func (d *Daemon) ImportRun() {
	for {
		for _, cfg := range d.Workers {
			worker := workers.New(cfg)
			if worker != nil {
				d.importProcess(worker)
			}
		}
		time.Sleep(time.Duration(d.Sleep) * time.Second)
	}
}

func (d *Daemon) importProcess(w *workers.Worker) {
	tr := transport.New(d.Params)
	for {
		if data := tr.Client.GetEntities(w.Queue); data != nil {
			w.Run(data.Data)
		}
		time.Sleep(time.Duration(w.Sleep) * time.Second)
	}
}
