package transport

import (
	"analitics/pkg/config"
)

func New(transportType string, cfg map[string]interface{}) (result TransportInterface) {
	if cfg == nil {
		config.Log().Error().Msg("Not defined configuration for http client!")
		return
	}
	if settings, ok := cfg[transportType].(map[string]interface{}); ok {
		if transport := factory.CreateInstance(transportType + "-" + settings["type"].(string)); transport != nil {
			result = transport.Init(settings)
		}
	} else {
		config.Log().Error().Msg("Incorrect configuration for http client!")
		return
	}
	return
}
