package transport

import (
	"analitics/pkg/logger"
	"github.com/rs/zerolog"
)

type Transport struct {
	Client *Client
}

var transportLogger *logger.Logger

func New(cfg map[string]interface{}) *Transport {
	t := &Transport{}
	if cfg == nil {
		Logger().Error().Msg("Not defined configuration for http client!")
		return nil
	}
	if client, ok := cfg["Client"].(map[string]interface{}); ok {
		t.Client = NewClient(client)
	} else {
		Logger().Error().Msg("Incorrect configuration for http client!")
		return nil
	}
	return t
}

func Logger() *zerolog.Logger {
	if transportLogger == nil {
		transportLogger = logger.New(true)
	}
	return transportLogger.Logger()
}
