package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/vandi37/Calculator/internal/application"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	app := application.New(application.STD_CONFIG, application.LOG_FILE)

	go app.RunAgent(ctx)
	app.Run(ctx)
}
