package database

import (
	"analitics/pkg/config"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"strings"

	mclickhouse "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Datastore struct {
	config  map[string]interface{}
	connect *sql.DB
}

var (
	Storage *Datastore = &Datastore{}
)

func New(cfg map[string]interface{}) *Datastore {
	Storage.config = cfg
	switch strings.ToLower(Storage.config["type"].(string)) {
	case "clickhouse":
		Storage.connect = Storage.ClickhouseConnect()
	}
	return Storage
}

func Reconnect() *Datastore {
	err := Storage.connect.Close()
	if err != nil {
		config.Logger.Error().Err(err).Msg("")
	}
	New(Storage.config)
	return Storage
}

func Migrate() {
	if config.Application.MigrateUp {
		if Storage.connect == nil {
			New(config.Application.Database)
		}
		err := Storage.MigrateUp()
		if err != nil {
			if err.Error() == "no change" {
				config.Logger.Info().Msgf("Migration: %s!", err.Error())
			} else {
				config.Logger.Error().Msgf("Migration error: %s!", err.Error())
			}
		}
	}
}

func (ds *Datastore) MigrateUp() error {
	driver, err := mclickhouse.WithInstance(ds.connect, &mclickhouse.Config{})
	if err != nil {
		config.Logger.Fatal().Err(err).Msg("Migration error!")
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", ds.config["name"].(string), driver)
	if err != nil {
		return err
	}
	return m.Up()
}
