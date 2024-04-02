package transformers

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func Test_GenerateRandomEmailShort(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	shortMaxLength := int64(15)

	res, err := generateRandomEmail(randomizer, shortMaxLength, fullNameEmailType, []string{})

	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, true, transformer_utils.IsValidEmail(res), fmt.Sprintf(`The expected email should be have a valid email format. Received:%s`, res))
	require.LessOrEqual(t, int64(len(res)), shortMaxLength, fmt.Sprintf("The email should be less than or equal to the max length. This is the error email:%s", res))
}

func Test_GenerateRandomEmail(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))

	res, err := generateRandomEmail(randomizer, int64(40), fullNameEmailType, []string{})

	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, true, transformer_utils.IsValidEmail(res), fmt.Sprintf(`The expected email should be have a valid email format. Received:%s`, res))
	require.LessOrEqual(t, int64(len(res)), int64(40), fmt.Sprintf("The email should be less than or equal to the max length. This is the error email:%s", res))
}

func Test_GenerateRandomEmail_Uuid(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))

	res, err := generateRandomEmail(randomizer, int64(40), uuidV4EmailType, []string{})

	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, true, transformer_utils.IsValidEmail(res), fmt.Sprintf(`The expected email should be have a valid email format. Received:%s`, res))
	require.LessOrEqual(t, int64(len(res)), int64(40), fmt.Sprintf("The email should be less than or equal to the max length. This is the error email:%s", res))
}
func Test_GenerateRandomEmail_Uuid_Small(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))

	res, err := generateRandomEmail(randomizer, int64(8), uuidV4EmailType, []string{})

	require.NoError(t, err)
	require.NotEmpty(t, res)
	require.Equal(t, true, transformer_utils.IsValidEmail(res), fmt.Sprintf(`The expected email should be have a valid email format. Received:%s`, res))
	require.LessOrEqual(t, int64(len(res)), int64(40), fmt.Sprintf("The email should be less than or equal to the max length. This is the error email:%s", res))
}

func Test_RandomEmailTransformer(t *testing.T) {
	maxLength := int64(40)
	mapping := fmt.Sprintf(`root = generate_email(max_length:%d)`, maxLength)
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

var seed, _ = transformer_utils.GenerateCryptoSeed()
var randomizer = rand.New(rand.NewSource(seed))

var mapping = fmt.Sprintf(`root = generate_email(max_length:%d)`, maxLength)
var ex, _ = bloblang.Parse(mapping)

func Test_generateRandomLastName(t *testing.T) {

	errgrp, _ := errgroup.WithContext(context.Background())

	i := 0
	for i < 10000 {
		errgrp.Go(func() error {
			_, err := ex.Query(nil)
			return err
		})
	}
	err := errgrp.Wait()
	require.NoError(t, err)
}
