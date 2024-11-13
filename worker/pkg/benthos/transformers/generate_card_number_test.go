package transformers

import (
	"strconv"
	"testing"
	"time"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/stretchr/testify/assert"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_GenerateValidLuhnCardNumber(t *testing.T) {
	val, err := generateValidLuhnCheckCardNumber(rng.New(time.Now().UnixNano()))

	assert.NoError(t, err)
	assert.Equal(t, len(strconv.FormatInt(val, 10)), 16, "The output card should be 16 characters long")
	assert.Equal(t, true, isValidLuhn(val), "The card number should pass luhn validation")
}

func Test_GenerateCardNumber(t *testing.T) {
	val, err := generateCardNumber(rng.New(time.Now().UnixNano()), false)

	assert.NoError(t, err)
	assert.Len(t, strconv.FormatInt(val, 10), 16, "The output card should be 16 characters long")
}

func Test_GenerateCardNumberTransformer(t *testing.T) {
	mapping := `root = generate_card_number(valid_luhn:true)`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Len(t, strconv.FormatInt(res.(int64), 10), 16, "The output card should be 16 characters long")
	assert.Equal(t, true, isValidLuhn(res.(int64)), "The output card number should pass luhn validation")
}

func Test_GenerateCardNumberTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_card_number()`
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the random int transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.Len(t, strconv.FormatInt(res.(int64), 10), 16, "The output card should be 16 characters long")
}

func isValidLuhn(cc int64) bool {
	return (cc%10+checksum(cc/10))%10 == 0
}

func checksum(number int64) int64 {
	var luhn int64

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 {
			cur *= 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number /= 10
	}
	return luhn % 10
}
