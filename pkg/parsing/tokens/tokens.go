// This package has token standards for my project
//
// The tokens are used in all parts of program so it is much easier to tack inside of the program
package tokens

import "fmt"

type TokenKind int

const (
	Number TokenKind = iota << iota // idk why i did it ))
	Addition
	Subtraction
	Multiplication
	Division
	BracketOpen
	BracketClose
	EOF = -2
)

func (k TokenKind) String() string {
	switch k {
	case Number:
		return "[number]"
	case Addition:
		return "[addition]"
	case Subtraction:
		return "[subtraction]"
	case Multiplication:
		return "[multiplication]"
	case Division:
		return "[division]"
	case BracketOpen:
		return "[opening bracket]"
	case BracketClose:
		return "[closing bracket]"
	case EOF:
		return "[eof]"
	default:
		return "[unknown kind]"
	}
}

type Token struct {
	Kind  TokenKind `json:"kind"`
	Value float64   `json:"value"`
}

func (t Token) String() string {
	if t.Kind == Number {
		return fmt.Sprint(t.Value)
	}
	return t.Kind.String()
}

var EmptyToken = Token{Kind: -1, Value: -1}
var EOFToken = Token{Kind: EOF, Value: -2}

func BuildToken(kind TokenKind, val float64) Token {
	return Token{kind, val}
}
