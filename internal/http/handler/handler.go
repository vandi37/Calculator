package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/vandi37/Calculator/pkg/calc_service"
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
	path string
	fn   http.HandlerFunc
	calc *calc_service.Calculator
}

func SendJson(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func NewHandler(path string, calc *calc_service.Calculator) *Handler {
	res := &Handler{path, nil, calc}
	res.fn = CheckMethod(http.MethodPost, res.CalcHandler)
	return res
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.path != r.URL.Path {
		w.WriteHeader(http.StatusNotFound)
		SendJson(w, ResponseError{NotFound})
		return
	}

	if h.fn != nil {
		h.fn.ServeHTTP(w, r)
	}

}
