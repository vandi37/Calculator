package application

import (
	"os"

	"github.com/vandi37/Calculator/config"
	"github.com/vandi37/Calculator/internal/http/handler"
	"github.com/vandi37/Calculator/internal/http/server"
	"github.com/vandi37/Calculator/pkg/logger"
)

type Application struct {
	config string
}

func New(config string) *Application {
	return &Application{config}
}

func (a *Application) Run() {
	// Creating logger
	logger := logger.New(os.Stderr)

	// Loading config
	config, err := config.LoadConfig(a.config)
	if err != nil {
		logger.Fatalln(err)
	}

	// Creating handler
	handler := handler.NewHandler(config.Path, logger)

	// Creating server
	server := server.New(handler, config.Port)

	// Running server
	server.Run()
	// The program end
}
