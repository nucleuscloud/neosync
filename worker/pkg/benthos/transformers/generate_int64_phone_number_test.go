package transformers

import (
	"strconv"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomIntPhoneNumber(t *testing.T) {
	res, err := generateRandomInt64PhoneNumber(rng.New(time.Now().UnixNano()))
	assert.NoError(t, err)

	numStr := strconv.FormatInt(res, 10)
	assert.Equal(t, len(numStr), 10, "The length of the output phone number should be the same as the input phone number")
}

func Test_GenerateRandomIntPhoneNumberTransformer(t *testing.T) {
	mapping := `root = generate_int64_phone_number()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the int64 phone transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	resStr := strconv.FormatInt(res.(int64), 10)

	assert.Equal(t, len(resStr), 10, "The length of the output phone number should be the same as the input phone number")
	assert.IsType(t, "", resStr, "The actual value type should be a string")
}
