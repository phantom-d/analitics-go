package workers

import (
	"analitics/pkg/config"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
)

type ProductPrice struct {
	PriceGuid  string `mapstructure:"price_guid"`
	Value      int64  `mapstructure:"value"`
	LastUpdate int64  `mapstructure:"last_update"`
}

type ProductPrices struct {
	EntityId    int64          `mapstructure:"entity_id"`
	ProductGuid string         `mapstructure:"product_guid"`
	Prices      []ProductPrice `mapstructure:"prices"`
}

func (w *Worker) ProductPriceSave(row map[string]interface{}) (result interface{}, err error) {
	item := ProductPrices{}
	err = mapstructure.Decode(row, &item)
	if err != nil {
		config.Logger.Error().Err(err).Msg("")
		return
	}
	// TODO: Добавить обработку данных очереди
	config.Logger.Info().
		Dict("message_json", zerolog.Dict().
			Str("queue", w.Queue),
		).
		Msg(fmt.Sprintf("%+v", item))
	return
}

func (w *Worker) ProductPriceExtractId(items []map[string]interface{}) (result []string, err error) {
	for _, row := range items {
		item := ProductPrices{}
		err = mapstructure.Decode(row, &item)
		if err != nil {
			config.Logger.Error().Err(err).Msg("")
			return
		}
		result = append(result, item.ProductGuid)
	}
	return
}
