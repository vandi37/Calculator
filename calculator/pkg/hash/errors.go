package hash

import "errors"

var (
	InvalidPassword = errors.New("invalid password")
	InvalidBase64   = errors.New("invalid base64")
)
