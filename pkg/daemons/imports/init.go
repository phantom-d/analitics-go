package imports

import (
	"analitics/pkg/config"
	"analitics/pkg/database"
	"context"
	"os"
	"time"
)

type Worker struct {
	Name        string        `mapstructure:"Name"`
	MemoryLimit uint64        `mapstructure:"MemoryLimit"`
	Queue       string        `mapstructure:"Queue"`
	Enabled     bool          `mapstructure:"Enabled"`
	Sleep       time.Duration `mapstructure:"Sleep"`
	Params      map[string]interface{}
	Parent      string
	Context     *config.Context
	ctx         context.Context
	signalChan  chan os.Signal
	done        chan struct{}
}

type WorkerInterface interface {
	BeforeRun() (interface{}, error)
	AfterRun() (interface{}, error)
	BeforeIteration([]map[string]interface{}) (interface{}, error)
	AfterIteration([]map[string]interface{}) (interface{}, error)
	Data() *Worker
	GetStatus() (bool, error)
	Run() error
	SetData(worker *Worker)
	Save(*database.Datastore, map[string]interface{}) (interface{}, error)
	ExtractId([]map[string]interface{}) ([]string, error)
	Terminate(os.Signal)
}

type resultProcess struct {
	PackageID  int64
	Queue      string
	Duration   time.Duration
	Total      int
	Memory     uint64
	ErrorItems []map[string]interface{}
}

type resultLog struct {
	Queue             string  `json:"queue"`
	Duration          float64 `json:"duration"`
	DurationFormatted string  `json:"duration_formatted"`
	Total             int     `json:"total"`
	Imported          int     `json:"imported"`
	Errors            int     `json:"errors"`
}

type Factory map[string]func() WorkerInterface

var factory = make(Factory)

func init() {
	factory.Register("ProductPrices", func() WorkerInterface { return &ProductPrices{} })
}

func (factory *Factory) Register(name string, factoryFunc func() WorkerInterface) {
	(*factory)[name] = factoryFunc
}

func (factory *Factory) CreateInstance(name string) (result WorkerInterface) {
	if factoryFunc, ok := (*factory)[name]; ok {
		result = factoryFunc()
	}
	return
}
