package database

import (
	"analitics/pkg/config"
	"database/sql"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

type Datastore struct {
	config  ConnectionInterface
	connect *gorm.DB
}

type ConnectionInterface interface {
	Connect() *gorm.DB
}

func New(cfg map[string]interface{}, start bool) (result *Datastore) {
	ds := &Datastore{config: factory.CreateInstance(cfg["type"].(string))}
	err := mapstructure.Decode(cfg, &ds.config)
	if err != nil {
		config.Logger.Fatal().Err(err).Msg("Decode config for database connection")
		return
	}
	if start {
		ds.connect = ds.config.Connect()
	}
	result = ds
	return
}

func (ds *Datastore) Close() *Datastore {
	db, err := ds.connect.DB()
	if err != nil {
		config.Logger.Error().Err(err).Msg("Reconnect")
	}
	err = db.Close()
	if err != nil {
		config.Logger.Error().Err(err).Msg("Close connection")
	}
	return ds
}

func (ds *Datastore) Connect() (result *gorm.DB) {
	var (
		reconnect bool
		db        *sql.DB
		err       error
	)
	if ds.connect == nil {
		reconnect = true
	} else {
		db, err = ds.connect.DB()
		if err != nil {
			reconnect = true
		}
	}
	if ds.connect == nil {
		reconnect = true
	} else if err = db.Ping(); err != nil {
		reconnect = true
	}
	if reconnect {
		ds.Reconnect()
	}
	result = ds.connect
	return
}

func (ds *Datastore) Reconnect() *Datastore {
	if ds.connect != nil {
		ds.Close()
	}
	ds.connect = ds.config.Connect()
	return ds
}
