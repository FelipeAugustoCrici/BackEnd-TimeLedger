// Package crypto fornece criptografia AES-256-GCM para valores financeiros.
// Cada valor é cifrado com um nonce aleatório de 12 bytes e armazenado como
// base64url (nonce || ciphertext) para ser salvo como TEXT no Postgres.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

// deriveKey deriva uma chave AES-256 a partir da variável FIELD_ENCRYPT_KEY.
// Usa SHA-256 para garantir exatamente 32 bytes independente do tamanho da env.
func deriveKey() ([]byte, error) {
	raw := os.Getenv("FIELD_ENCRYPT_KEY")
	if raw == "" {
		return nil, errors.New("FIELD_ENCRYPT_KEY não definida")
	}
	h := sha256.Sum256([]byte(raw))
	return h[:], nil
}

// EncryptFloat64 cifra um float64 e retorna base64url(nonce || ciphertext).
func EncryptFloat64(value float64) (string, error) {
	key, err := deriveKey()
	if err != nil {
		return "", err
	}

	plaintext := []byte(strconv.FormatFloat(value, 'f', -1, 64))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes.NewCipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("cipher.NewGCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("rand nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// DecryptFloat64 descriptografa um valor produzido por EncryptFloat64.
func DecryptFloat64(encoded string) (float64, error) {
	key, err := deriveKey()
	if err != nil {
		return 0, err
	}

	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		// Tenta com padding (dados legados gravados com URLEncoding)
		if padded, err2 := base64.URLEncoding.DecodeString(encoded); err2 == nil {
			data = padded
		} else if v, parseErr := strconv.ParseFloat(encoded, 64); parseErr == nil {
			// Dados legados antes da migração (número puro)
			return v, nil
		} else {
			return 0, fmt.Errorf("base64 decode: %w", err)
		}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, fmt.Errorf("aes.NewCipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, fmt.Errorf("cipher.NewGCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return 0, errors.New("ciphertext muito curto")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return 0, fmt.Errorf("gcm.Open: %w", err)
	}

	return strconv.ParseFloat(string(plaintext), 64)
}
