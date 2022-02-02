package database

type Factory map[string]func() Connection

var factory = make(Factory)

func init() {
	factory.Register("clickhouse", func() Connection { return &Clickhouse{} })
}

func (factory *Factory) Register(name string, factoryFunc func() Connection) {
	(*factory)[name] = factoryFunc
}

func (factory *Factory) CreateInstance(name string) Connection {
	return (*factory)[name]()
}
