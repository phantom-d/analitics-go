package application

import (
	"analitics-go/pkg/datastore"
	"database/sql"
	goflag "flag"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

type Config struct {
	ConfigPath string
	Debug      bool
	MigrateUp  bool
	Database   struct {
		Host     string `yaml:"host"`
		Port     int64  `yaml:"port"`
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

type Application struct {
	Config Config
	db     *sql.DB
}

func New() Application {
	cfg := Config{}
	app := Application{Config: cfg}
	return app

}

func (cfg Config) GetConfig() Config {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.StringVarP(&cfg.ConfigPath, "config", "c", "config.yaml", "Path to a config file")
	flag.BoolVar(&cfg.MigrateUp, "migrate", false, "Start with migrate up")
	flag.BoolVar(&cfg.Debug, "debug", false, "Enable debug mode")
	flag.Parse()
	content, err := ioutil.ReadFile(cfg.ConfigPath)
	if err != nil {
		panic(err)
	}
	content = []byte(os.ExpandEnv(string(content)))
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		panic(err)
	}
	return cfg
}

func (cfg Config) Run() {
	db := datastore.New(cfg)
	if cfg.MigrateUp {
		datastore.MigrateUp(db)
	}

	//server := transport.NewServer(db)
	//fmt.Println("server is starting...")
	//err := server.Start()
	//if err != nil {
	//	log.Error().Err(err).Msg("Server hasn't been started.")
	//	os.Exit(1)
	//}

}
