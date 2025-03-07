// This package is for getting tokens from plain text
//
// It is not life-needed for the project it is just a method to parse text
package lexer

import (
	"fmt"
	"strconv"

	"github.com/vandi37/Calculator/pkg/parsing"
	"github.com/vandi37/Calculator/pkg/parsing/tokens"
	"github.com/vandi37/vanerrors"
)

type Lexer struct {
	v []rune
}

func New(v []rune) Lexer {
	return Lexer{v}
}

func (l *Lexer) IsEmpty() bool {
	return len(l.v) <= 0
}

func (l *Lexer) nextRune() (rune, error) {
	r := l.v[0]
	l.v = l.v[1:]
	return r, nil
}

func (l *Lexer) backRune(r rune) {
	l.v = append([]rune{r}, l.v...)
}

func (l *Lexer) GetTokens() ([]tokens.Token, error) {
	t := []tokens.Token{}
	for len(l.v) > 0 {
		token, err := l.Next()
		if err != nil {
			return nil, err
		}
		if token == tokens.EOFToken {
			break
		}
		t = append(t, token)
	}
	return t, nil
}

func (l *Lexer) Next() (tokens.Token, error) {
	if l.IsEmpty() {
		return tokens.EOFToken, nil
	}
	t := tokens.EmptyToken
	r, err := l.nextRune()
	if err != nil {
		return t, err
	}

	switch r {
	case '+':
		t.Kind = tokens.Addition
	case '-':
		t.Kind = tokens.Subtraction
	case '*':
		t.Kind = tokens.Multiplication
	case '/':
		t.Kind = tokens.Division
	case '(':
		t.Kind = tokens.BracketOpen
	case ')':
		t.Kind = tokens.BracketClose
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		t.Kind = tokens.Number
		l.backRune(r)
		start, ok, err := l.buildInt()
		if err != nil {
			return t, err
		}
		if !ok {
			v, err := strconv.Atoi(start)
			if err != nil {
				return t, vanerrors.Wrap(parsing.UnknownParsingError, err)
			}
			t.Value = float64(v)
			break
		}

		after, ok, err := l.buildInt()
		if err != nil {
			return t, err
		}

		if ok {
			return t, vanerrors.New(UnexpectedChar, fmt.Sprintf("%c", r))
		}
		v, err := strconv.ParseFloat(start+"."+after, 64)
		if err != nil {
			return t, vanerrors.Wrap(parsing.UnknownParsingError, err)
		}
		t.Value = v
	default:
		return t, vanerrors.New(UnexpectedChar, fmt.Sprintf("%c", r))
	}
	return t, nil
}

func IsNum(r rune) bool {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	default:
		return false
	}
}

func (l *Lexer) buildInt() (string, bool, error) {
	current := []rune{}
	ok := false
	for {
		if l.IsEmpty() {
			break
		}
		next, err := l.nextRune()
		if err != nil {
			return "", false, err
		}
		if next == '.' || next == ',' {
			ok = true
			break
		}

		if !IsNum(next) {
			l.backRune(next)
			break
		}

		current = append(current, next)
	}

	return string(current), ok, nil
}
