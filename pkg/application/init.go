package application

import (
	"analitics/pkg/datastore"
	"analitics/pkg/transport"
	"database/sql"
	"github.com/fatih/structs"

	"fmt"
	"io/ioutil"
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Application struct {
	ConfigPath string
	Daemon     string
	Debug      bool
	MigrateUp  bool
	Database   map[string]interface{} `yaml:"database"`
	Daemons    map[string]Daemon
	db         *sql.DB
}

type Daemon struct {
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
	app.db = datastore.New(structs.Map(app))
	return app

}

func (app *Application) Run() {
	if app.MigrateUp {
		err := datastore.MigrateUp(app.db)
		if err != nil {
			panic(err)
		}
	}

	if app.Daemon == "watcher" {
		server := transport.NewServer(app.db)
		fmt.Println("server is starting...")
		err := server.Start()
		if err != nil {
			log.Error().Err(err).Msg("Server hasn't been started.")
			os.Exit(1)
		}
	}
}
