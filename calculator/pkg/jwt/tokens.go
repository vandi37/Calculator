package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	secret   []byte
	exr, nbf time.Duration
}

func New(secret string, exr, nbf time.Duration) *TokenService {
	return &TokenService{
		[]byte(secret),
		exr, nbf,
	}
}

func (s *TokenService) GetSecret() []byte {
	return s.secret
}

func (s *TokenService) GetExpiration() time.Duration {
	return s.exr
}

func (s *TokenService) GetNotBefore() time.Duration {
	return s.nbf
}

type Data struct {
	Key   string
	Value any
}

func (s *TokenService) Generate(sub string, data ...Data) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": sub,
		"aud": []string{"https://github.com/vandi37/Calculator"}, // The source code of the project (Expected using token only in the current service)
		"nbf": now.Add(s.nbf).Unix(),
		"iat": now.Unix(),
		"exp": now.Add(s.exr).Unix(),
		"iss": "https://github.com/vandi37/Calculator", // The source code of the project
	}
	for _, d := range data {
		claims[d.Key] = d.Value
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.secret)
}

func (s *TokenService) Parse(token string) (*jwt.Token, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	return parsedToken, nil
}
