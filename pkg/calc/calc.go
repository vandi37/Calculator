package calc

import (
	"errors"
	"net/http"
	"strings"

	"github.com/vandi37/Calculator/pkg/parsing/lexer"
	"github.com/vandi37/Calculator/pkg/parsing/parser"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
	"github.com/vandi37/vanerrors"
)

var allowed []string = []string{
	lexer.ItIsNotANumber,
	lexer.UnexpectedChar,
	parser.UnexpectedTokenKind,
	parser.UnexpectedEOF,
	parser.UnexpectedToken,
	parser.ExpectedKind,
	DivisionByZero,
}

func GetCode(target error) int {
	for _, s := range allowed {
		err := vanerrors.Simple(s)
		if errors.Is(target, err) {
			return http.StatusUnprocessableEntity
		}
	}
	return http.StatusInternalServerError
}
func Pre(expression string) (tree.Ast, error) {
	expression = strings.Replace(expression, " ", "", -1)
	lexer := lexer.New([]rune(expression))
	tokens, err := lexer.GetTokens()
	if err != nil {
		return tree.Ast{}, err
	}

	parser := parser.New(tokens)
	return parser.Ast()
}

type Calculator struct {
	send func(float64, float64, tree.ExprSep) (chan float64, error)
}

func New(send func(float64, float64, tree.ExprSep) (chan float64, error)) *Calculator {
	return &Calculator{send}
}

func (c *Calculator) Expression(expression tree.ExpressionType) (float64, error) {
	if expression == nil {
		return 0, vanerrors.New(UnknownExpression, "<nil>")
	}
	n, ok := expression.(tree.Num)
	if ok {
		return float64(n), nil
	}
	u, ok := expression.(tree.UnaryMinus)
	if ok {
		f, err := c.Expression(u.E)
		return -f, err
	}

	b, ok := expression.(tree.Expression)
	if !ok {
		return 0, vanerrors.New(UnknownExpression, expression.String())
	}

	left, err := c.Expression(b.Left)
	if err != nil {
		return 0, err
	}
	right, err := c.Expression(b.Right)
	if err != nil {
		return 0, err
	}

	if right == 0 && b.Sep == tree.Division {
		return 0, vanerrors.Simple(DivisionByZero)
	}

	getter, err := c.send(left, right, b.Sep)
	if err != nil {
		return 0, err
	}

	return <-getter, nil
}

func Calc(ast tree.Ast, send func(float64, float64, tree.ExprSep) (chan float64, error)) (float64, error) {
	return New(send).Expression(ast.ExpressionType)
}
