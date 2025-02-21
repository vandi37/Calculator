package handler

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vandi37/Calculator/internal/config"
	"github.com/vandi37/Calculator/internal/wait"
)

type Handler struct {
	*gin.Engine
	Waiter *wait.Waiter
}

func SendJson(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func NewHandler(path config.Path, waiter *wait.Waiter) *Handler {
	res := &Handler{gin.New(), waiter}
	res.Use(gin.Recovery())
	res.POST(path.Calc, ContentType(res.CalcHandler))
	res.GET(path.Get, ContentType(res.ExpressionsHandler))
	res.GET(strings.TrimSuffix(path.Get, "/")+"/:id", ContentType(res.ById))
	res.GET(path.Task, ContentType(res.Task))
	res.POST(path.Task, res.FinishTask)
	return res
}
