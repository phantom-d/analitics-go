package database

import (
	"analitics/pkg/config"
	"database/sql"
	"github.com/mitchellh/mapstructure"
)

type Datastore struct {
	config  Connection
	connect *sql.DB
}

type Connection interface {
	Connect() *sql.DB
	MigrateUp(source string)
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

func Migrate() {
	if config.Application.MigrateUp {
		New(config.Application.Database, false).
			config.
			MigrateUp("file://migrations")
	}
}

func (ds *Datastore) Close() *Datastore {
	err := ds.connect.Close()
	if err != nil {
		config.Logger.Error().Err(err).Msg("Close connection")
	}
	return ds
}

func (ds *Datastore) Connect() (result *sql.DB) {
	var reconnect bool
	if ds.connect == nil {
		reconnect = true
	} else if err := ds.connect.Ping(); err != nil {
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
		err := ds.connect.Close()
		if err != nil {
			config.Logger.Error().Err(err).Msg("Reconnect")
		}
	}
	ds.connect = ds.config.Connect()
	return ds
}
