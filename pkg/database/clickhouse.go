package database

import (
	"analitics/pkg/config"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"github.com/ClickHouse/clickhouse-go"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func (ds *Datastore) ClickhouseConnect() *sql.DB {
	dsn := "tcp://" + ds.config["host"].(string)
	dsn += ":" + strconv.Itoa(ds.config["port"].(int))
	dsn += "?compress=true"
	dsn += "&database=" + ds.config["name"].(string)

	if _, ok := ds.config["user"].(string); ok {
		dsn += "&username=" + ds.config["user"].(string)
		if _, ok := ds.config["pass"].(string); ok {
			dsn += "&password=" + ds.config["pass"].(string)
		}
	}

	if certPath, ok := ds.config["cert-path"].(string); ok {
		if _, err := os.Stat(certPath); err == nil {
			caCert, err := ioutil.ReadFile(certPath)
			if err != nil {
				log.Fatalf("Couldn't load file: %s", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			if err := clickhouse.RegisterTLSConfig(certPath, &tls.Config{RootCAs: caCertPool}); err != nil {
				log.Fatalf("Couldn't register tls config: %s", err)
			}
			dsn += "&secure=true&tls_config=" + certPath
		}
	}

	if config.Application.Debug {
		dsn += "&debug=true"
	}

	connect, err := sql.Open("clickhouse", dsn)
	if err != nil {
		config.Logger.Fatal().Err(err).Msg("Error connection to clickhouse")
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
