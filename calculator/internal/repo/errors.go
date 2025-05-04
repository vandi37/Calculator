package repo

import (
	"errors"
)

var (
	UsernameTaken      = errors.New("username already taken")
	UserNotFound       = errors.New("user not found")
	NodeNotFound       = errors.New("node not found")
	ExpressionNotFound = errors.New("expression not found")
	InvalidExpression  = errors.New("invalid expression")
	InvalidNode        = errors.New("invalid node")
)
