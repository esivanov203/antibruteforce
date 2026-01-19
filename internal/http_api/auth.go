package http_api

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	IP       string `json:"ip"`
}

type AuthResponse struct {
	OK bool `json:"ok"`
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	ip := net.ParseIP(req.IP)
	if ip == nil {
		http.Error(w, "invalid ip", http.StatusBadRequest)
		return
	}

	result := s.app.Auth(req.Login, req.Password, ip)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(AuthResponse{OK: result}); err != nil {
		// todo error
		log.Printf("auth handler encode: %v", err)
	}
}
