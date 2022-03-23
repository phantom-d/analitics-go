package transport

type Package struct {
	PackageID int64                    `json:"package_id,omitempty"`
	Data      []map[string]interface{} `json:"data,omitempty"`
	Name      string                   `json:"name,omitempty"`
	Message   *string                  `json:"message,omitempty"`
}

type ErrorItems struct {
	Header struct {
		Source       string `json:"source"`
		SourceDetail string `json:"source_detail"`
		Date         string `json:"date"`
		Sign         string `json:"sign"`
	} `json:"header"`
	Data ErrorItemsData `json:"data"`
}

type ErrorItemsData struct {
	Queue string   `json:"queue"`
	Guids []string `json:"guids"`
}

type Confirm struct {
	PackageID int64  `json:"package_id"`
	Type      string `json:"type"`
}

type TransportInterface interface {
	Init(cfg map[string]interface{}) TransportInterface
	GetEntities(string) (*Package, error)
	ConfirmPackage(string, int64)
	ResendErrorItems(string, []string) bool
}

type Factory map[string]func() TransportInterface

var factory = make(Factory)

func init() {
	factory.Register("client-http", func() TransportInterface { return &HttpClient{} })
}

func (factory *Factory) Register(name string, factoryFunc func() TransportInterface) {
	(*factory)[name] = factoryFunc
}

func (factory *Factory) CreateInstance(name string) (result TransportInterface) {
	if factoryFunc, ok := (*factory)[name]; ok {
		result = factoryFunc()
	}
	return
}
