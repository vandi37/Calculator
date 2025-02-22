package application

import (
	"context"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vandi37/Calculator/internal/agent/get"
	"github.com/vandi37/Calculator/internal/config"
	"github.com/vandi37/Calculator/internal/ms"
	"github.com/vandi37/Calculator/internal/transport/handler"
	"github.com/vandi37/Calculator/internal/transport/server"
	"github.com/vandi37/Calculator/internal/wait"
	"github.com/vandi37/Calculator/pkg/logger"
	"go.uber.org/zap"
)

type Application struct {
	config string
}

func New(config string) *Application {
	return &Application{config}
}

func (a *Application) Run(ctx context.Context) {
	gin.SetMode(gin.ReleaseMode)
	logger := logger.ConsoleAndFile("logs." + time.Now().Format("15'04.01-02") + ".log")

	config, err := config.LoadConfig(a.config)
	if err != nil {
		logger.Fatal("error loading config", zap.Error(err))
	}

	// building agent
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	get.RunMultiple(config.ComputingPower, config.Path.Task, config.Port.Api, config.MaxAgentErrors, config.AgentPeriodicityMs, logger)

	// Server
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	handler := handler.NewHandler(config.Path, wait.New(ms.From(*config), logger), logger)

	server := server.New(handler, config.Port.Api)

	go func() {
		if err := server.Run(); err != nil {
			logger.Fatal("error running server", zap.Error(err))
		}
	}()

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	<-ctx.Done()
	logger.Info("program exit")
	os.Exit(0)
}
