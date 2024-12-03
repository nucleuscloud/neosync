package license

type CascadeLicense struct {
	isValid bool
}

// Checks multiple licenses in input order to see if any are valid
func NewCascadeLicense(licenses ...EEInterface) *CascadeLicense {
	isValid := false
	for _, l := range licenses {
		if l.IsValid() {
			isValid = true
			break
		}
	}
	return &CascadeLicense{isValid: isValid}
}

func (c *CascadeLicense) IsValid() bool {
	return c.isValid
}
