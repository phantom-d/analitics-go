package datastore

import (
	"fmt"
	"log"
	"os"

	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"io/ioutil"

	"github.com/ClickHouse/clickhouse-go"
	"github.com/golang-migrate/migrate/v4"

	mclickhouse "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func New(cfg map[string]interface{}) *sql.DB {
	return InitDB(cfg["Database"].(map[string]interface{}), cfg["Debug"].(bool))
}

func InitDB(dbParams map[string]interface{}, debug bool) *sql.DB {
	dsn := "tcp://" + dbParams["Host"].(string)
	dsn += ":" + dbParams["Port"].(string)
	dsn += "?compress=true&username=" + dbParams["User"].(string)
	dsn += "&password=" + dbParams["Pass"].(string)
	dsn += "&database=" + dbParams["Name"].(string)

	certPath := dbParams["CertPath"].(string)
	if _, err := os.Stat(certPath); err == nil {
		caCert, err := ioutil.ReadFile(certPath)
		if err != nil {
			log.Fatalf("Couldn't load file", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		config := &tls.Config{RootCAs: caCertPool}

		clickhouse.RegisterTLSConfig("yandex-cloud", config)
		dsn += "&secure=true&tls_config=yandex-cloud"
	}

	if debug {
		dsn += "&debug=true"
	}

	connect, err := sql.Open("clickhouse", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := connect.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
		return nil
	}

	return connect
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
