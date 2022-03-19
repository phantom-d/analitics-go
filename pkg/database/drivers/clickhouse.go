package drivers

import (
	"analitics/pkg/config"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	baseClick "github.com/ClickHouse/clickhouse-go"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type Clickhouse struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Pass     string `mapstructure:"pass"`
	CertPath string `mapstructure:"cert-path"`
}

func (click *Clickhouse) Connect() *gorm.DB {
	connect, err := sql.Open("clickhouse", click.dsn("tcp"))
	if err != nil {
		config.Logger.Fatal().Err(err).Msg("Error connection to clickhouse")
	}
	if err := connect.Ping(); err != nil {
		fmt.Println(err)
		return nil
	}
	db, err := gorm.Open(clickhouse.New(clickhouse.Config{Conn: connect}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		config.Logger.Fatal().Err(err).Msg("Error connection to clickhouse")
	}

	return db
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

			if err := baseClick.RegisterTLSConfig(click.CertPath, &tls.Config{RootCAs: caCertPool}); err != nil {
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
