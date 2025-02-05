package transformers

import (
	"fmt"
	"math/rand"
	"testing"

	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/require"
)

func Test_GenerateRandomEmailShort(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	shortMaxLength := int64(15)

	res, err := generateRandomEmail(randomizer, shortMaxLength, GenerateEmailType_FullName, []string{})

	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, true, transformer_utils.IsValidEmail(res), fmt.Sprintf(`The expected email should be have a valid email format. Received:%s`, res))
	require.LessOrEqual(t, int64(len(res)), shortMaxLength, fmt.Sprintf("The email should be less than or equal to the max length. This is the error email:%s", res))
}

func Test_GenerateRandomEmail(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))

	res, err := generateRandomEmail(randomizer, int64(40), GenerateEmailType_FullName, []string{})

	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, true, transformer_utils.IsValidEmail(res), fmt.Sprintf(`The expected email should be have a valid email format. Received:%s`, res))
	require.LessOrEqual(t, int64(len(res)), int64(40), fmt.Sprintf("The email should be less than or equal to the max length. This is the error email:%s", res))
}

func Test_GenerateRandomEmail_Uuid(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))

	res, err := generateRandomEmail(randomizer, int64(40), GenerateEmailType_UuidV4, []string{})

	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, true, transformer_utils.IsValidEmail(res), fmt.Sprintf(`The expected email should be have a valid email format. Received:%s`, res))
	require.LessOrEqual(t, int64(len(res)), int64(40), fmt.Sprintf("The email should be less than or equal to the max length. This is the error email:%s", res))
}
func Test_GenerateRandomEmail_Uuid_Small(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))

	res, err := generateRandomEmail(randomizer, int64(8), GenerateEmailType_UuidV4, []string{})

	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, true, transformer_utils.IsValidEmail(res), fmt.Sprintf(`The expected email should be have a valid email format. Received:%s`, res))
	require.LessOrEqual(t, int64(len(res)), int64(40), fmt.Sprintf("The email should be less than or equal to the max length. This is the error email:%s", res))
}

func Test_RandomEmailTransformer(t *testing.T) {
	maxLength := int64(40)
	mapping := fmt.Sprintf(`root = generate_email(max_length:%d, email_type:%q)`, maxLength, GenerateEmailType_UuidV4)
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err)
	require.NotEmpty(t, ex)

	res, err := ex.Query(nil)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	resStr, ok := res.(string)
	require.True(t, ok)
	require.NotEmpty(t, resStr)

	require.LessOrEqual(t, int64(len(resStr)), maxLength, fmt.Sprintf("The email should be less than or equal to the max length. This is the error email:%s", res))

	require.Equal(t, true, transformer_utils.IsValidEmail(res.(string)), "The expected email should have a valid email format")
}

func Test_RandomEmailTransformer_NoOptions(t *testing.T) {
	mapping := `root = generate_email()`
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err)
	require.NotEmpty(t, ex)

	res, err := ex.Query(nil)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	resStr, ok := res.(string)
	require.True(t, ok)
	require.NotEmpty(t, resStr)

	require.NotEmptyf(t, resStr, fmt.Sprintf("The email should be less than or equal to the max length. This is the error email:%s", res))
}
