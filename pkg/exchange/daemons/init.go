package daemons

import (
	"analitics/pkg/logger"
	"database/sql"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"reflect"
	"strings"
)

type Daemon struct {
	Name   string
	Config map[string]interface{}
	Params map[string]interface{}
	db     *sql.DB
	logger *logger.Logger
}

func New(app map[string]interface{}) *Daemon {
	d := &Daemon{}
	d.logger = logger.New(app["Debug"].(bool))
	name := app["Daemon"].(string)
	if daemon, ok := app["Daemons"].(map[string]interface{})[name].(map[string]interface{}); ok {
		if daemon["Enabled"].(bool) {
			d.Name = name
			d.Config = daemon
			if params, ok := daemon["Params"].(map[string]interface{}); ok {
				d.Params = params
			}
			return d
		} else {
			log.Info().Msgf("Daemon '%s' is disabled!", name)
		}
	} else {
		log.Info().Msgf("Daemon '%s' not found!", name)
	}
	return nil
}

func (d *Daemon) Run() {
	d.Logger().Info().Msgf("Start daemon '%s'!", d.Name)
	args := make(map[string]interface{}, 0)
	DynamicCall(d, strings.Title(strings.ToLower(d.Name))+"Run", args)
}

func (d *Daemon) Logger() *zerolog.Logger {
	return d.logger.Logger()
}

func DynamicCall(obj interface{}, fn string, args map[string]interface{}) (res []reflect.Value) {
	method := reflect.ValueOf(obj).MethodByName(fn)
	var inputs []reflect.Value
	for _, v := range args {
		inputs = append(inputs, reflect.ValueOf(v))
	}
	return method.Call(inputs)
}
