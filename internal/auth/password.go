package auth

// Implementação de funções para uso do Argon2id no hash e verificação de senha

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// Parâmetros seguros para Argon2id
const (
	argonTime    uint32 = 1
	argonMemory  uint32 = 64 * 1024 // 64 MB
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

// passwordPepper é um segredo adicional aplicado às senhas antes do hash.
// Configurado via variável de ambiente PASSWORD_PEPPER.
var passwordPepper = os.Getenv("PASSWORD_PEPPER")

func addPepper(password string) string {
	if passwordPepper == "" {
		return password
	}
	return password + passwordPepper
}

func HashPasswordArgon2(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("falha ao gerar salt: %w", err)
	}

	hash := argon2.IDKey([]byte(addPepper(password)), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argonMemory, argonTime, argonThreads, b64Salt, b64Hash)
	return encoded, nil
}

func CheckPasswordArgon2(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false
	}

	params := strings.Split(parts[3], ",")
	var memory uint32
	var time uint32
	var threads uint8

	for _, p := range params {
		kvs := strings.SplitN(p, "=", 2)
		if len(kvs) != 2 {
			continue
		}
		switch kvs[0] {
		case "m":
			fmt.Sscanf(kvs[1], "%d", &memory)
		case "t":
			fmt.Sscanf(kvs[1], "%d", &time)
		case "p":
			n := 0
			fmt.Sscanf(kvs[1], "%d", &n)
			threads = uint8(n)
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	testHash := argon2.IDKey([]byte(addPepper(password)), salt, time, memory, threads, uint32(len(hash)))
	return subtleCompare(hash, testHash)
}

// HashPassword é a função de alto nível usada pelo resto do sistema.
// Ela sempre gera hashes usando Argon2id para novos passwords.
func HashPassword(password string) (string, error) {
	return HashPasswordArgon2(password)
}

// CheckPassword detecta automaticamente o tipo de hash armazenado e
// utiliza o verificador apropriado (Argon2id ou o algoritmo legado).
func CheckPassword(password string, passwordHash string) bool {
	// Hashes Argon2id gerados por HashPasswordArgon2 sempre iniciam com este prefixo.
	if strings.HasPrefix(passwordHash, "$argon2id$") {
		return CheckPasswordArgon2(password, passwordHash)
	}

	// Caso não seja Argon2id, assumimos hash legado (por exemplo, bcrypt).
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return false
	}
	return true
}

// subtleCompare faz uma comparação em tempo constante entre dois slices.
func subtleCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
