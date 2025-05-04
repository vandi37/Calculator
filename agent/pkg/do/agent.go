package do

import (
	"time"

	pb "github.com/vandi37/Calculator-Models"
)

func Do(req *pb.Task) (float64, error) {
	var f float64
	switch req.Operation {
	case pb.Operation_ADD:
		f = req.Arg1 + req.Arg2
	case pb.Operation_SUBTRACT:
		f = req.Arg1 - req.Arg2
	case pb.Operation_MULTIPLY:
		f = req.Arg1 * req.Arg2
	case pb.Operation_DIVIDE:
		if req.Arg2 == 0 {
			return 0, DivisionByZero
		}
		f = req.Arg1 / req.Arg2
	default:
		return 0, UnknownOperation
	}
	time.Sleep(time.Millisecond * time.Duration(req.OperationTime))
	return f, nil
}
