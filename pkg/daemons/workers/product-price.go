package workers

import (
	"analitics/pkg/config"
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

func (pp *ProductPrices) Save() (result interface{}, err error) {
	// TODO: Добавить обработку данных очереди
	for _, price := range pp.Prices {
		if pp.checkExist(price) {

		}
	}
	return
}

func (pp *ProductPrices) ExtractId(items []map[string]interface{}) (result []string, err error) {
	for _, row := range items {
		item := ProductPrices{}
		err = mapstructure.Decode(row, &item)
		if err != nil {
			config.Logger.Error().Err(err).Msg("")
			continue
		}
		result = append(result, item.ProductGuid)
	}
	return
}

func (pp *ProductPrices) checkExist(price ProductPrice) (result bool) {
	return
}
