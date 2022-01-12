package datastore

import (
	"analitics-go/pkg/application"
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

type AppConfig struct {
	application.Config
}

func New(cfg *application.Config) *sql.DB {
	return InitDB(cfg)
}

func InitDB(cfg *application.Config) *sql.DB {
	dbParams := cfg.Database
	dsn := "tcp://" + dbParams.Host + ":" + string(dbParams.Port) + "?compress=true&username=" + dbParams.User + "&password=" + dbParams.Pass + "&database=" + dbParams.Name
	if _, err := os.Stat(dbParams.CertPath); err == nil {
		caCert, err := ioutil.ReadFile(dbParams.CertPath)
		if err != nil {
			log.Fatalf("Couldn't load file", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		config := &tls.Config{RootCAs: caCertPool}

		clickhouse.RegisterTLSConfig("yandex-cloud", config)
		dsn += "&secure=true&tls_config=yandex-cloud"
	}

	if cfg.Debug {
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
