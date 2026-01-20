package httpapi

import (
	"encoding/json"
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

	err := s.app.ResetBuckets(req.Login, req.IP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
