package imports

import (
	"analitics/pkg/config"
	"analitics/pkg/database"

	"github.com/mitchellh/mapstructure"
	"github.com/phantom-d/go-daemons/imports"

	"database/sql"
	"fmt"
	"time"
)

type ProductPrice struct {
	PriceGuid  string `mapstructure:"price_guid"`
	Value      int64  `mapstructure:"value"`
	LastUpdate int64  `mapstructure:"last_update"`
}

type ProductPrices struct {
	*imports.Worker
}

type ProductPricesData struct {
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

func (pp *ProductPrices) SetData(data *imports.Worker) {
	pp.Worker = data
}

func (pp ProductPrices) GetEntities() (result interface{}, err error) {
	return
}

func (pp ProductPrices) Processing(data interface{}, result *imports.ResultProcess) (err error) {
	return
}

func (pp *ProductPrices) Save(ds *database.Datastore, data map[string]interface{}) (result interface{}, err error) {
	item := &ProductPricesData{}
	err = mapstructure.Decode(data, &item)
	if err != nil {
		config.Log().Error().Err(err).Msgf("Worker '%s' processing", pp.Name)
		return
	}
	inserts := make([][]interface{}, 0, len(item.Prices))
	for _, price := range item.Prices {
		insert := make([]interface{}, 0, 5)
		//if item.checkExist(price, ds) {
		//	continue
		//}
		timeSrc := time.Unix(price.LastUpdate, 0)
		date := fmt.Sprintf("%d-%02d-%02d", timeSrc.Year(), timeSrc.Month(), timeSrc.Day())
		dateTime := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
			timeSrc.Year(), timeSrc.Month(), timeSrc.Day(),
			timeSrc.Hour(), timeSrc.Minute(), timeSrc.Second())
		insert = append(insert, item.ProductGuid, price.Value, price.PriceGuid, date, dateTime)
		inserts = append(inserts, insert)
	}
	if len(inserts) == 0 {
		return
	}
	query := `INSERT INTO product_price VALUES (?,?,?,?,?)`
	tx, err := ds.Connect().Begin()
	if err != nil {
		config.Log().Error().Err(err).Msg("begin transaction")
		return
	}
	txOK := false
	defer func() {
		if !txOK {
			err = tx.Rollback()
		}
	}()
	stmt, err := tx.Prepare(query)
	if err != nil {
		config.Log().Error().Err(err).Msg("Prepare stmt")
		return
	}
	for _, insert := range inserts {
		_, err = stmt.Exec(insert...)
		if err != nil {
			config.Log().Error().Err(err).Msg("Loading data")
			return
		}
	}

	err = stmt.Close()
	if err != nil {
		config.Log().Error().Err(err).Msg("Close stmt")
		return
	}

	err = tx.Commit()
	if err != nil {
		config.Log().Error().Err(err).Msg("commit transaction")
		return
	}
	txOK = true
	return
}

func (pp *ProductPrices) ExtractId(items interface{}) (result []string, err error) {
	for _, row := range items.([]map[string]interface{}) {
		item := ProductPricesData{}
		err = mapstructure.Decode(row, &item)
		if err != nil {
			config.Log().Error().Err(err).Msg("")
			continue
		}
		result = append(result, item.ProductGuid)
	}
	return
}

func (pd *ProductPricesData) checkExist(price ProductPrice, ds *database.Datastore) (result bool) {
	config.Log().Debug().Msgf("ProductPrices.checkExist: %+v", price)
	query := `SELECT * FROM product_price WHERE product = ? AND price_type = ? ORDER BY price_time DESC LIMIT 1`
	item := &productPrice{}
	rows, err := ds.Connect().Query(query, pd.ProductGuid, price.PriceGuid)
	if err != nil {
		if err != sql.ErrNoRows {
			config.Log().Error().Err(err).Msg("ProductPrices.checkExist")
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&item.Product, &item.Price, &item.PriceType, &item.PriceDate, &item.PriceTime); err != nil {
			config.Log().Error().Err(err).Msg("ProductPrices.checkExist")
		}
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
	ds.Close()
	return
}
