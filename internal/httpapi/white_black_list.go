package httpapi

import (
	"net/http"
	"strings"
)

// AddSubnetRequest SubnetRequest { "subnet": "192.1.1.0/25"}.
type SubnetRequest struct {
	Subnet string `json:"subnet"`
}

func (a *SubnetRequest) Validate() []string {
	var errs []string
	if strings.TrimSpace(a.Subnet) == "" {
		errs = append(errs, "subnet required")
	}

	return errs
}

func (s *Server) addBlacklistHandler(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	var req SubnetRequest
	respBody := NewResponseBody(w, r)
	if !respBody.validate(&req) {
		return
	}

	err := s.app.AddToBlacklist(req.Subnet)
	respBody.sendResult(http.StatusCreated, true, err)
}

func (s *Server) removeBlacklistHandler(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	subnet := r.URL.Query().Get("subnet")
	req := SubnetRequest{Subnet: subnet}

	respBody := NewResponseBody(w, r)
	if !respBody.validate(&req) {
		return
	}

	err := s.app.RemoveFromBlacklist(req.Subnet)
	respBody.sendResult(http.StatusNoContent, true, err)
}

func (s *Server) addWhitelistHandler(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	var req SubnetRequest
	respBody := NewResponseBody(w, r)
	if !respBody.validate(&req) {
		return
	}

	err := s.app.AddToWhitelist(req.Subnet)
	respBody.sendResult(http.StatusCreated, true, err)
}

func (s *Server) removeWhitelistHandler(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	subnet := r.URL.Query().Get("subnet")
	req := SubnetRequest{Subnet: subnet}

	respBody := NewResponseBody(w, r)
	if !respBody.validate(&req) {
		return
	}

	err := s.app.RemoveFromWhitelist(req.Subnet)
	respBody.sendResult(http.StatusNoContent, true, err)
}
