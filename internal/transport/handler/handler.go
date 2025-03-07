package handler

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vandi37/Calculator/internal/config"
	"github.com/vandi37/Calculator/internal/wait"
	"go.uber.org/zap"
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

func NewHandler(path config.Path, waiter *wait.Waiter, logger *zap.Logger) *Handler {
	router := &Handler{gin.New(), waiter}
	router.Use(gin.Recovery(), Logging(logger), CORSMiddleware())
	router.POST(path.Calc, ContentType(router.CalcHandler))
	router.GET(path.Get, ContentType(router.ExpressionsHandler))
	router.GET(strings.TrimSuffix(path.Get, "/")+"/:id", ContentType(router.ById))
	router.GET(path.Task, ContentType(router.Task))
	router.POST(path.Task, router.FinishTask)
	return router
}
