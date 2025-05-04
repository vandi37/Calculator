package handler

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/vandi37/Calculator/internal/service"
	"go.uber.org/zap"
)

type Handler struct {
	*gin.Engine
	Service service.Service
}

func SendJson(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func New(service service.Service, logger *zap.Logger) *Handler {
	router := &Handler{gin.New(), service}
	router.Use(gin.Recovery(), Logging(logger), CORSMiddleware(), ContentType())

	v1 := router.Group("/api/v1")
	v1.HEAD("/ping", router.PingHandler)
	withAuth := v1.Group("/", router.AuthMiddleware())
	withAuth.POST("/calculate", router.CalcHandler)
	withAuth.GET("/expressions", router.ExpressionsHandler)
	withAuth.GET("/expressions/:id", router.GetByIdHandler)
	v1.POST("/register", router.RegisterHandler)
	v1.POST("/login", router.LoginHandler)
	withAuth.PATCH("/username", router.ChangeUsernameHandler)
	withAuth.PATCH("/password", router.ChangePasswordHandler)
	withAuth.DELETE("/delete", router.DeleteHandler)
	return router
}
