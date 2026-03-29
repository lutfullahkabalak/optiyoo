package handlers

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func isBcryptHash(stored string) bool {
	return strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$")
}

// passwordsMatch returns true if plain matches stored (bcrypt or legacy plaintext).
func passwordsMatch(stored, plain string) bool {
	if stored == "" {
		return false
	}
	if isBcryptHash(stored) {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)) == nil
	}
	return stored == plain
}
