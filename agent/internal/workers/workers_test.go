package workers_test

import (
	"agent/internal/workers"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/vandi37/Calculator-Models"
	"go.uber.org/zap/zaptest"
)

func TestRunMultiple(t *testing.T) {
	t.Run("starts multiple workers", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := zaptest.NewLogger(t)
		client := NewTestingClient()

		go workers.RunMultiple(ctx, 3, 0, logger, client, func(req *pb.Task) (float64, error) {
			return 0, nil
		})

		time.Sleep(100 * time.Millisecond)

		close(client.tasks)
		var streamCount int
		for req := client.REQUEST(); req != nil; req = client.REQUEST() {
			if req.RequestType == SendStream {
				streamCount++
			}
		}
		assert.Equal(t, 3, streamCount)
	})

	t.Run("stops when context is cancelled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		logger := zaptest.NewLogger(t)
		client := NewTestingClient()

		done := make(chan struct{})
		go func() {
			workers.RunMultiple(ctx, 2, 0, logger, client, func(req *pb.Task) (float64, error) {
				return 0, nil
			})
			close(done)
		}()

		cancel()

		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("workers didn't stop after context cancellation")
		}
	})
}

func TestRun(t *testing.T) {
	t.Run("handles task processing successfully", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := zaptest.NewLogger(t)
		client := NewTestingClient()

		go workers.Run(ctx, 1, 0, logger, client, func(req *pb.Task) (float64, error) {
			return req.Arg1, nil
		})

		task := &pb.Task{
			Id:        "test1",
			Arg1:      5,
			Arg2:      3,
			Operation: pb.Operation_ADD,
		}
		client.TASK(task)
		<-client.queue // pass request for stream
		var resultReq Request
		select {
		case resultReq = <-client.queue:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timeout waiting for result")
		}

		require.Equal(t, SendResult, resultReq.RequestType)
		assert.Equal(t, "test1", resultReq.Result.Id)
		assert.Equal(t, task.Arg1, resultReq.Result.Result)
	})

	t.Run("handles task processing error", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := zaptest.NewLogger(t)
		client := NewTestingClient()

		go workers.Run(ctx, 1, 0, logger, client, func(req *pb.Task) (float64, error) {
			return 0, errors.New("processing error")
		})

		task := &pb.Task{Id: "test2"}
		client.TASK(task)
		<-client.queue // pass request for stream

		var errorReq Request
		select {
		case errorReq = <-client.queue:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timeout waiting for error")
		}

		require.Equal(t, SendError, errorReq.RequestType)
		assert.Equal(t, "processing error", errorReq.Error.Error)
	})

	t.Run("handles stream error", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := zaptest.NewLogger(t)
		client := NewTestingClient()

		client.TO_ERROR()

		workers.Run(ctx, 1, 0, logger, client, func(req *pb.Task) (float64, error) {
			return 0, nil
		})
	})

	t.Run("EOF", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := zaptest.NewLogger(t)
		client := NewTestingClient()

		go workers.Run(ctx, 1, 5, logger, client, func(req *pb.Task) (float64, error) {
			return 0, nil
		})

		close(client.tasks)

		time.Sleep(100 * time.Millisecond)

		var streamCount int
		for req := client.REQUEST(); req != nil; req = client.REQUEST() {
			if req.RequestType == SendStream {
				streamCount++
			}
		}
		assert.Equal(t, streamCount, 6)
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		logger := zaptest.NewLogger(t)
		client := NewTestingClient()

		done := make(chan struct{})
		go func() {
			workers.Run(ctx, 1, 0, logger, client, func(req *pb.Task) (float64, error) {
				return 0, nil
			})
			close(done)
		}()

		cancel()

		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("worker didn't stop after context cancellation")
		}
	})
}
