package httpapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Validatable interface {
	Validate() []string
}

type ResponseBody struct {
	OK     bool     `json:"ok"`
	Errors []string `json:"errors"`
	w      http.ResponseWriter
	r      *http.Request
}

func NewResponseBody(w http.ResponseWriter, r *http.Request) *ResponseBody {
	w.Header().Set("Content-Type", "application/json")
	return &ResponseBody{w: w, r: r}
}

func (rb *ResponseBody) sendResult(status int, result bool, err error) {
	if err != nil {
		rb.Errors = append(rb.Errors, err.Error())
		rb.w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(rb.w).Encode(rb); err != nil {
			log.Printf("ERROR: encode: %v", err)
		}
		return
	}

	rb.OK = result
	rb.w.WriteHeader(status)
	if err := json.NewEncoder(rb.w).Encode(rb); err != nil {
		log.Printf("encode: %v", err)
	}
}

func (rb *ResponseBody) validate(req Validatable) bool {
	if err := json.NewDecoder(rb.r.Body).Decode(&req); err != nil {
		rb.Errors = append(rb.Errors, fmt.Sprintf("json format error: %v", err))
		rb.w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(rb.w).Encode(rb); err != nil {
			log.Printf("ERROR: auth handler encode: %v", err)
		}
		return false
	}

	if errs := req.Validate(); errs != nil {
		rb.Errors = append(rb.Errors, errs...)
		rb.w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(rb.w).Encode(rb); err != nil {
			log.Printf("ERROR: auth handler encode: %v", err)
		}
		return false
	}

	return true
}
