package sshtunnel

import (
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
