package transport

import (
	"analitics/pkg/config"
)

type Transport struct {
	Client *Client
}

func New(cfg map[string]interface{}) *Transport {
	t := &Transport{}
	if cfg == nil {
		config.Log().Error().Msg("Not defined configuration for http client!")
		return nil
	}
	if client, ok := cfg["client"].(map[string]interface{}); ok {
		t.Client = NewClient(client)
	} else {
		config.Log().Error().Msg("Incorrect configuration for http client!")
		return nil
	}
	return t
}
