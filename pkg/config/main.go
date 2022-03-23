package config

import (
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

func App() *Config {
	return application
}

func Log() *zerolog.Logger {
	return &logger
}

func GetConfig() *Config {
	content, err := ioutil.ReadFile(application.ConfigPath)
	if err != nil {
		Log().Fatal().Err(err).Msg("Not read config file!")
	}
	content = []byte(os.ExpandEnv(string(content)))
	if err := yaml.Unmarshal(content, &application); err != nil {
		panic(err)
	}
	return application
}

func GetLogger() zerolog.Logger {
	logLevel := zerolog.InfoLevel
	if application.Debug {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	return logger
}
