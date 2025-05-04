package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/vandi37/Calculator/internal/application"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := application.New()
	app.Run(ctx)
}
