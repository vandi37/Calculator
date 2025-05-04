package stream

import (
	"context"
	"time"

	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		latency := time.Since(start)
		method := info.FullMethod
		statusCode := codes.OK

		if st, ok := status.FromError(err); ok {
			statusCode = st.Code()
		} else if err != nil {
			statusCode = codes.Internal
		}

		fields := []zap.Field{
			zap.String("method", method),
			zap.Duration("latency", latency),
			zap.Int32("status_code", int32(statusCode)),
		}

		switch {
		case statusCode == codes.Internal:
			logger.Error("grpc request", fields...)
		default:
			logger.Info("grpc request", fields...)
		}

		if err != nil {
			logger.Error("grpc error", zap.Error(err), zap.String("method", method))
		}

		return resp, err
	}
}

type StreamService struct {
	pb.UnimplementedTaskServiceServer
	service service.Service
	logger  *zap.Logger
}

func New(service service.Service, logger *zap.Logger) *StreamService {
	return &StreamService{
		service: service,
		logger:  logger,
	}
}

func (s *StreamService) ToServer() *grpc.Server {
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(LoggerInterceptor(s.logger)))
	pb.RegisterTaskServiceServer(grpcServer, s)
	return grpcServer
}

// TaskStream implements stream.TaskServiceServer.
func (s *StreamService) TaskStream(void *pb.Void, stream pb.TaskService_TaskStreamServer) error {
	for {
		task, ok := <-s.service.Tasks()
		if !ok {
			return nil
		}
		if err := stream.Send(task); err != nil {
			return err
		}
	}
}

// SendResult implements stream.TaskServiceServer.
func (s *StreamService) SendResult(ctx context.Context, res *pb.Result) (*pb.Void, error) {
	err := s.service.DoTask(ctx, res)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.Void{}, nil
}

// SendError implements stream.TaskServiceServer.
func (s *StreamService) SendError(ctx context.Context, res *pb.Error) (*pb.Void, error) {
	err := s.service.DoError(ctx, res)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.Void{}, nil
}

var _ pb.TaskServiceServer = (*StreamService)(nil)
