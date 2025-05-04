package appservice

import (
	"context"

	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/internal/repo"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
	"go.uber.org/zap"
)

var _ repo.Callback = (*Service)(nil)

// SendError implements repo.Callback.
func (s *Service) SendError(ctx context.Context, err error) {
	s.logger.Error("callback error", zap.Error(err))
}

// SendResult implements repo.Callback.
func (s *Service) SendResult(ctx context.Context, tasks []pb.Task) {
	for i := range tasks {
		task := &tasks[i]
		task.OperationTime = s.msGetter.Get(tree.Operation(task.Operation))
		select {
		case s.tasks <- task:
		case <-ctx.Done():
			s.logger.Warn("couldn't finish sending tasks")
			return
		}
		s.logger.Info("sent task", zap.String("task", task.Id))
	}
}
