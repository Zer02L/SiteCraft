package ai

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
)

const encPrefix = "enc:v1:"

// getMasterKey derives a 32-byte key for AES-256 from environment or built-in default seed.
func getMasterKey() []byte {
	seed := os.Getenv("SITECRAFT_SECRET_KEY")
	if seed == "" {
		seed = os.Getenv("ADMIN_SECRET_KEY")
	}
	if seed == "" {
		// Stable default seed for local/standalone instances
		seed = "sitecraft-default-ai-encryption-master-key-seed-v1"
	}
	hash := sha256.Sum256([]byte(seed))
	return hash[:]
}

// Encrypt encrypts plaintext using AES-256-GCM and returns a base64 string with prefix "enc:v1:".
// If the input is empty, an empty string is returned.
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	// If already encrypted, return as is
	if strings.HasPrefix(plaintext, encPrefix) {
		return plaintext, nil
	}

	block, err := aes.NewCipher(getMasterKey())
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return encPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts an AES-256-GCM encrypted string.
// If the input doesn't start with "enc:v1:", it returns the raw input (graceful backward compatibility).
func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	if !strings.HasPrefix(ciphertext, encPrefix) {
		// Not encrypted, return plaintext directly
		return ciphertext, nil
	}

	raw := strings.TrimPrefix(ciphertext, encPrefix)
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(getMasterKey())
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, encryptedMessage := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
