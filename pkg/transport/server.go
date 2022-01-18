package transport

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type Server struct {
	router *mux.Router
}

func NewServer() *Server {
	s := &Server{}
	s.router = mux.NewRouter()
	s.router.HandleFunc("/v1/daemon/status", s.GetDaemonsStatusV1).Methods(http.MethodGet)
	return s
}

func (s *Server) Start() error {
	return http.ListenAndServe(":"+os.Getenv("HTTP_BIND"), s.router)
}

func (s *Server) GetDaemonsStatusV1(w http.ResponseWriter, r *http.Request) {
	daemonsStatus := ""
	response, err := json.Marshal(daemonsStatus)
	if err != nil {
		Logger().Error().Err(err).Msg("Response hasn't been marshaled.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(response)
}
