package do

import "errors"

var (
	DivisionByZero   = errors.New("division by zero")
	UnknownOperation = errors.New("unknown operation")
)
