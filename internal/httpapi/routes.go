package httpapi

import (
	"log"
	"net/http"
)

func (s *Server) setRowerHandlers() {
	s.router.HandleFunc("/", s.rootHandler).Methods(http.MethodGet)

	s.router.HandleFunc("/auth", s.authHandler).Methods(http.MethodPost)
	s.router.HandleFunc("/bucket/reset", s.resetBucketHandler).Methods(http.MethodPost)

	s.router.HandleFunc("/blacklist", s.addBlacklistHandler).Methods(http.MethodPost)
	s.router.HandleFunc("/blacklist", s.removeBlacklistHandler).Methods(http.MethodDelete)

	s.router.HandleFunc("/whitelist", s.addWhitelistHandler).Methods(http.MethodPost)
	s.router.HandleFunc("/whitelist", s.removeWhitelistHandler).Methods(http.MethodDelete)
}

func (s *Server) rootHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Welcome to ABF Service!"))
	if err != nil {
		// todo
		log.Printf("handler root: %v", err)
	}
}
