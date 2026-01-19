package http_api

import (
	"encoding/json"
	"net/http"
)

// SubnetRequest { "subnet": "192.1.1.0/25"}
type AddSubnetRequest struct {
	Subnet string `json:"subnet"`
}

func (s *Server) addBlacklistHandler(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	var req AddSubnetRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Subnet == "" {
		http.Error(w, "missing subnet", http.StatusBadRequest)
		return
	}

	if err := s.app.AddToBlackList(req.Subnet); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) removeBlacklistHandler(w http.ResponseWriter, r *http.Request) {
	subnet := r.URL.Query().Get("subnet")
	if subnet == "" {
		http.Error(w, "missing subnet", http.StatusBadRequest)
		return
	}

	if err := s.app.RemoveFromBlackList(subnet); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) addWhitelistHandler(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	var req AddSubnetRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Subnet == "" {
		http.Error(w, "missing subnet", http.StatusBadRequest)
		return
	}

	if err := s.app.AddToWhiteList(req.Subnet); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) removeWhitelistHandler(w http.ResponseWriter, r *http.Request) {
	subnet := r.URL.Query().Get("subnet")
	if subnet == "" {
		http.Error(w, "missing subnet", http.StatusBadRequest)
		return
	}

	if err := s.app.RemoveFromWhiteList(subnet); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
