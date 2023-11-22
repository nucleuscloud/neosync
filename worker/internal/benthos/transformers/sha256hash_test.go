package transformers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateSHA256Hash(t *testing.T) {

	res, err := GenerateRandomSHA256Hash()
	assert.NoError(t, err)

	assert.IsType(t, "", res)
}
