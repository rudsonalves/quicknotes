package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type PasswordUtils interface {
	HashPassword(password string) (string, error)
}

type passwordUtils struct{}

func NewPasswordUtils() PasswordUtils {
	return &passwordUtils{}
}

func (passwordUtils) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("falha ao gerar o hash da senha")
	}
	return string(hash), nil
}
