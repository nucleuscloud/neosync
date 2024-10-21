package license

import (
	"crypto/ed25519"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

//go:embed neosync_ee_pub.pem
var publicKeyPEM string

// The expected base64 decoded structure of the EE_LICENSE file
type licenseFile struct {
	Contents  string `json:"contents"`
	Signature string `json:"signature"`
}

// The expecteed base64 decoded structure of the EE_LICENSE.contents file
type licenseContents struct {
	Version    string    `json:"version"`
	Id         string    `json:"id"`
	ExpiryDate time.Time `json:"expiry_date"`
}

// Retrieves the EE license from the environment
func GetLicenseFromEnv() (*licenseContents, bool, error) {
	input := viper.GetString("EE_LICENSE")
	if input == "" {
		return nil, false, nil
	}
	pk, err := parsePublicKey(publicKeyPEM)
	if err != nil {
		return nil, false, err
	}
	contents, err := getLicense(input, pk)
	if err != nil {
		return nil, false, err
	}
	return contents, true, nil
}

func getLicense(licenseData string, publicKey ed25519.PublicKey) (*licenseContents, error) {
	var license licenseFile
	err := json.Unmarshal([]byte(licenseData), &license)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal license data from input: %w", err)
	}
	contents, err := base64.StdEncoding.DecodeString(license.Contents)
	if err != nil {
		return nil, fmt.Errorf("unable to decode contents: %w", err)
	}
	signature, err := base64.StdEncoding.DecodeString(license.Signature)
	if err != nil {
		return nil, fmt.Errorf("unable to decode signature: %w", err)
	}
	ok := ed25519.Verify(publicKey, contents, signature)
	if !ok {
		return nil, errors.New("unable to verify contents against public key")
	}

	var lc licenseContents
	err = json.Unmarshal(contents, &lc)
	if err != nil {
		return nil, fmt.Errorf("contents verified, but unable to unmarshal license contents from input: %w", err)
	}

	return &lc, nil
}

func parsePublicKey(data string) (ed25519.PublicKey, error) {
	block, _ := pem.Decode([]byte(data))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the ee public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DER encoded public key: %v", err)
	}

	switch pub := pub.(type) {
	case ed25519.PublicKey:
		return pub, nil
	default:
		return nil, fmt.Errorf("unsupported public key: %T", pub)
	}
}
