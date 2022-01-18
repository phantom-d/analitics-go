package datastore

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"os"
	"strings"

	mclickhouse "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Datastore struct {
	config  map[string]interface{}
	connect *sql.DB
}

func New(cfg map[string]interface{}) *sql.DB {
	ds := &Datastore{config: cfg["Database"].(map[string]interface{})}
	switch strings.ToLower(ds.config["type"].(string)) {
	case "clickhouse":
		return ds.ClickhouseConnect(cfg["Debug"].(bool))
	}
	return nil
}

func MigrateUp(db *sql.DB) error {
	dbName := os.Getenv("CLICKHOUSE_DB")
	driver, err := mclickhouse.WithInstance(db, &mclickhouse.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", dbName, driver)
	if err != nil {
		return err
	}
	return m.Up()
}
