package sshtunnel

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"golang.org/x/crypto/ssh"
)

func getPrivateKeyAuthMethod(keyBytes []byte, passphrase *string) (ssh.AuthMethod, error) {
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

func parseSshKey(keyString string) (ssh.PublicKey, error) {
	// Parse the key
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyString)) //nolint
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	return publicKey, nil
}

// Auth Method is optional and will return nil if there is no valid method.
// Will only return error if unable to parse the private key into an auth method
func getTunnelAuthMethodFromSshConfig(auth *mgmtv1alpha1.SSHAuthentication) (ssh.AuthMethod, error) {
	if auth == nil {
		return nil, nil
	}
	switch config := auth.AuthConfig.(type) {
	case *mgmtv1alpha1.SSHAuthentication_Passphrase:
		return ssh.Password(config.Passphrase.Value), nil
	case *mgmtv1alpha1.SSHAuthentication_PrivateKey:
		authMethod, err := getPrivateKeyAuthMethod([]byte(config.PrivateKey.Value), config.PrivateKey.Passphrase)
		if err != nil {
			return nil, err
		}
		return authMethod, nil
	default:
		return nil, nil
	}
}

type DtoTunnelConfig struct {
	Addr         string
	ClientConfig *ssh.ClientConfig
}

// Converts the proto SSHTunnel into a config that can be plugged in to ssh.Dial
func GetTunnelConfigFromSSHDto(tunnel *mgmtv1alpha1.SSHTunnel) (*DtoTunnelConfig, error) {
	if tunnel == nil {
		return nil, errors.New("tunnel config is nil")
	}

	hostcallback, err := buildHostKeyCallback(tunnel)
	if err != nil {
		return nil, fmt.Errorf("unable to build host key callback: %w", err)
	}

	authmethod, err := getTunnelAuthMethodFromSshConfig(tunnel.GetAuthentication())
	if err != nil {
		return nil, fmt.Errorf("unable to parse ssh auth method: %w", err)
	}

	authmethods := []ssh.AuthMethod{}
	if authmethod != nil {
		authmethods = append(authmethods, authmethod)
	}
	return &DtoTunnelConfig{
		Addr: getSshAddr(tunnel),
		ClientConfig: &ssh.ClientConfig{
			User:            tunnel.GetUser(),
			Auth:            authmethods,
			HostKeyCallback: hostcallback,
			Timeout:         15 * time.Second, // todo: make configurable
		},
	}, nil
}

func getSshAddr(tunnel *mgmtv1alpha1.SSHTunnel) string {
	host := tunnel.GetHost()
	port := tunnel.GetPort()
	if port > 0 {
		return net.JoinHostPort(host, strconv.FormatInt(int64(port), 10))
	}
	return host
}

func buildHostKeyCallback(tunnel *mgmtv1alpha1.SSHTunnel) (ssh.HostKeyCallback, error) {
	if tunnel.GetKnownHostPublicKey() != "" {
		publickey, err := parseSshKey(tunnel.GetKnownHostPublicKey())
		if err != nil {
			return nil, fmt.Errorf("unable to parse ssh known host public key: %w", err)
		}
		return ssh.FixedHostKey(publickey), nil
	} else {
		return ssh.InsecureIgnoreHostKey(), nil //nolint:gosec // the user has chosen to not provide a known host public key
	}
}
