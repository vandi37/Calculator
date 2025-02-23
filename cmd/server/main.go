package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/vandi37/Calculator/internal/application"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	app := application.New(application.STD_CONFIG, application.LOG_SERVER_FILE)
	app.Run(ctx)
}
