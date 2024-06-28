package internal

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// HashPassword hashes the provided pass with some salt
func HashPassword(password string, salt string) (string, error) {
	hasher := sha256.New()
	_, err := hasher.Write([]byte(password))
	if err != nil {
		return "", err
	}

	hashedPassword := hasher.Sum([]byte(salt))
	hashedPasswordStr := hex.EncodeToString(hashedPassword)

	return hashedPasswordStr, nil
}

func GenerateRandomSalt() (string, error) {
	salt := make([]byte, 16)

	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("error generating random salt: %w", err)
	}

	saltStr := hex.EncodeToString(salt)

	return saltStr, nil
}
