package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// gerar um enconde de uma string aleatória
func GenerateTokenKey(email string) string {
	r := make([]byte, 32)
	rand.Read(r)
	return base64.URLEncoding.EncodeToString(r)
}
