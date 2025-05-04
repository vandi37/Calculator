package application

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/vandi37/Calculator/internal/config"
	"github.com/vandi37/Calculator/internal/ms"
	"github.com/vandi37/Calculator/internal/repo/expressionrepo"
	"github.com/vandi37/Calculator/internal/repo/userrepo"
	"github.com/vandi37/Calculator/internal/service/appservice"
	"github.com/vandi37/Calculator/internal/transport/handler"
	"github.com/vandi37/Calculator/internal/transport/server"
	"github.com/vandi37/Calculator/internal/transport/stream"
	"github.com/vandi37/Calculator/pkg/hash"
	"github.com/vandi37/Calculator/pkg/jwt"
	"github.com/vandi37/Calculator/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

var (
	LOG_FILE = "logs." + time.Now().Format("15'04.01-02") + ".log"
)

const DB_NAME = "calculator"

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
	// Connect to mongo
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(a.config.MongoUri))
	if err != nil {
		a.logger.Fatal("error creating client", zap.Error(err))
	}
	defer client.Disconnect(ctx)
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Ping mongo
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		a.logger.Fatal("error pinging client", zap.Error(err))
	}
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Getting durations
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	d, err := time.ParseDuration(a.config.ResetTaskDuration)
	if err != nil {
		a.logger.Fatal("error parsing duration", zap.Error(err))
	}
	expire, err := time.ParseDuration(a.config.JWT.Expires)
	if err != nil {
		a.logger.Fatal("error parsing duration", zap.Error(err))
	}
	notBefore, err := time.ParseDuration(a.config.JWT.NotBefore)
	if err != nil {
		a.logger.Fatal("error parsing duration", zap.Error(err))
	}
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Creating repos
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	db := client.Database(DB_NAME)
	userRepo := userrepo.New(db)
	expressionRepo := expressionrepo.New(db, d)
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Creating service
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	service := appservice.New(
		a.logger,
		ms.From(a.config.Time),
		userRepo, expressionRepo,
		hash.NewPasswordService(nil),
		jwt.New(a.config.JWT.Secret, expire, notBefore),
	)
	service.Init(ctx)
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Creating handler
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	handler := handler.New(service, a.logger)
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Creating servers
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	server := server.New(handler, a.config.Port)
	defer server.Shutdown(ctx)

	lis, err := net.Listen("tcp", fmt.Sprint(":", a.config.GRPCProt))
	if err != nil {
		a.logger.Fatal("failed to listen", zap.Error(err))
	}
	defer lis.Close()

	grpcServer := stream.New(service, a.logger).ToServer()
	defer grpcServer.GracefulStop()
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Running servers
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	go func() {
		if err := server.Run(a.logger); err != nil {
			a.logger.Fatal("error running server", zap.Error(err))
		}
	}()

	go func() {
		a.logger.Info("grpc server running")
		if err := grpcServer.Serve(lis); err != nil {
			a.logger.Fatal("error running grpc server", zap.Error(err))
		}
	}()
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Waiting for context to be done
	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	<-ctx.Done()
	a.logger.Info("server finish")
}
