package workers

import (
	"analitics/pkg/config"
	"analitics/pkg/database"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"time"
)

type ProductPriceData struct {
	PriceGuid  string `mapstructure:"price_guid"`
	Value      int64  `mapstructure:"value"`
	LastUpdate int64  `mapstructure:"last_update"`
}

type ProductPrices struct {
	EntityId    int64              `mapstructure:"entity_id"`
	ProductGuid string             `mapstructure:"product_guid"`
	Prices      []ProductPriceData `mapstructure:"prices"`
}

type ProductPrice struct {
	Product   string `gorm:"type:UUID;size:36;index:product_price_product_price_type,type:bloom_filter(0.01),granularity:2"`
	Price     int64  `gorm:"type:UInt32;default:0"`
	PriceType string `gorm:"type:UUID;size:36;index:product_price_product_price_type"`
	PriceDate string `gorm:"type:Date"`
	PriceTime string `gorm:"type:DateTime('Europe/Moscow')"`
}

func (pp *ProductPrices) Migrate(ds *database.Datastore) (err error) {
	err = ds.Connect().Set(
		"gorm:table_options",
		"ENGINE=MergeTree() PARTITION BY toYYYYMM(price_date) ORDER BY (price_date, price_time)",
	).AutoMigrate(&ProductPrice{})
	return
}

func (pp *ProductPrices) Save(ds *database.Datastore) (result interface{}, err error) {
	inserts := make([]map[string]interface{}, 0, len(pp.Prices))
	for _, price := range pp.Prices {
		//if pp.checkExist(price, ds) {
		//	continue
		//}
		timeSrc := time.Unix(price.LastUpdate, 0)
		date := fmt.Sprintf("%d-%02d-%02d", timeSrc.Year(), timeSrc.Month(), timeSrc.Day())
		dateTime := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			timeSrc.Year(), timeSrc.Month(), timeSrc.Day(),
			timeSrc.Hour(), timeSrc.Minute(), timeSrc.Second())
		insert := map[string]interface{}{
			"Product":   pp.ProductGuid,
			"Price":     price.Value,
			"PriceType": price.PriceGuid,
			"PriceDate": date,
			"PriceTime": dateTime,
		}
		inserts = append(inserts, insert)
	}
	if len(inserts) == 0 {
		return
	}

	tx := ds.Connect().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Error; err != nil {
		return
	}

	if err = tx.Model(&ProductPrice{}).Create(inserts).Error; err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit().Error
	if err != nil {
		config.Logger.Error().Err(err).Msg("commit transaction")
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

//func (pp *ProductPrices) checkExist(price ProductPriceData, ds *database.Datastore) (result bool) {
//	config.Logger.Debug().Msgf("ProductPrices.checkExist: %+v", price)
//	query := `SELECT * FROM product_price WHERE product = ? AND price_type = ? ORDER BY price_time DESC LIMIT 1`
//	item := &productPrice{}
//	rows, err := ds.Connect().Query(query, pp.ProductGuid, price.PriceGuid)
//	if err != nil {
//		if err != sql.ErrNoRows {
//			config.Logger.Error().Err(err).Msg("ProductPrices.checkExist")
//		}
//		return
//	}
//	defer rows.Close()
//	for rows.Next() {
//		if err := rows.Scan(&item.Product, &item.Price, &item.PriceType, &item.PriceDate, &item.PriceTime); err != nil {
//			config.Logger.Error().Err(err).Msg("ProductPrices.checkExist")
//		}
//	}
//	if item.Product != "" {
//		timeSrc := time.Unix(price.LastUpdate, 0)
//		timeTo, _ := time.Parse(time.RFC3339, item.PriceTime)
//		if timeTo.Unix() <= timeSrc.Unix() {
//			result = true
//		}
//		if price.Value == item.Price {
//			result = true
//		}
//	}
//	ds.Close()
//	return
//}
