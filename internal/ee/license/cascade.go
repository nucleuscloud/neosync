package license

import "time"

type CascadeLicense struct {
	isValid   bool
	expiresAt time.Time
}

// Checks multiple licenses in input order to see if any are valid
func NewCascadeLicense(licenses ...EEInterface) *CascadeLicense {
	isValid := false
	expiresAt := time.Time{}
	for _, l := range licenses {
		if l.IsValid() {
			isValid = true
			expiresAt = l.ExpiresAt()
			break
		}
	}
	return &CascadeLicense{isValid: isValid, expiresAt: expiresAt}
}

func (c *CascadeLicense) IsValid() bool {
	return c.isValid
}

func (c *CascadeLicense) ExpiresAt() time.Time {
	return c.expiresAt
}
