package hash

import (
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	password []byte
}

func ParsePassword(password string) (Password, error) {
	if len(password) > 72 {
		return Password{}, InvalidPassword
	}
	return Password{password: []byte(password)}, nil
}

func (p *Password) GetBytes() []byte {
	clone := make([]byte, len(p.password))
	copy(clone, p.password)
	return clone
}

type PasswordService struct {
	cost int
}

func (s *PasswordService) HashPassword(password string) (hash string, err error) {
	var parsed Password
	parsed, err = ParsePassword(password)
	if err != nil {
		return
	}

	byteHash, err := bcrypt.GenerateFromPassword(parsed.GetBytes(), s.cost)
	if err != nil {
		return
	}
	hash = base64.StdEncoding.EncodeToString(byteHash)
	return
}

func (s *PasswordService) ComparePassword(password, hash string) error {
	var parsed Password
	parsed, err := ParsePassword(password)
	if err != nil {
		return err
	}

	byteHash, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return InvalidBase64
	}

	err = bcrypt.CompareHashAndPassword(byteHash, parsed.GetBytes())
	return err
}

func (s *PasswordService) GetCost() int {
	return s.cost
}

func NewPasswordService(cost *int) *PasswordService {
	costData := bcrypt.DefaultCost
	if cost != nil {
		costData = *cost
	}
	return &PasswordService{
		cost: costData,
	}
}
