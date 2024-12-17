package handler

import (
	"encoding/json"
	"net/http"

	"github.com/vandi37/vanerrors"
)

const (
	UnknownCalculatorError = "unknown calculator error"
)

type Error struct {
	Main  string `json:"main"`
	Cause any    `json:"cause"`
}

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
		SendJson(w, ResponseError{GetErrUnprocessableEntity(err)})
		return
	}

	SendJson(w, ResponseOK{res})
}

func GetErrUnprocessableEntity(target error) string {
	err := vanerrors.Get(target)
	if err == nil {
		err = vanerrors.NewWrap(UnknownCalculatorError, err, vanerrors.EmptyHandler)
	}
	all := err.UnwrapAll()
	if len(all) == 0 {
		return UnknownCalculatorError
	}

	return all[len(all)-1].Error()
}
