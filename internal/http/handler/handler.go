package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/vandi37/Calculator/pkg/calc"
	"github.com/vandi37/Calculator/pkg/logger"
)

const (
	NotFound         = "page not found"
	InvalidBody      = "invalid body"
	MethodNotAllowed = "method not allowed"
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

func SendJson(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// New handler
func NewHandler(path string, logger *logger.Logger) *Handler {
	return &Handler{path, logger}
}

// Serves http
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.path != r.URL.Path {
		w.WriteHeader(http.StatusNotFound)
		err := SendJson(w, ResponseError{NotFound})
		if err != nil {
			h.logger.Fatalln(err)
		}
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		err := SendJson(w, ResponseError{MethodNotAllowed})
		if err != nil {
			h.logger.Fatalln(err)
		}
		return
	}
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Expression == "" {
		w.WriteHeader(http.StatusBadRequest)
		err := SendJson(w, ResponseError{InvalidBody})
		if err != nil {
			h.logger.Fatalln(err)
		}
		return
	}

	res, err := calc.Calc(req.Expression)
	if err != nil {
		w.WriteHeader(422)
		err := SendJson(w, ResponseError{err.Error()})
		if err != nil {
			h.logger.Fatalln(err)
		}
		return
	}
	h.logger.Printf("expression %s resulted to %.4f", req.Expression, res)

	err = SendJson(w, ResponseOK{res})
	if err != nil {
		h.logger.Fatalln(err)
	}
}
