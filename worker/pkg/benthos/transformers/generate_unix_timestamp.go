package transformers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateUnixTimestamp

func init() {
	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_unixtimestamp", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		return func() (any, error) {
			val, err := generateRandomUnixTimestamp()
			if err != nil {
				return false, fmt.Errorf("unable to run generate_unixtimestamp: %w", err)
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func (t *GenerateUnixTimestamp) Generate(opts any) (any, error) {
	return generateRandomUnixTimestamp()
}

func generateRandomUnixTimestamp() (int64, error) {
	// get the current UTC time
	currentTime := time.Now().Unix()

	// generate a random number of seconds
	maxSeconds := int64(365 * 24 * 60 * 60) // Max seconds in a year
	randomSeconds, err := rand.Int(rand.Reader, big.NewInt(maxSeconds+1))
	if err != nil {
		return 0, err
	}

	// subtract the random number of seconds from the current time
	randomUnixTimestamp := currentTime - randomSeconds.Int64()

	return randomUnixTimestamp, nil
}
