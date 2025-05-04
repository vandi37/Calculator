package ms

import (
	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/internal/config"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
)

type MsGetter struct {
	config.Time
}

func From(t config.Time) *MsGetter {
	return &MsGetter{t}
}

func (g *MsGetter) Get(operation tree.Operation) int32 {
	switch pb.Operation(operation) {
	case pb.Operation_ADD:
		return g.AdditionMs
	case pb.Operation_SUBTRACT:
		return g.SubtractionMs
	case pb.Operation_MULTIPLY:
		return g.MultiplicationMs
	case pb.Operation_DIVIDE:
		return g.DivisionMs
	default:
		return -1
	}
}
