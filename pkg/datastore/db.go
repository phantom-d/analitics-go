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

func InitDB(debug bool) *sql.DB {
	dbHost := os.Getenv("CLICKHOUSE_HOST")
	dbPort := os.Getenv("CLICKHOUSE_PORT")
	dbName := os.Getenv("CLICKHOUSE_DB")
	dbUser := os.Getenv("CLICKHOUSE_USER")
	dbPass := os.Getenv("CLICKHOUSE_PASS")
	dbCaCert := os.Getenv("CLICKHOUSE_CA")

	dsn := "tcp://" + dbHost + ":" + string(dbPort) + "?compress=true&username=" + dbUser + "&password=" + dbPass + "&database=" + dbName
	if _, err := os.Stat(dbCaCert); err == nil {
		caCert, err := ioutil.ReadFile(dbCaCert)
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
