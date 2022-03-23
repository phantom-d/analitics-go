package transport

import (
	"analitics/pkg/config"
)

type Package struct {
	PackageID int64                    `json:"package_id,omitempty"`
	Data      []map[string]interface{} `json:"data,omitempty"`
	Name      string                   `json:"name,omitempty"`
	Message   *string                  `json:"message,omitempty"`
}

type ErrorItems struct {
	Header struct {
		Source       string `json:"source"`
		SourceDetail string `json:"source_detail"`
		Date         string `json:"date"`
		Sign         string `json:"sign"`
	} `json:"header"`
	Data ErrorItemsData `json:"data"`
}

type ErrorItemsData struct {
	Queue string   `json:"queue"`
	Guids []string `json:"guids"`
}

type Confirm struct {
	PackageID int64  `json:"package_id"`
	Type      string `json:"type"`
}

type Transport interface {
	GetEntities(string) (*Package, error)
	ConfirmPackage(string, int64)
	ResendErrorItems(string, []string) bool
}

func New(cfg map[string]interface{}) (result Transport) {
	if cfg == nil {
		config.Log().Error().Msg("Not defined configuration for http client!")
		return
	}
	if client, ok := cfg["client"].(map[string]interface{}); ok {
		result = NewClient(client)
	} else {
		config.Log().Error().Msg("Incorrect configuration for http client!")
		return
	}
	return
}
