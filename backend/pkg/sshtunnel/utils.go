package sshtunnel

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
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

// Auth Method is optional and will return nil if there is no valid method.
// Will only return error if unable to parse the private key into an auth method
func GetTunnelAuthMethodFromSshConfig(auth *mgmtv1alpha1.SSHAuthentication) (ssh.AuthMethod, error) {
	if auth == nil {
		return nil, nil
	}
	switch config := auth.AuthConfig.(type) {
	case *mgmtv1alpha1.SSHAuthentication_Passphrase:
		return ssh.Password(config.Passphrase.Value), nil
	case *mgmtv1alpha1.SSHAuthentication_PrivateKey:
		authMethod, err := GetPrivateKeyAuthMethod([]byte(config.PrivateKey.Value), config.PrivateKey.Passphrase)
		if err != nil {
			return nil, err
		}
		return authMethod, nil
	default:
		return nil, nil
	}
}
