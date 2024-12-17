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

type ResponseOK struct {
	Result float64 `json:"result"`
}

type ResponseError struct {
	Error string `json:"error"`
}

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

func NewHandler(path string, logger *logger.Logger) *Handler {
	return &Handler{path, logger}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if h.path != r.URL.Path {
		w.WriteHeader(http.StatusNotFound)
		SendJson(w, ResponseError{NotFound})
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		SendJson(w, ResponseError{MethodNotAllowed})
	}
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Expression == "" {
		w.WriteHeader(http.StatusBadRequest)
		SendJson(w, ResponseError{InvalidBody})
		return
	}

	res, err := calc.Calc(req.Expression)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		SendJson(w, ResponseError{err.Error()})
		return
	}
	h.logger.Printf("expression %s resulted to %.4f", req.Expression, res)

	SendJson(w, ResponseOK{res})
}
