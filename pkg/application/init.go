package application

import (
	"analitics/pkg/logger"
	"database/sql"
	"github.com/fatih/structs"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"

	"analitics/pkg/datastore"
	"analitics/pkg/exchange/daemons"
)

type Application struct {
	ConfigPath string
	Daemon     string
	Debug      bool
	MigrateUp  bool
	Database   map[string]interface{} `yaml:"database"`
	Daemons    map[string]Daemon
	db         *sql.DB
	logger     *logger.Logger
}

type Daemon struct {
	Enabled     bool                   `yaml:"enabled"`
	MemoryLimit int64                  `yaml:"memory-limit"`
	Sleep       int64                  `yaml:"sleep"`
	Threads     []Threads              `yaml:"threads"`
	Params      map[string]interface{} `yaml:"params"`
}

type Threads struct {
	Name    string `yaml:"name"`
	Queue   string `yaml:"queue"`
	Enabled bool   `yaml:"enabled"`
	Sleep   int64  `yaml:"sleep"`
}

func New(app *Application) *Application {
	app.GetConfig()
	app.logger = logger.New(app.Debug)
	return app
}

func (app *Application) GetConfig() *Application {
	content, err := ioutil.ReadFile(app.ConfigPath)
	if err != nil {
		panic(err)
	}
	content = []byte(os.ExpandEnv(string(content)))
	if err := yaml.Unmarshal(content, &app); err != nil {
		panic(err)
	}
	return app

}

func (app *Application) Run() {
	if app.MigrateUp {
		err := datastore.MigrateUp(app.db)
		if err != nil {
			panic(err)
		}
	}
	daemon := daemons.New(structs.Map(app))
	if daemon != nil {
		daemon.Run()
	}
}

func (app *Application) Logger() *zerolog.Logger {
	return app.logger.Logger()
}
