package httpapi

import (
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

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	var req AuthRequest
	respBody := NewResponseBody(w, r)
	if !respBody.validate(&req) {
		return
	}

	result, err := s.app.Auth(req.Login, req.Password, req.IP)
	respBody.sendResult(http.StatusOK, result, err)
}
