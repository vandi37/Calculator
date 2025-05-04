package service

import (
	"errors"
	"net/http"

	"github.com/vandi37/Calculator/internal/repo"
	"github.com/vandi37/Calculator/pkg/hash"
	"github.com/vandi37/Calculator/pkg/parsing/lexer"
	"github.com/vandi37/Calculator/pkg/parsing/parser"
	"github.com/vandi37/vanerrors"
)

var (
	InvalidToken = errors.New("invalid token")
	Closed       = errors.New("closed")
)

var unprocessableEntity []string = []string{
	lexer.ItIsNotANumber,
	lexer.UnexpectedChar,
	parser.UnexpectedTokenKind,
	parser.UnexpectedEOF,
	parser.UnexpectedToken,
	parser.ExpectedKind,
}

func GetCode(target error) int {
	if errors.Is(target, repo.UsernameTaken) {
		return http.StatusConflict
	} else if errors.Is(target, repo.UserNotFound) ||
		errors.Is(target, repo.NodeNotFound) ||
		errors.Is(target, repo.ExpressionNotFound) {
		return http.StatusNotFound
	} else if errors.Is(target, repo.InvalidExpression) ||
		errors.Is(target, repo.InvalidNode) ||
		errors.Is(target, hash.InvalidBase64) ||
		errors.Is(target, InvalidToken) {
		return http.StatusBadRequest
	} else if errors.Is(target, hash.InvalidPassword) {
		return http.StatusUnauthorized
	}
	for _, s := range unprocessableEntity {
		err := vanerrors.Simple(s)
		if errors.Is(target, err) {
			return http.StatusUnprocessableEntity
		}
	}
	return http.StatusInternalServerError
}
