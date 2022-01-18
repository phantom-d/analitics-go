package daemons

import (
	"analitics/pkg/logger"
	"database/sql"
	"github.com/rs/zerolog"
	"reflect"
	"strings"
)

type Daemon struct {
	Name   string
	Config map[string]interface{}
	Debug  bool
	Params map[string]interface{}
	Sleep  int64
	db     *sql.DB
	logger *logger.Logger
}

func New(name string, cfg map[string]interface{}, debug bool) *Daemon {
	d := &Daemon{Debug: debug}
	d.logger = logger.New(debug)
	if cfg != nil {
		if cfg["Enabled"].(bool) {
			d.Name = name
			d.Config = cfg
			if sleep, ok := cfg["Sleep"].(int64); ok {
				d.Sleep = sleep
			}
			if params, ok := cfg["Params"].(map[string]interface{}); ok {
				d.Params = params
			}
			return d
		} else {
			d.Logger().Info().Msgf("Daemon '%s' is disabled!", name)
		}
	} else {
		d.Logger().Info().Msgf("Daemon '%s' not found!", name)
	}
	return nil
}

func DynamicCall(obj interface{}, fn string, args map[string]interface{}) (res []reflect.Value) {
	method := reflect.ValueOf(obj).MethodByName(fn)
	var inputs []reflect.Value
	for _, v := range args {
		inputs = append(inputs, reflect.ValueOf(v))
	}
	return method.Call(inputs)
}

func (d *Daemon) Logger() *zerolog.Logger {
	return d.logger.Logger()
}

func (d *Daemon) Run(app map[string]interface{}) {
	d.Logger().Info().Msgf("Start daemon '%s'!", d.Name)
	args := make(map[string]interface{}, 0)
	args["arg0"] = app
	DynamicCall(d, strings.Title(strings.ToLower(d.Name))+"Run", args)
}

func (d *Daemon) Fork(app map[string]interface{}) {

}
