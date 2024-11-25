package license

import "github.com/spf13/viper"

// Conforms to the EE License for Neosync Cloud
type CloudLicense struct {
	isCloud bool
}

var _ EEInterface = (*CloudLicense)(nil)

func NewCloudLicense(isCloud bool) *CloudLicense {
	return &CloudLicense{isCloud: isCloud}
}

func NewCloudLicenseFromEnv() *CloudLicense {
	return &CloudLicense{isCloud: viper.GetBool("NEOSYNC_CLOUD")}
}

func (c *CloudLicense) IsValid() bool {
	return c.isCloud
}
