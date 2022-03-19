package database

import "analitics/pkg/database/drivers"

type Factory map[string]func() ConnectionInterface

var factory = make(Factory)

func init() {
	factory.Register("clickhouse", func() ConnectionInterface { return &drivers.Clickhouse{} })
}

func (factory *Factory) Register(name string, factoryFunc func() ConnectionInterface) {
	(*factory)[name] = factoryFunc
}

func (factory *Factory) CreateInstance(name string) ConnectionInterface {
	return (*factory)[name]()
}
