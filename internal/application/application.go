package application

import (
	"context"
	"time"

	"github.com/vandi37/Calculator/internal/agent/get"
	"github.com/vandi37/Calculator/internal/config"
	"github.com/vandi37/Calculator/internal/ms"
	"github.com/vandi37/Calculator/internal/transport/handler"
	"github.com/vandi37/Calculator/internal/transport/server"
	"github.com/vandi37/Calculator/internal/wait"
	"github.com/vandi37/Calculator/pkg/logger"
	"go.uber.org/zap"
)

var (
	LOG_SERVER_FILE = "logs.server." + time.Now().Format("15'04.01-02") + ".log"
	LOG_AGENT_FILE  = "logs.agent." + time.Now().Format("15'04.01-02") + ".log"
	LOG_FILE        = "logs." + time.Now().Format("15'04.01-02") + ".log"
)

type Application struct {
	config config.Config
	logger *zap.Logger
}

func New(logPath string) *Application {
	logger := logger.ConsoleAndFile(logPath)
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("error loading config", zap.Error(err))
	}
	if err := cfg.Save(config.STD_JSON); err != nil {
		logger.Fatal("error saving config", zap.Error(err))
	}
	return &Application{*cfg, logger}
}

func (a *Application) RunAgent(ctx context.Context) {
	get.RunMultiple(ctx, a.config.ComputingPower, a.config.Path.Task, a.config.Port, a.config.MaxAgentErrors, a.config.AgentPeriodicityMs, a.logger)
	a.logger.Info("agent finish")
}

func (a *Application) Run(ctx context.Context) {
	handler := handler.NewHandler(a.config.Path, wait.New(ms.From(a.config.Time), a.logger), a.logger)

	server := server.New(handler, a.config.Port)

	go func() {
		if err := server.Run(a.logger); err != nil {
			a.logger.Fatal("error running server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	server.Close()
	a.logger.Info("server finish")
}
