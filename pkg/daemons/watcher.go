package daemons

func (d *Daemon) WatcherRun() {
	for _, cfg := range d.Workers {
		daemon := New(cfg.Name)
		if daemon != nil {
			daemon.Run()
		}
	}
}
