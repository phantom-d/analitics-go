package transport

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type Server struct {
	router *mux.Router
	db     *sql.DB
}

func NewServer(db *sql.DB) Server {
	s := Server{}
	s.db = db
	s.router = mux.NewRouter()

	s.router.HandleFunc("/v1/daemon/status", s.GetOrderHistoryV1).Methods(http.MethodGet)

	return s
}

func (s Server) Start() error {
	return http.ListenAndServe(":"+os.Getenv("HTTP_BIND"), s.router)
}

func (s Server) GetOrderHistoryV1(w http.ResponseWriter, r *http.Request) {
	daemonsStatus := ""
	response, err := json.Marshal(daemonsStatus)
	if err != nil {
		log.Error().Err(err).Msg("Response hasn't been marshaled.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(response)
}
