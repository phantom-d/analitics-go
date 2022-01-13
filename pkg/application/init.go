package application

import (
	"analitics/pkg/datastore"
	"analitics/pkg/transport"

	goflag "flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fatih/structs"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type Application struct {
	ConfigPath string
	Daemon     string
	Debug      bool
	MigrateUp  bool
	Database   struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Name     string `yaml:"name"`
		User     string `yaml:"user"`
		Pass     string `yaml:"pass"`
		CertPath string `yaml:"caCert"`
	}
	Daemons map[string]Daemon
}

type Daemon struct {
	MemoryLimit int64                `yaml:"memory-limit"`
	Sleep       int64                `yaml:"sleep"`
	Daemons     map[string]Gorutines `yaml:"daemons"`
	Params      map[string]Params    `yaml:"params"`
}

type Gorutines struct {
	Enabled bool  `yaml:"enabled"`
	Sleep   int64 `yaml:"sleep"`
}

type Params struct {
	Host string `yaml:"host"`
	User string `yaml:"username"`
	Pass string `yaml:"password"`
}

func New() *Application {
	app := &Application{}
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.StringVarP(&app.ConfigPath, "config", "c", "config.yaml", "Path to a config file")
	flag.StringVarP(&app.Daemon, "daemon", "d", "watcher", "Daemon name to starting")
	flag.BoolVar(&app.MigrateUp, "migrate", false, "Start with migrate up")
	flag.BoolVar(&app.Debug, "debug", false, "Enable debug mode")
	flag.Parse()
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
	return app

}

func (app *Application) Run() {
	if app.MigrateUp {
		db := datastore.New(structs.Map(app))
		err := datastore.MigrateUp(db)
		if err != nil {
			panic(err)
		}
	}

	if app.Daemon == "watcher" {
		db := datastore.New(structs.Map(app))
		server := transport.NewServer(db)
		fmt.Println("server is starting...")
		err := server.Start()
		if err != nil {
			log.Error().Err(err).Msg("Server hasn't been started.")
			os.Exit(1)
		}
	}
}
