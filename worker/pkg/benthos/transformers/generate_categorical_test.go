package transformers

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateCategorical(t *testing.T) {
	categories := []string{"test", "me", "please", "sir"}

	res := generateCategorical(rng.New(time.Now().UnixNano()), categories)

	valueInCategory := false
	for _, cat := range categories {
		if cat == res {
			valueInCategory = true
			break
		}
	}

	assert.True(t, valueInCategory, "The generated caetgories should exist in the categories slice")
}

func Test_GenerateCategoricalTransformer(t *testing.T) {
	categories := []string{"test", "me", "please", "sir"}

	stringVal := strings.Join(categories, ",")

	mapping := fmt.Sprintf(`root = generate_categorical(categories:%q)`, stringVal)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the generate categories transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.IsType(t, "", res, "The returned city should be a string")

	valueInCategory := false
	for _, cat := range categories {
		if cat == res {
			valueInCategory = true
			break
		}
	}

	assert.True(t, valueInCategory, "The generated caetgories should exist in the categories slice")
}
