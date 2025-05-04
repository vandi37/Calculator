// This package has modules for ast tree
//
// It doesn't need test just a model package
package tree

import (
	"fmt"

	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/pkg/parsing/tokens"
)

type Ast struct {
	Expression ExpressionType
}

type ExpressionType interface {
	fmt.Stringer
	expression()
}

type Operation pb.Operation

func SepFrom(kind tokens.TokenKind) (Operation, bool) {
	switch kind {
	case tokens.Addition:
		return Operation(pb.Operation_ADD), true
	case tokens.Subtraction:
		return Operation(pb.Operation_SUBTRACT), true
	case tokens.Multiplication:
		return Operation(pb.Operation_MULTIPLY), true
	case tokens.Division:
		return Operation(pb.Operation_DIVIDE), true
	default:
		return Operation(-1), false
	}
}

func (s Operation) String() string {
	switch pb.Operation(s) {
	case pb.Operation_ADD:
		return "+"
	case pb.Operation_SUBTRACT:
		return "-"
	case pb.Operation_MULTIPLY:
		return "*"
	case pb.Operation_DIVIDE:
		return "/"
	default:
		return "[unknown separator]"
	}
}

type Expression struct {
	Left  ExpressionType
	Operation   Operation
	Right ExpressionType
}

func (e Expression) expression() {}

func (e Expression) String() string {
	return fmt.Sprintf("(%s) %s (%s)", e.Left.String(), e.Operation.String(), e.Right.String())
}

type Num float64

func (n Num) expression() {}

func (n Num) String() string {
	return fmt.Sprint(float64(n))
}
