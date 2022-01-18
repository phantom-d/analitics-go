package daemons

func (d *Daemon) importHandler(app map[string]interface{}) {
	if threads, ok := d.Config["Threads"].([]interface{}); ok {
		for _, thread := range threads {
			name := thread.(map[string]interface{})["Name"].(string)
			cfg := app["Daemons"].(map[string]interface{})[name].(map[string]interface{})
			daemon := New(name, cfg, d.Debug)
			daemon.Run(app)
		}
	}
}
