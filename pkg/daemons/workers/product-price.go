package workers

import (
	"analitics/pkg/config"
	"analitics/pkg/database"
	"database/sql"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"time"
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

type productPrice struct {
	Product   string `mapstructure:"product"`
	Price     int64  `mapstructure:"price"`
	PriceType string `mapstructure:"price_type"`
	PriceDate string `mapstructure:"price_date"`
	PriceTime string `mapstructure:"price_time"`
}

func (pp *ProductPrices) Save() (result interface{}, err error) {
	inserts := make([][]interface{}, 0, len(pp.Prices))
	for _, price := range pp.Prices {
		insert := make([]interface{}, 0, 5)
		if pp.checkExist(price) {
			continue
		}
		timeSrc := time.Unix(price.LastUpdate, 0)
		date := fmt.Sprintf("%d-%02d-%02d", timeSrc.Year(), timeSrc.Month(), timeSrc.Day())
		dateTime := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			timeSrc.Year(), timeSrc.Month(), timeSrc.Day(),
			timeSrc.Hour(), timeSrc.Minute(), timeSrc.Second())
		insert = append(insert, pp.ProductGuid, price.Value, price.PriceGuid, date, dateTime)
		inserts = append(inserts, insert)
	}
	if len(inserts) == 0 {
		return
	}
	db := database.Storage.Connect()
	query := `INSERT INTO product_price VALUES (?,?,?,?,?)`
	tx, err := db.Begin()
	if err != nil {
		config.Logger.Error().Err(err).Msg("begin transaction")
		return
	}
	txOK := false
	defer func() {
		if !txOK {
			err = tx.Rollback()
		}
	}()
	stmt, err := tx.Prepare(query)
	for _, insert := range inserts {
		_, err = stmt.Exec(insert...)
		if err != nil {
			config.Logger.Error().Err(err).Msg("loading COPY data")
			return
		}
	}

	err = stmt.Close()
	if err != nil {
		config.Logger.Error().Err(err).Msg("close COPY stmt")
		return
	}

	err = tx.Commit()
	if err != nil {
		config.Logger.Error().Err(err).Msg("commit transaction")
		return
	}
	txOK = true
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
	db := database.Storage.Connect()
	query := `SELECT * FROM product_price WHERE product = ? AND price_type = ? ORDER BY price_time DESC LIMIT 1`
	stmt, err := db.Prepare(query)
	if err != nil {
		config.Logger.Error().Err(err).Msg("")
	}
	defer stmt.Close()

	item := &productPrice{}
	err = stmt.QueryRow(pp.ProductGuid, price.PriceGuid).
		Scan(&item.Product, &item.Price, &item.PriceType, &item.PriceDate, &item.PriceTime)
	if err != nil && err != sql.ErrNoRows {
		config.Logger.Error().Err(err).Msg("")
	}
	if item.Product != "" {
		timeSrc := time.Unix(price.LastUpdate, 0)
		timeTo, _ := time.Parse(time.RFC3339, item.PriceTime)
		if timeTo.Unix() <= timeSrc.Unix() {
			result = true
		}
		if price.Value == item.Price {
			result = true
		}
	}
	return
}
