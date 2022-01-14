package exchange

import (
	"analitics/pkg/application"
	"reflect"
)

type Exchange struct {
	Config application.Daemon
}

func New(name string, app *application.Application) *Exchange {
	cfg := app.GetConfig()
	exch := &Exchange{}

	if daemon, found := cfg.Daemons[name]; found {
		exch.Config = daemon
	}
	return exch
}

func (exch *Exchange) Start() {

}

func DynamicCall(obj interface{}, fn string, args map[string]interface{}) (res []reflect.Value) {
	method := reflect.ValueOf(obj).MethodByName(fn)
	var inputs []reflect.Value
	for _, v := range args {
		inputs = append(inputs, reflect.ValueOf(v))
	}
	return method.Call(inputs)
}
