package application

import (
	"agent/internal/config"
	"agent/internal/workers"
	"agent/pkg/do"
	"agent/pkg/logger"
	"context"
	"time"

	pb "github.com/vandi37/Calculator-Models"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	LOG_FILE = "logs." + time.Now().Format("15'04.01-02") + ".log"
)

type Application struct {
	config config.Config
	logger *zap.Logger
}

func New() *Application {
	cfg, err := config.LoadConfig()
	if err != nil {
		zap.New(logger.Setup(), zap.AddStacktrace(zap.ErrorLevel)).Fatal("error loading config", zap.Error(err))
	}
	logger := logger.ConsoleAndFile(cfg.LogFile)
	return &Application{*cfg, logger}
}

func (a *Application) Run(ctx context.Context) {
	// Creating client
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	cc, err := grpc.NewClient(a.config.GrpcPath, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		a.logger.Fatal("failed to create client", zap.Error(err))
	}
	defer cc.Close()
	client := pb.NewTaskServiceClient(cc)
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Running workers
	workers.RunMultiple(ctx, a.config.ComputingPower, a.config.RetryCount, a.logger, client, do.Do)
	// Waiting for context to be done
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	<-ctx.Done()
	a.logger.Info("server finish")
}
