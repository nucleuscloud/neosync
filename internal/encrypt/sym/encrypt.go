package sym_encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

var (
	ErrEmptyPassword  = errors.New("password cannot be empty")
	ErrEmptyPlaintext = errors.New("plaintext cannot be empty")
	ErrInvalidCipher  = errors.New("invalid ciphertext format")
)

type Encryptor struct {
	aead cipher.AEAD
}

type Interface interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

func NewEncryptor(password string) (*Encryptor, error) {
	if password == "" {
		return nil, ErrEmptyPassword
	}

	aesCipher, err := aes.NewCipher(deriveKey(password))
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	return &Encryptor{aead: aead}, nil
}

func deriveKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", ErrEmptyPlaintext
	}

	nonce := make([]byte, e.aead.NonceSize())
	_, err := rand.Read(nonce)
	if err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := e.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	nonceSize := e.aead.NonceSize()
	if len(decoded) < nonceSize {
		return "", ErrInvalidCipher
	}

	nonce, decodedCiphertext := decoded[:nonceSize], decoded[nonceSize:]
	plaintext, err := e.aead.Open(nil, nonce, decodedCiphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
