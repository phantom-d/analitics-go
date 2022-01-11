package main

import (
	goflag "flag"
	flag "github.com/spf13/pflag"

	"analitics-go/pkg/datastore"
	//"analitics-go/transport"
	//"github.com/rs/zerolog/log"
)

type Config struct {
	configPath string
	debug      bool
	migrateUp  bool
}

var Cfg = Config{}

func init() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	flag.StringVarP(&Cfg.configPath, "config", "c", "config.yaml", "Path to a config file")
	flag.BoolVar(&Cfg.migrateUp, "migrate", false, "Start with migrate up")
	flag.BoolVar(&Cfg.debug, "debug", false, "Enable debug mode")

	flag.Parse()
}

func main() {
	db := datastore.InitDB(Cfg.debug)
	if Cfg.migrateUp {
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
