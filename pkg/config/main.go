package config

import (
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

func GetConfig() *Config {
	content, err := ioutil.ReadFile(Application.ConfigPath)
	if err != nil {
		Logger.Fatal().Err(err).Msg("Not read config file!")
	}
	content = []byte(os.ExpandEnv(string(content)))
	if err := yaml.Unmarshal(content, &Application); err != nil {
		panic(err)
	}
	return Application
}

func GetLogger() zerolog.Logger {
	logLevel := zerolog.InfoLevel
	if Application.Debug {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	return Logger
}
