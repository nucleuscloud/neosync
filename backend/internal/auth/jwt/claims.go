package authjwt

import (
	"context"
)

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope       string   `json:"scope"`
	Permissions []string `json:"permissions,omitempty"`
}

// Validate does nothing for this example, but we need
// it to satisfy validator.CustomClaims interface.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}
