package auth

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateResetToken gera um token aleatório seguro para reset de senha.
func GenerateResetToken() (string, error) {
	bytes := make([]byte, 32) // 64 caracteres hexadecimais
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
