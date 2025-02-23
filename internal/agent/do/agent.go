package do

import (
	"time"

	"github.com/vandi37/Calculator/internal/agent/module"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
)

func Do(req module.Request) float64 {
	var f float64
	switch req.Operation {
	case tree.Addition:
		f = req.Arg1 + req.Arg2
	case tree.Subtraction:
		f = req.Arg1 - req.Arg2
	case tree.Multiplication:
		f = req.Arg1 * req.Arg2
	case tree.Division:
		f = req.Arg1 / req.Arg2
	default:
		return 0
	}
	time.Sleep(time.Millisecond * time.Duration(req.OperationTimeMs))
	return f
}
