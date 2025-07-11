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

const (
	eeLicenseEvKey = "EE_LICENSE"
)

// The expected base64 decoded structure of the EE_LICENSE file
type licenseFile struct {
	License   string `json:"license"`
	Signature string `json:"signature"`
}

type EEInterface interface {
	IsValid() bool
	ExpiresAt() time.Time
}

var _ EEInterface = (*EELicense)(nil)

type EELicense struct {
	contents *licenseContents
}

func (e *EELicense) IsValid() bool {
	return e.contents != nil && e.contents.IsValid()
}

func (e *EELicense) ExpiresAt() time.Time {
	if e.contents == nil {
		return time.Now().UTC()
	}
	return e.contents.ExpiresAt
}

type ValidLicense struct {
}

func (v *ValidLicense) IsValid() bool {
	return true
}

func (v *ValidLicense) ExpiresAt() time.Time {
	return time.Now().UTC().Add(time.Hour * 24 * 365 * 10)
}

func NewValidLicense() *ValidLicense {
	return &ValidLicense{}
}

// Retrieves the EE license from the environment
// If not enabled, will still return valid struct.
// Errors if not able to properly parse a provided EE license from the environment
func NewFromEnv() (*EELicense, error) {
	lc, _, err := getLicenseFromEnv()
	if err != nil {
		return nil, err
	}
	return newFromLicenseContents(lc), nil
}

func newFromLicenseContents(contents *licenseContents) *EELicense {
	return &EELicense{contents: contents}
}

// The expected base64 decoded structure of the EE_LICENSE.contents file
type licenseContents struct {
	Version    string    `json:"version"`
	Id         string    `json:"id"`
	IssuedTo   string    `json:"issued_to"`
	CustomerId string    `json:"customer_id"`
	IssuedAt   time.Time `json:"issued_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (l *licenseContents) IsValid() bool {
	return time.Now().UTC().Before(l.ExpiresAt)
}

// Retrieves the EE license from the environment
func getLicenseFromEnv() (*licenseContents, bool, error) {
	input := viper.GetString(eeLicenseEvKey)
	if input == "" {
		return nil, false, nil
	}
	pk, err := parsePublicKey(publicKeyPEM)
	if err != nil {
		return nil, false, fmt.Errorf("unable to parse ee public key: %w", err)
	}
	contents, err := getLicense(input, pk)
	if err != nil {
		return nil, false, fmt.Errorf("failed to parse provided ee license: %w", err)
	}
	return contents, true, nil
}

// Expected the license data to be a base64 encoded json string that matches the licenseFile structure.
func getLicense(licenseData string, publicKey ed25519.PublicKey) (*licenseContents, error) {
	licenseDataContents, err := base64.StdEncoding.DecodeString(licenseData)
	if err != nil {
		return nil, fmt.Errorf("unable to decode license data: %w", err)
	}

	var license licenseFile
	err = json.Unmarshal(licenseDataContents, &license)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal license data from input: %w", err)
	}
	contents, err := base64.StdEncoding.DecodeString(license.License)
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
		return nil, fmt.Errorf(
			"contents verified, but unable to unmarshal license contents from input: %w",
			err,
		)
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
