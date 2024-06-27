package utils

import (
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("falha ao gerar o hash da senha")
	}
	return string(hash), nil
}

func ValidatePassword(hashPassword, password string) bool { //FIXME: this is a check password and not a validate
	return bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password)) == nil
}

func IsEmailValid(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9,\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(email)
}

func IsPasswordValid(password string) bool {
	// Check if the password has at least 6 characters and contains only letters and numbers
	hasLength := len(password) > 5

	// Check if the password contains at least one letter
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)

	// Check if the password contains at least one number
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)

	return hasLength && hasLetter && hasNumber
}
