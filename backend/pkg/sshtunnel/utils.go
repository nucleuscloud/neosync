package sshtunnel

import (
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

func GetPrivateKeyAuthMethod(keyBytes []byte, passphrase *string) (ssh.AuthMethod, error) {
	if passphrase != nil && *passphrase != "" {
		return getEncryptedPrivateKeyAuthMethod(keyBytes, []byte(*passphrase))
	}
	return getPlaintextPrivateKeyAuthMethod(keyBytes)
}

func getEncryptedPrivateKeyAuthMethod(keyBytes, passphrase []byte) (ssh.AuthMethod, error) {
	signer, err := ssh.ParsePrivateKeyWithPassphrase(keyBytes, passphrase)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

func getPlaintextPrivateKeyAuthMethod(keyBytes []byte) (ssh.AuthMethod, error) {
	key, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}

func ParseSshKey(keyString string) (ssh.PublicKey, error) {
	// The key string is usually in the format "type base64-encoded-key".
	// First, decode the base64 part.
	parts := strings.Split(keyString, " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid key format")
	}
	keyBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("base64 decoding failed: %v", err)
	}

	// Parse the key
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey(keyBytes) //nolint
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	return publicKey, nil
}
