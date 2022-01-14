package daemons

type Daemon struct {
	Config map[string]string
}

func New(cfg map[string]string) *Daemon {
	return nil
}

func (d Daemon) GroupHandlers() {

}
