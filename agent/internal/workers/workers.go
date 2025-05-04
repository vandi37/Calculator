package workers

import (
	"context"
	"io"

	pb "github.com/vandi37/Calculator-Models"
	"go.uber.org/zap"
)

type DoingFunc func(req *pb.Task) (float64, error)

func RunMultiple(ctx context.Context, num, retryCount int, logger *zap.Logger, client pb.TaskServiceClient, doing DoingFunc) {
	logger.Info("starting workers", zap.Int("workers", num))
	for i := 0; i < num; i++ {
		go Run(ctx, i, retryCount, logger, client, doing)
	}
	<-ctx.Done()
}

func Run(ctx context.Context, id, retryCount int, logger *zap.Logger, client pb.TaskServiceClient, doing DoingFunc) {
	worker := zap.Int("worker", id)
	logger.Info("started", worker)
	stream, err := client.TaskStream(ctx, &pb.Void{})
	if err != nil {
		logger.Error("stream failed", worker, zap.Error(err))
		return
	}
	defer stream.CloseSend()
	tasks, errors := func() (<-chan *pb.Task, <-chan error) {
		errors, tasks := make(chan error), make(chan *pb.Task)
		go func() {
			defer close(errors)
			defer close(tasks)
			for {
				task, err := stream.Recv()
				if err != nil {
					errors <- err
					return
				}
				tasks <- task
			}
		}()
		return tasks, errors
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errors:
			if err == io.EOF {
				logger.Error("stream closed")
			} else {
				logger.Error("stream failed", worker, zap.Error(err))
			}
			if retryCount > 0 {
				logger.Info("retrying", worker)
				Run(ctx, id, retryCount-1, logger, client, doing)
			}
			return
		case task := <-tasks:
			logger.Debug("got task", worker, zap.String("id", task.Id), zap.Float64("arg1", task.Arg1), zap.String("operation", task.Operation.String()), zap.Float64("arg2", task.Arg2))
			res, err := doing(task)
			if err != nil {
				logger.Debug("task failed", worker, zap.String("id", task.Id), zap.Error(err))
				_, err := client.SendError(ctx, &pb.Error{Id: task.Id, Error: err.Error()})
				if err != nil {
					logger.Error("sending error failed", worker, zap.String("id", task.Id), zap.Error(err))
				}
				continue
			}
			_, err = client.SendResult(ctx, &pb.Result{
				Id:     task.Id,
				Result: res,
			})
			if err != nil {
				logger.Error("sending result failed", worker, zap.String("id", task.Id), zap.Error(err))
				continue
			}

			logger.Debug("sended result", worker, zap.String("id", task.Id), zap.Float64("result", res), zap.Float64("arg1", task.Arg1), zap.String("operation", task.Operation.String()), zap.Float64("arg2", task.Arg2))
		}
	}
}
