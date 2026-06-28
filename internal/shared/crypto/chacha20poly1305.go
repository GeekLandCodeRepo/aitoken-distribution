package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

type ChaCha20Poly1305Crypto struct {
	key []byte
}

func NewChaCha20Poly1305Crypto(key string) *ChaCha20Poly1305Crypto {
	keyBytes := make([]byte, chacha20poly1305.KeySize)
	copy(keyBytes, []byte(key))
	return &ChaCha20Poly1305Crypto{key: keyBytes}
}

func (c *ChaCha20Poly1305Crypto) Encrypt(plaintext string) (string, error) {
	aead, err := chacha20poly1305.New(c.key)
	if err != nil {
		return "", fmt.Errorf("failed to create chacha20-poly1305 cipher: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (c *ChaCha20Poly1305Crypto) Decrypt(encoded string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	aead, err := chacha20poly1305.New(c.key)
	if err != nil {
		return "", fmt.Errorf("failed to create chacha20-poly1305 cipher: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
