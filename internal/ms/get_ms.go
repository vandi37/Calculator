package ms

import (
	"github.com/vandi37/Calculator/internal/config"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
)

type MsGetter struct {
	TimeAdditionMs       int
	TimeSubtractionMs    int
	TimeMultiplicationMs int
	TimeDivisionMs       int
}

func From(t config.Time) *MsGetter {
	return &MsGetter{
		TimeAdditionMs:       t.AdditionMs,
		TimeSubtractionMs:    t.SubtractionMs,
		TimeMultiplicationMs: t.MultiplicationMs,
		TimeDivisionMs:       t.DivisionMs,
	}
}

func (g *MsGetter) Get(operation tree.ExprSep) int {
	switch operation {
	case tree.Addition:
		return g.TimeAdditionMs
	case tree.Subtraction:
		return g.TimeSubtractionMs
	case tree.Multiplication:
		return g.TimeMultiplicationMs
	case tree.Division:
		return g.TimeDivisionMs
	default:
		return -1
	}
}
