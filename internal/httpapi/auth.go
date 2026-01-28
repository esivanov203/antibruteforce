package httpapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	IP       string `json:"ip"`
}

func (a *AuthRequest) Validate() []string {
	var errs []string
	if strings.TrimSpace(a.Login) == "" {
		errs = append(errs, "login required")
	}
	if strings.TrimSpace(a.Password) == "" {
		errs = append(errs, "password required")
	}
	if strings.TrimSpace(a.IP) == "" {
		errs = append(errs, "ip required")
	}

	return errs
}

type AuthResponse struct {
	OK     bool     `json:"ok"`
	Errors []string `json:"errors"`
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req  AuthRequest
		resp AuthResponse
	)
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Errors = append(resp.Errors, fmt.Sprintf("json format error: %v", err))
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("ERROR: auth handler encode: %v", err)
		}
		return
	}

	if errs := req.Validate(); errs != nil {
		resp.Errors = append(resp.Errors, errs...)
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("ERROR: auth handler encode: %v", err)
		}
		return
	}

	result, err := s.app.Auth(req.Login, req.Password, req.IP)
	if err != nil {
		resp.Errors = append(resp.Errors, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("ERROR: auth handler encode: %v", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(AuthResponse{OK: result}); err != nil {
		// todo error
		log.Printf("auth handler encode: %v", err)
	}
}
