package datastore

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"os"
	"reflect"
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
	name := ds.config["type"].(string)
	args := make(map[string]interface{}, 0)
	args["arg0"] = cfg["Debug"].(bool)
	result := DynamicCall(ds, strings.Title(strings.ToLower(name))+"Connect", args)
	value := reflect.ValueOf(result[0]).Interface().(*sql.DB)
	return value
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

func DynamicCall(obj interface{}, fn string, args map[string]interface{}) (res []reflect.Value) {
	method := reflect.ValueOf(obj).MethodByName(fn)
	var inputs []reflect.Value
	for _, v := range args {
		inputs = append(inputs, reflect.ValueOf(v))
	}
	return method.Call(inputs)
}
