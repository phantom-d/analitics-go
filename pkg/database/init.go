package database

import "analitics/pkg/database/drivers"

type Factory map[string]func() Connection

var factory = make(Factory)

func init() {
	factory.Register("clickhouse", func() Connection { return &drivers.Clickhouse{} })
}

func (factory *Factory) Register(name string, factoryFunc func() Connection) {
	(*factory)[name] = factoryFunc
}

func (factory *Factory) CreateInstance(name string) (result Connection) {
	if factoryFunc, ok := (*factory)[name]; ok {
		result = factoryFunc()
	}
	return
}
