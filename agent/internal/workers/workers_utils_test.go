package workers_test

import (
	"context"
	"errors"
	"io"

	pb "github.com/vandi37/Calculator-Models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type TestingClient struct {
	nextError bool
	queue     chan Request
	tasks     chan *pb.Task
}

func NewTestingClient() *TestingClient {
	return &TestingClient{
		queue: make(chan Request, 64),
		tasks: make(chan *pb.Task, 16),
	}
}

func (t *TestingClient) REQUEST() *Request {
	select {
	case req := <-t.queue:
		return &req
	default:
		return nil
	}
}

func (t *TestingClient) TO_ERROR() {
	t.nextError = true
}

func (t *TestingClient) TASK(task *pb.Task) {
	t.tasks <- task
}

type RequestType byte

const (
	SendResult RequestType = iota
	SendError
	SendStream
)

type Request struct {
	RequestType RequestType
	Result      *pb.Result
	Error       *pb.Error
	Stream      *pb.Void
}

func (t *TestingClient) getError() error {
	if t.nextError {
		t.nextError = false
		return TestError
	}
	return nil
}

func (t *TestingClient) Close() {
	close(t.queue)
	close(t.tasks)
}

var TestError = errors.New("test error")

// SendError implements stream.TaskServiceClient.
func (t *TestingClient) SendError(ctx context.Context, in *pb.Error, opts ...grpc.CallOption) (*pb.Void, error) {
	t.queue <- Request{RequestType: SendError, Error: in}
	return &pb.Void{}, t.getError()
}

// SendResult implements stream.TaskServiceClient.
func (t *TestingClient) SendResult(ctx context.Context, in *pb.Result, opts ...grpc.CallOption) (*pb.Void, error) {
	t.queue <- Request{RequestType: SendResult, Result: in}
	return &pb.Void{}, t.getError()
}

// TaskStream implements stream.TaskServiceClient.
func (t *TestingClient) TaskStream(ctx context.Context, in *pb.Void, opts ...grpc.CallOption) (grpc.ServerStreamingClient[pb.Task], error) {
	t.queue <- Request{RequestType: SendStream, Stream: in}
	if err := t.getError(); err != nil {
		return nil, err
	}
	return NewTestStream(t.tasks), nil
}

var _ pb.TaskServiceClient = (*TestingClient)(nil)

type TestStream struct {
	queue chan *pb.Task
}

func NewTestStream(queue chan *pb.Task) *TestStream {
	return &TestStream{queue}
}

var (
	Unimplemented = errors.New("error sended if unimplemented")
)

// CloseSend implements grpc.ServerStreamingClient.
func (t *TestStream) CloseSend() error {
	return Unimplemented
}

// Context implements grpc.ServerStreamingClient.
func (t *TestStream) Context() context.Context {
	return context.Background()
}

// Header implements grpc.ServerStreamingClient.
func (t *TestStream) Header() (metadata.MD, error) {
	return nil, Unimplemented
}

// Recv implements grpc.ServerStreamingClient.
func (t *TestStream) Recv() (*pb.Task, error) {
	task, ok := <-t.queue
	if !ok {
		return nil, io.EOF
	}
	return task, nil
}

// RecvMsg implements grpc.ServerStreamingClient.
func (t *TestStream) RecvMsg(m any) error {
	return Unimplemented
}

// SendMsg implements grpc.ServerStreamingClient.
func (t *TestStream) SendMsg(m any) error {
	return Unimplemented
}

// Trailer implements grpc.ServerStreamingClient.
func (t *TestStream) Trailer() metadata.MD {
	return make(metadata.MD)
}

var _ grpc.ServerStreamingClient[pb.Task] = (*TestStream)(nil)
