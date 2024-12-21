package handler

import (
	"encoding/json"
	"net/http"

	"github.com/vandi37/Calculator/internal/http/input"
	"github.com/vandi37/Calculator/internal/http/resp"
)

const (
	UnknownCalculatorError = "unknown calculator error"
)

func (h *Handler) CalcHandler(w http.ResponseWriter, r *http.Request) {
	var req input.Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Expression == "" {
		w.WriteHeader(http.StatusBadRequest)
		SendJson(w, resp.ResponseError{Error: InvalidBody})
		return
	}

	res, err := h.calc.Run(req.Expression)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		SendJson(w, resp.ResponseError{Error: err.Error()})
		return
	}

	SendJson(w, resp.ResponseOK{Result: res})
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	SendJson(w, resp.ResponseError{Error: NotFound})
}
