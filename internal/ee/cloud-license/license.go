package cloudlicense

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

//go:embed neosync_cloud_pub.pem
var publicKeyPEM string

// The expected base64 decoded structure of the NEOSYNC_CLOUD_LICENSE file
type licenseFile struct {
	License   string `json:"license"`
	Signature string `json:"signature"`
}

type Interface interface {
	IsValid() bool
}

var _ Interface = (*CloudLicense)(nil)

type CloudLicense struct {
	contents *licenseContents
}

// Determines if Neosync Cloud is enabled.
// If not enabled, returns a valid struct where IsValid returns false
// If enabled but no license if provided, returns an error
func NewFromEnv() (*CloudLicense, error) {
	lc, isEnabled, err := getFromEnv()
	if err != nil {
		return nil, err
	}
	if !isEnabled {
		return &CloudLicense{contents: nil}, nil
	}
	return &CloudLicense{contents: lc}, nil
}

func (c *CloudLicense) IsValid() bool {
	return c.contents != nil && c.contents.IsValid()
}

type licenseContents struct {
	Version   string    `json:"version"`
	Id        string    `json:"id"`
	IssuedTo  string    `json:"issued_to"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (l *licenseContents) IsValid() bool {
	return time.Now().UTC().Before(l.ExpiresAt)
}

const (
	cloudLicenseEvKey = "NEOSYNC_CLOUD_LICENSE"
	cloudEnabledEvKey = "NEOSYNC_CLOUD"
)

func getFromEnv() (*licenseContents, bool, error) {
	isCloud := viper.GetBool(cloudEnabledEvKey)
	if !isCloud {
		return nil, false, nil
	}

	input := viper.GetString(cloudLicenseEvKey)
	if input == "" {
		return nil, false, fmt.Errorf("%s was true but no license was found at %s", cloudEnabledEvKey, cloudLicenseEvKey)
	}
	pk, err := parsePublicKey(publicKeyPEM)
	if err != nil {
		return nil, false, fmt.Errorf("unable to parse neosync cloud public key: %w", err)
	}
	contents, err := getLicense(input, pk)
	if err != nil {
		return nil, false, fmt.Errorf("failed to parse provided license: %w", err)
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
		return nil, fmt.Errorf("contents verified, but unable to unmarshal license contents from input: %w", err)
	}

	return &lc, nil
}

func parsePublicKey(data string) (ed25519.PublicKey, error) {
	block, _ := pem.Decode([]byte(data))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
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
