package lexer

import (
	"fmt"

	"github.com/vandi37/vanerrors"
)

const (
	ItIsNotANumber = "this is not a number"
	UnexpectedChar = "unexpected char"
)

func IsNotANumber(r rune) error {
	return vanerrors.New(ItIsNotANumber, fmt.Sprintf("char %c should be 0-9", r))

}
