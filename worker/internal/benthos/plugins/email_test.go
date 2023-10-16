package neosync_plugins

import (
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessEmail(t *testing.T) {

	tests := []struct {
		email               string
		preserveLength      bool
		preserveDomain      bool
		expectedError       bool
		expectedLength      int
		checkEmailStructure bool
	}{
		{"evis@gmail.com", true, false, false, 14, false}, // checks preserve length
		{"evis@gmail.com", false, true, false, 0, true},   // checks preserve domain
		{"evis@gmail.com", true, true, false, 14, false},  // checks preserve length and domain
	}

	for _, tt := range tests {
		res, err := ProcessEmail(tt.email, tt.preserveLength, tt.preserveDomain)

		if tt.expectedError {
			assert.Error(t, err)
			continue
		}

		assert.NoError(t, err)

		if tt.expectedLength > 0 {
			assert.Equal(t, tt.expectedLength, len(res))
		}

		if tt.preserveDomain {
			assert.Equal(t, strings.Split(res, "@")[1], "gmail.com")
		}

		if tt.preserveDomain && tt.preserveLength {
			assert.Equal(t, strings.Split(res, "@")[1], "gmail.com")
			assert.Equal(t, tt.expectedLength, len(res))

		}

	}

}

func TestEmailTransformer(t *testing.T) {
	mapping := `root = this.emailtransformer(true, true)`
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the uuid transformer")

	testVal := "evis@gmail.com"

	res, err := ex.Query(testVal)
	assert.NoError(t, err)

	assert.Len(t, res.(string), len(testVal), "Generated email must be the same length as the input email")
	assert.Equal(t, strings.Split(res.(string), "@")[1], "gmail.com")
}
