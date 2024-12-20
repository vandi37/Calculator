package handler

import (
	"encoding/json"
	"net/http"
)

const (
	UnknownCalculatorError = "unknown calculator error"
)

func (h *Handler) CalcHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Expression == "" {
		w.WriteHeader(http.StatusBadRequest)
		SendJson(w, ResponseError{InvalidBody})
		return
	}

	res, err := h.calc.Run(req.Expression)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		SendJson(w, ResponseError{err.Error()})
		return
	}

	SendJson(w, ResponseOK{res})
}
