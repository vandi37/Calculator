// This package has modules for ast tree
package tree

import (
	"fmt"

	"github.com/vandi37/Calculator/pkg/parsing/tokens"
)

type Ast struct {
	ExpressionType
}

type ExpressionType interface {
	fmt.Stringer
	expression()
}

type ExprSep int

const (
	Addition ExprSep = iota << iota // idk why i did it ))
	Subtraction
	Multiplication
	Division
)

func SepFrom(kind tokens.TokenKind) (ExprSep, bool) {
	switch kind {
	case tokens.Addition:
		return Addition, true
	case tokens.Subtraction:
		return Subtraction, true
	case tokens.Multiplication:
		return Multiplication, true
	case tokens.Division:
		return Division, true
	default:
		return ExprSep(-1), false
	}
}

func (s ExprSep) String() string {
	switch s {
	case Addition:
		return "+"
	case Subtraction:
		return "-"
	case Multiplication:
		return "*"
	case Division:
		return "/"
	default:
		return "[unknown separator]"
	}
}

type Expression struct {
	Left  ExpressionType
	Sep   ExprSep
	Right ExpressionType
}

func (e Expression) expression() {}

func (e Expression) String() string {
	return fmt.Sprintf("(%s) %s (%s)", e.Left.String(), e.Sep.String(), e.Right.String())
}

type Num float64

func (n Num) expression() {}

func (n Num) String() string {
	return fmt.Sprint(float64(n))
}
