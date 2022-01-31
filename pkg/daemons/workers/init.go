package workers

type Factory map[string]func() Job

var factory = make(Factory)

func init() {
	factory.Register("ProductPrices", func() Job { return &ProductPrices{} })
}

func (factory *Factory) Register(name string, factoryFunc func() Job) {
	(*factory)[name] = factoryFunc
}

func (factory *Factory) CreateInstance(name string) Job {
	return (*factory)[name]()
}
