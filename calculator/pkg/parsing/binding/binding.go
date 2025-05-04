// This package is used to understand the binding power in an ast tree
//
// It doesn't need test just a model package
package binding

import "github.com/vandi37/Calculator/pkg/parsing/tokens"

type Power int

const (
	Lowest Power = iota
	Additive
	Multiplicative
)

func GetPower(kind tokens.TokenKind) (Power, bool) {
	switch kind {
	case tokens.Addition, tokens.Subtraction:
		return Additive, true
	case tokens.Multiplication, tokens.Division:
		return Multiplicative, true
	default:
		return Lowest, false
	}
}
