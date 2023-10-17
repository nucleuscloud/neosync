package neosync_transformers

import (
	"fmt"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func TestNullTransformer(t *testing.T) {
	mapping := `root = transformernull()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the null transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	fmt.Println("res", res)

	assert.Equal(t, res, "null", "Generated phone number must be the same length as the input phone number")
}
