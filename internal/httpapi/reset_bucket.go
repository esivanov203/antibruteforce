package httpapi

import (
	"net/http"
	"strings"
)

type ResetBucketRequest struct {
	Login string `json:"login"`
	IP    string `json:"ip"`
}

func (a *ResetBucketRequest) Validate() []string {
	var errs []string
	if strings.TrimSpace(a.Login) == "" {
		errs = append(errs, "login required")
	}
	if strings.TrimSpace(a.IP) == "" {
		errs = append(errs, "ip required")
	}

	return errs
}

func (s *Server) resetBucketHandler(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	var req ResetBucketRequest
	respBody := NewResponseBody(w, r)
	if !respBody.validate(&req) {
		return
	}

	err := s.app.ResetBuckets(req.Login, req.IP)
	respBody.sendResult(http.StatusOK, true, err)
}
