package threads

import (
	"analitics/pkg/logger"
	"database/sql"
	"github.com/rs/zerolog"
)

type Thread struct {
	Config map[string]interface{}
	Item   map[string]interface{}
	db     *sql.DB
	logger *logger.Logger
}

func New(cfg map[string]interface{}, db *sql.DB, logger *logger.Logger) *Thread {
	if cfg["Enabled"].(bool) {
		th := &Thread{Config: cfg, db: db, logger: logger}
		return th
	} else {
		logger.Logger().Info().Msgf("Tread '%s' is disabled!", cfg)
	}
	return nil
}

func (th *Thread) RenewConnection() {

}

func (th *Thread) Logger() *zerolog.Logger {
	return th.logger.Logger()
}
