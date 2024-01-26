package sshtunnel

import (
	"fmt"

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
	// Parse the key
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyString)) //nolint
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	return publicKey, nil
}
