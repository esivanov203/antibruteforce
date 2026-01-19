package http_api

import (
	"encoding/json"
	"net"
	"net/http"
)

type ResetBucketRequest struct {
	Login string `json:"login"`
	IP    string `json:"ip"`
}

func (s *Server) resetBucketHandler(w http.ResponseWriter, r *http.Request) {
	var req ResetBucketRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	ip := net.ParseIP(req.IP)
	if ip == nil {
		http.Error(w, "invalid ip", http.StatusBadRequest)
		return
	}

	err := s.app.ResetBuckets(req.Login, ip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
