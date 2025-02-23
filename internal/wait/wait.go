package wait

import (
	"fmt"
	"slices"
	"sync"

	"github.com/vandi37/Calculator/internal/agent/module"
	"github.com/vandi37/Calculator/internal/ms"
	"github.com/vandi37/Calculator/internal/status"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
	"github.com/vandi37/vanerrors"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

type Waiter struct {
	processes map[int]status.Status
	waiting   []int
	msGetter  *ms.MsGetter
	logger    *zap.Logger
	mu        sync.Mutex
}

func New(msGetter *ms.MsGetter, logger *zap.Logger) *Waiter {
	return &Waiter{map[int]status.Status{}, []int{}, msGetter, logger, sync.Mutex{}}
}

func (w *Waiter) Add(expression string) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	keys := maps.Keys(w.processes)
	slices.Sort(keys)
	id := 0
	if len(keys) > 0 {
		id = keys[len(keys)-1] + 1
	}
	w.processes[id] = status.New(status.Nothing, status.NothingValue{}, expression)
	w.logger.Debug("process added", zap.Int("id", id), zap.String("expression", expression))
	return id
}

func (w *Waiter) StartWaiting(id int, arg1, arg2 float64, operation tree.ExprSep) (<-chan float64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	s, ok := w.processes[id]
	if !ok {
		w.logger.Warn("process id not found", zap.Int("id", id))
		return nil, vanerrors.New(IdNotFound, fmt.Sprint(id))
	}

	if s.Level != status.Nothing {
		w.logger.Warn("process status is not 'Nothing'", zap.Int("id", id), zap.Int("status", int(s.Level)), zap.String("expression", s.Expression))
		return nil, vanerrors.New(StatusIsNotNothing, s.Level.String())
	}

	back := make(chan float64)
	w.processes[id] = status.New(status.Waiting, status.WaitingValue{
		Request: module.Request{
			Id:              id,
			Arg1:            arg1,
			Arg2:            arg2,
			Operation:       operation,
			OperationTimeMs: w.msGetter.Get(operation),
		},
		Back: back,
	}, s.Expression)
	w.waiting = append(w.waiting, id)
	w.logger.Debug("added waiting", zap.Int("id", id), zap.String("expression", s.Expression))
	return back, nil
}

func (w *Waiter) GetJob() (any, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.waiting) <= 0 {
		w.logger.Debug("getting job is rejected: no jobs are needed to do")
		return nil, false
	}

	id := w.waiting[0]
	s, ok := w.processes[id]
	if !ok || s.Level != status.Waiting {
		w.logger.Warn("status is not matching", zap.Int("id", id), zap.String("expression", s.Expression))
		return nil, false
	}
	w.waiting = w.waiting[1:]
	w.processes[id] = status.New(status.Processing, status.ProcessingValue{Request: *s.Value.GerRequest(), Back: s.Value.GetBack()}, s.Expression)
	w.logger.Debug("sended job", zap.Int("id", id), zap.String("expression", s.Expression), zap.String("value", s.Value.String()))
	return s.Value.ForJson(), true
}

func (w *Waiter) FinishJob(id int, result float64) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	s, ok := w.processes[id]
	if !ok {
		w.logger.Warn("process id not found", zap.Int("id", id))
		return vanerrors.New(IdNotFound, fmt.Sprint(id))
	}

	if s.Level != status.Processing {
		w.logger.Warn("process status is not 'Processing", zap.Int("id", id), zap.Int("status", int(s.Level)), zap.String("expression", s.Expression))
		return vanerrors.New(StatusIsNotProcessing, s.Level.String())
	}
	s.Value.GetBack() <- result
	close(s.Value.GetBack())

	w.processes[id] = status.New(status.Nothing, status.NothingValue{}, s.Expression)
	w.logger.Debug("process got result", zap.Int("id", id), zap.String("expression", s.Expression), zap.String("task", s.Value.String()), zap.Float64("result", result))
	return nil
}

func (w *Waiter) Finish(id int, result float64) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	s, ok := w.processes[id]
	if !ok {
		w.logger.Warn("process id not found", zap.Int("id", id))
		return vanerrors.New(IdNotFound, fmt.Sprint(id))
	}

	if s.Level != status.Nothing {
		w.logger.Warn("process status is not 'Nothing'", zap.Int("id", id), zap.Int("status", int(s.Level)), zap.String("expression", s.Expression))
		return vanerrors.New(StatusIsNotNothing, s.Level.String())
	}

	w.processes[id] = status.New(status.Finished, status.FinishedValue(result), s.Expression)
	w.logger.Debug("process finished", zap.Int("id", id), zap.String("expression", s.Expression), zap.Float64("result", result))
	return nil
}

func (w *Waiter) Error(id int, err error) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	s, ok := w.processes[id]
	if !ok {
		w.logger.Warn("process id not found", zap.Int("id", id))
		return vanerrors.New(IdNotFound, fmt.Sprint(id))
	}
	switch s.Level {
	case status.Waiting, status.Processing:
		close(s.Value.GetBack())
	}
	w.processes[id] = status.New(status.Error, status.ErrorValue{Error: err}, s.Expression)
	w.logger.Debug("process got error", zap.Int("id", id), zap.String("expression", s.Expression), zap.Error(err))
	return nil
}

func (w *Waiter) WaitingFunc(id int) func(float64, float64, tree.ExprSep) (<-chan float64, error) {
	return func(arg1, arg2 float64, operation tree.ExprSep) (<-chan float64, error) {
		return w.StartWaiting(id, arg1, arg2, operation)
	}
}

func (w *Waiter) GetAll() []any {
	all := []any{}
	keys := maps.Keys(w.processes)
	slices.Sort(keys)
	for _, key := range keys {
		all = append(all, w.Get(key))
	}
	return all
}

func (w *Waiter) Exist(id int) bool {
	_, ok := w.processes[id]
	return ok
}

func (w *Waiter) Get(id int) any {
	s, ok := w.processes[id]
	if !ok {
		return nil
	}

	return s.ForJson(id)
}
