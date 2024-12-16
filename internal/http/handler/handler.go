package handler

import (
	"encoding/json"
	"net/http"

	"github.com/vandi37/Calculator/pkg/calc"
	"github.com/vandi37/Calculator/pkg/logger"
)

const (
	NotFound    = "page not found"
	InvalidBody = "invalid body"
)

// Request with expression
type Request struct {
	Expression string `json:"expression"`
}

// Ok response
type ResponseOK struct {
	Result float64 `json:"result"`
}

// Error response
type ResponseError struct {
	Error string `json:"error"`
}

// Handler for http
type Handler struct {
	path   string
	logger *logger.Logger
}

// New handler
func NewHandler(path string, logger *logger.Logger) *Handler {
	return &Handler{path, logger}
}

// Serves http
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.path != r.URL.Path {
		w.WriteHeader(http.StatusNotFound)
		err := json.NewEncoder(w).Encode(ResponseError{NotFound})
		if err != nil {
			h.logger.Fatalln(err)
		}
		return
	}
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(ResponseError{InvalidBody})
		if err != nil {
			h.logger.Fatalln(err)
		}
		return
	}

	res, err := calc.Calc(req.Expression)
	if err != nil {
		w.WriteHeader(422)
		err := json.NewEncoder(w).Encode(ResponseError{err.Error()})
		if err != nil {
			h.logger.Fatalln(err)
		}
		return
	}
	h.logger.Printf("expression %s resulted to %.4f", req.Expression, res)

	err = json.NewEncoder(w).Encode(ResponseOK{res})
	if err != nil {
		h.logger.Fatalln(err)
	}
}
