package stream_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/internal/service/mock_service"
	"github.com/vandi37/Calculator/internal/transport/stream"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestLoggerInterceptor(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		returnErr    error
		expectError  bool
		expectLog    bool
		expectedCode codes.Code
	}{
		{
			name:         "Success request",
			method:       "/package.Service/Method",
			returnErr:    nil,
			expectError:  false,
			expectLog:    true,
			expectedCode: codes.OK,
		},
		{
			name:         "Internal error",
			method:       "/package.Service/Method",
			returnErr:    status.Error(codes.Internal, "internal error"),
			expectError:  true,
			expectLog:    true,
			expectedCode: codes.Internal,
		},
		{
			name:         "Not found error",
			method:       "/package.Service/Method",
			returnErr:    status.Error(codes.NotFound, "not found"),
			expectError:  true,
			expectLog:    true,
			expectedCode: codes.NotFound,
		},
		{
			name:         "Non-status error",
			method:       "/package.Service/Method",
			returnErr:    errors.New("some error"),
			expectError:  true,
			expectLog:    true,
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			defer logger.Sync()

			interceptor := stream.LoggerInterceptor(logger)

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "response", tt.returnErr
			}

			info := &grpc.UnaryServerInfo{
				FullMethod: tt.method,
			}

			ctx := context.Background()
			resp, err := interceptor(ctx, "request", info, handler)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.returnErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "response", resp)
			}
		})
	}
}

func TestStreamService_TaskStream(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mock_service.MockService)
		expectErr bool
		sendCount int
	}{
		{
			name: "Successful stream with 3 tasks",
			setupMock: func(m *mock_service.MockService) {
				ch := make(chan *pb.Task, 3)
				ch <- &pb.Task{Id: "1"}
				ch <- &pb.Task{Id: "2"}
				ch <- &pb.Task{Id: "3"}
				close(ch)

				m.EXPECT().Tasks().Return(ch).AnyTimes()
			},
			expectErr: false,
			sendCount: 3,
		},
		{
			name: "Empty task channel",
			setupMock: func(m *mock_service.MockService) {
				ch := make(chan *pb.Task)
				close(ch)

				m.EXPECT().Tasks().Return(ch)
			},
			expectErr: false,
			sendCount: 0,
		},
		{
			name: "Error sending task",
			setupMock: func(m *mock_service.MockService) {
				ch := make(chan *pb.Task, 1)
				ch <- &pb.Task{Id: "1"}
				close(ch)

				m.EXPECT().Tasks().Return(ch)
			},
			expectErr: true,
			sendCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockService(ctrl)
			tt.setupMock(mockService)

			s := stream.New(mockService, zap.NewNop())

			mockStream := &mockTaskServiceTaskStreamServer{
				sendFunc: func(task *pb.Task) error {
					if tt.expectErr && tt.sendCount > 0 {
						return errors.New("send error")
					}
					return nil
				},
			}

			err := s.TaskStream(&pb.Void{}, mockStream)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.sendCount, mockStream.sendCount)
		})
	}
}

func TestStreamService_SendResult(t *testing.T) {
	tests := []struct {
		name        string
		result      *pb.Result
		setupMock   func(*mock_service.MockService, *pb.Result)
		expectErr   bool
		expectedErr error
	}{
		{
			name:   "Success",
			result: &pb.Result{Id: "1", Result: 42},
			setupMock: func(m *mock_service.MockService, res *pb.Result) {
				m.EXPECT().DoTask(gomock.Any(), res).Return(nil)
			},
			expectErr: false,
		},
		{
			name:   "Service error",
			result: &pb.Result{Id: "1", Result: 42},
			setupMock: func(m *mock_service.MockService, res *pb.Result) {
				m.EXPECT().DoTask(gomock.Any(), res).Return(errors.New("some error"))
			},
			expectErr:   true,
			expectedErr: status.Error(codes.NotFound, "some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockService(ctrl)
			tt.setupMock(mockService, tt.result)

			s := stream.New(mockService, zap.NewNop())

			resp, err := s.SendResult(context.Background(), tt.result)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, &pb.Void{}, resp)
			}
		})
	}
}

func TestStreamService_SendError(t *testing.T) {
	tests := []struct {
		name        string
		errorMsg    *pb.Error
		setupMock   func(*mock_service.MockService, *pb.Error)
		expectErr   bool
		expectedErr error
	}{
		{
			name:     "Success",
			errorMsg: &pb.Error{Id: "1", Error: "error"},
			setupMock: func(m *mock_service.MockService, err *pb.Error) {
				m.EXPECT().DoError(gomock.Any(), err).Return(nil)
			},
			expectErr: false,
		},
		{
			name:     "Service error",
			errorMsg: &pb.Error{Id: "1", Error: "error"},
			setupMock: func(m *mock_service.MockService, err *pb.Error) {
				m.EXPECT().DoError(gomock.Any(), err).Return(errors.New("some error"))
			},
			expectErr:   true,
			expectedErr: status.Error(codes.NotFound, "some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_service.NewMockService(ctrl)
			tt.setupMock(mockService, tt.errorMsg)

			s := stream.New(mockService, zap.NewNop())

			resp, err := s.SendError(context.Background(), tt.errorMsg)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, &pb.Void{}, resp)
			}
		})
	}
}

func TestStreamService_ToServer(t *testing.T) {
	t.Run("Creates gRPC server with interceptor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockService := mock_service.NewMockService(ctrl)
		s := stream.New(mockService, zap.NewNop())

		grpcServer := s.ToServer()

		assert.NotNil(t, grpcServer)
	})
}

type mockTaskServiceTaskStreamServer struct {
	grpc.ServerStream
	sendFunc  func(*pb.Task) error
	sendCount int
}

func (m *mockTaskServiceTaskStreamServer) Send(task *pb.Task) error {
	m.sendCount++
	return m.sendFunc(task)
}

func (m *mockTaskServiceTaskStreamServer) Context() context.Context {
	return context.Background()
}

func (m *mockTaskServiceTaskStreamServer) SetHeader(metadata.MD) error {
	return nil
}

func (m *mockTaskServiceTaskStreamServer) SendHeader(metadata.MD) error {
	return nil
}

func (m *mockTaskServiceTaskStreamServer) SetTrailer(metadata.MD) {
}

func (m *mockTaskServiceTaskStreamServer) SendMsg(msg interface{}) error {
	return nil
}

func (m *mockTaskServiceTaskStreamServer) RecvMsg(msg interface{}) error {
	return nil
}
