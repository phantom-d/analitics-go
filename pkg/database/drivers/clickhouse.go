package drivers

import (
	"analitics/pkg/config"

	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/ClickHouse/clickhouse-go"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Clickhouse struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Pass     string `mapstructure:"pass"`
	CertPath string `mapstructure:"cert-path"`
}

func (click *Clickhouse) Connect() *sql.DB {
	connect, err := sql.Open("clickhouse", click.dsn("tcp"))
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

func (click *Clickhouse) MigrateUp(source string) {
	m, err := migrate.New(source, click.dsn("clickhouse"))
	if err != nil {
		config.Logger.Fatal().Err(err).Msg("Migration")
	}
	config.Logger.Info().Msg("Migration: start...")
	err = m.Up()
	if err != nil {
		if err.Error() == "no change" {
			config.Logger.Info().Msgf("Migration: %s!", err.Error())
		} else {
			config.Logger.Error().Err(err).Msg("Migration")
		}
	}
	config.Logger.Info().Msg("Migration: end...")
}

func (click *Clickhouse) dsn(scheme string) (result string) {
	result = scheme + "://" + click.Host
	result += ":" + strconv.Itoa(click.Port)
	result += "?compress=true&x-multi-statement=true&x-migrations-table-engine=MergeTree"
	result += "&database=" + click.Name

	if click.User != "" {
		result += "&username=" + click.User
		if click.Pass != "" {
			result += "&password=" + click.Pass
		}
	}

	if click.CertPath != "" {
		if _, err := os.Stat(click.CertPath); err == nil {
			caCert, err := ioutil.ReadFile(click.CertPath)
			if err != nil {
				log.Fatalf("Couldn't load file: %s", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			if err := clickhouse.RegisterTLSConfig(click.CertPath, &tls.Config{RootCAs: caCertPool}); err != nil {
				log.Fatalf("Couldn't register tls config: %s", err)
			}
			result += "&secure=true&tls_config=" + click.CertPath
		}
	}

	if config.Application.Debug {
		result += "&debug=true"
	}
	return
}
