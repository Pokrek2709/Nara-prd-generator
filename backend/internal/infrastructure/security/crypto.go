package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrInvalidCiphertext = errors.New("invalid or corrupted ciphertext")
	ErrDecryptionFailed  = errors.New("decryption failed")
)

// Encrypt encrypts plaintext using AES-256-GCM with a key derived from secret.
// Returns base64-encoded nonce + ciphertext.
func Encrypt(plaintext, secret string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	if secret == "" {
		return "", errors.New("encryption secret cannot be empty")
	}

	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt decrypts a base64-encoded AES-256-GCM ciphertext using secret.
// If the input is not a valid base64 or decryption fails, but looks like legacy unencrypted text (e.g. sk-...),
// it gracefully returns the original text to maintain backward compatibility.
func Decrypt(ciphertextBase64, secret string) (string, error) {
	if ciphertextBase64 == "" {
		return "", nil
	}
	if secret == "" {
		return "", errors.New("decryption secret cannot be empty")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		// Not base64, likely legacy plaintext key
		return ciphertextBase64, nil
	}

	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		// Too short for nonce + tag; return as plaintext if it contains standard key patterns
		if strings.HasPrefix(ciphertextBase64, "sk-") || strings.HasPrefix(ciphertextBase64, "gsk_") {
			return ciphertextBase64, nil
		}
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// If decryption failed, check if it was legacy plaintext that happened to be valid base64
		if strings.HasPrefix(ciphertextBase64, "sk-") || strings.HasPrefix(ciphertextBase64, "gsk_") {
			return ciphertextBase64, nil
		}
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}
