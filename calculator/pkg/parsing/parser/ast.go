// In this package tokens are parsed to and ast
package parser

import (
	"slices"
	"unicode"

	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/pkg/parsing/binding"
	"github.com/vandi37/Calculator/pkg/parsing/lexer"
	"github.com/vandi37/Calculator/pkg/parsing/tokens"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
	"github.com/vandi37/vanerrors"
)

type Parser struct {
	t []tokens.Token
}

func RemoveWhitespace(s string) string {
	return string(slices.DeleteFunc([]rune(s), unicode.IsSpace))
}

func Build(expression string) (tree.Ast, error) {
	expression = RemoveWhitespace(expression)
	lexer := lexer.New([]rune(expression))
	tokens, err := lexer.GetTokens()
	if err != nil {
		return tree.Ast{}, err
	}

	parser := New(tokens)
	return parser.Ast()
}

func (p *Parser) Ast() (tree.Ast, error) {
	expr, err := p.Expression(binding.Lowest)
	if err == nil && len(p.t) > 0 {
		err = vanerrors.New(UnexpectedToken, p.t[0].String())
	}
	return tree.Ast{Expression: expr}, err
}

func New(t []tokens.Token) Parser {
	return Parser{t}
}

func (p *Parser) Next() (tokens.Token, bool) {
	if len(p.t) <= 0 {
		return tokens.EOFToken, false
	}
	t := p.t[0]
	p.t = p.t[1:]
	return t, true
}

func (p *Parser) Peek() (tokens.Token, bool) {
	if len(p.t) <= 0 {
		return tokens.EOFToken, false
	}
	return p.t[0], true
}

func (p *Parser) Move() {
	if len(p.t) <= 0 {
		return
	}
	p.t = p.t[1:]
}

func (p *Parser) ExpectKindError(kind tokens.TokenKind) error {
	if len(p.t) > 0 && p.t[0].Kind == kind {
		p.Move()
		return nil
	}
	return vanerrors.New(ExpectedKind, kind.String())
}

func (p *Parser) PrimExpression() (tree.ExpressionType, error) {
	t, ok := p.Next()
	if !ok {
		return nil, vanerrors.Simple(UnexpectedEOF)
	}

	switch t.Kind {
	case tokens.Addition:
		return p.PrimExpression()
	case tokens.Number:
		return tree.Num(t.Value), nil
	case tokens.Subtraction:
		right, err := p.PrimExpression()
		if err != nil {
			return nil, err
		}
		return tree.Expression{Left: tree.Num(0), Operation: tree.Operation(pb.Operation_SUBTRACT), Right: right}, nil
	case tokens.BracketOpen:
		expr, err := p.Expression(binding.Lowest)
		if err != nil {
			return nil, err
		}
		return expr, p.ExpectKindError(tokens.BracketClose)
	default:
		return nil, vanerrors.New(UnexpectedTokenKind, t.Kind.String())
	}
}

func (p *Parser) Expression(bp binding.Power) (tree.ExpressionType, error) {
	left, err := p.PrimExpression()
	if err != nil {
		return nil, err
	}
	for t, ok := p.Peek(); ok; t, ok = p.Peek() {
		if t == tokens.EOFToken {
			break
		}
		power, ok := binding.GetPower(t.Kind)
		if !ok || power <= bp {
			break
		}
		sep, ok := tree.SepFrom(t.Kind)
		if !ok {
			break
		}
		p.Move()
		left, err = p.BinExpr(left, sep, power)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *Parser) BinExpr(left tree.ExpressionType, operation tree.Operation, bp binding.Power) (tree.ExpressionType, error) {
	right, err := p.Expression(bp)
	if err != nil {
		return nil, err
	}
	return tree.Expression{
		Left:      left,
		Operation: operation,
		Right:     right,
	}, nil
}
