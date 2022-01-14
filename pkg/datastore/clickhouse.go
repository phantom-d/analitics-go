package datastore

import (
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

func (ds *Datastore) ClickhouseConnect(debug bool) *sql.DB {
	dsn := "tcp://" + ds.config["host"].(string)
	dsn += ":" + strconv.Itoa(ds.config["port"].(int))
	dsn += "?compress=true&username=" + ds.config["user"].(string)
	dsn += "&password=" + ds.config["pass"].(string)
	dsn += "&database=" + ds.config["name"].(string)

	certPath := ds.config["cert-path"].(string)
	if _, err := os.Stat(certPath); err == nil {
		caCert, err := ioutil.ReadFile(certPath)
		if err != nil {
			log.Fatalf("Couldn't load file: %s", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		config := &tls.Config{RootCAs: caCertPool}

		if err := clickhouse.RegisterTLSConfig(certPath, config); err != nil {
			log.Fatalf("Couldn't register tls config: %s", err)
		}
		dsn += "&secure=true&tls_config=" + certPath
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
