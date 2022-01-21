package workers

import (
	"analitics/pkg/config"
	"fmt"
	"github.com/mitchellh/mapstructure"
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

func (w *Worker) ProductPriceRun(data []map[string]interface{}) {
	for _, row := range data {
		item := ProductPrices{}
		err := mapstructure.Decode(row, &item)
		if err != nil {
			config.Logger.Error().Err(err).Msg("")
			continue
		}
		w.productPriceSave(item)
	}
}

func (w *Worker) productPriceSave(item ProductPrices) {
	fmt.Println(item)
}
