package main

import (
	"analitics/pkg/datastore"
	"analitics/pkg/exchange/daemons"
	"analitics/pkg/logger"
	"database/sql"
	goflag "flag"
	"github.com/fatih/structs"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
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

func main() {
	app := New(&Application{})
	app.Run()
}

func New(app *Application) *Application {
	app.GetConfig()
	app.logger = logger.New(app.Debug)
	return app
}

func (app *Application) GetConfig() *Application {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.StringVarP(&app.ConfigPath, "config", "c", "config.yaml", "Path to a config file")
	flag.StringVarP(&app.Daemon, "daemon", "d", "watcher", "Daemon name to starting")
	flag.BoolVar(&app.MigrateUp, "migrate", false, "Start with migrate up")
	flag.BoolVar(&app.Debug, "debug", false, "Enable debug mode")
	flag.Parse()

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
		db := datastore.New(structs.Map(app))
		err := datastore.MigrateUp(db)
		if err != nil {
			if err.Error() == "no change" {
				app.logger.Logger().Info().Msgf("Migration: %s!", err.Error())
			} else {
				app.logger.Logger().Error().Msgf("Migration error: %s!", err.Error())
			}
		}
	}
	daemon := daemons.New(app.Daemon, structs.Map(app.Daemons[app.Daemon]), app.Debug)
	if daemon != nil {
		daemon.Run(structs.Map(app))
	}
}
