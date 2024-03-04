package transformers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
)

func init() {
	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_utctimestamp", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		return func() (any, error) {
			val, err := GenerateRandomUTCTimestamp()
			if err != nil {
				return false, fmt.Errorf("unable to run generate_utctimestamp: %w", err)
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func GenerateRandomUTCTimestamp() (time.Time, error) {
	// get the current UTC time
	currentTime := time.Now().UTC()

	// generate a random number of seconds
	maxSeconds := big.NewInt(int64(365 * 24 * 60 * 60)) // Max seconds in a year
	randomSeconds, err := rand.Int(rand.Reader, maxSeconds)
	if err != nil {
		return time.Time{}, err
	}

	// subtract the random number of seconds from the current time
	randomTime := currentTime.Add(-time.Duration(randomSeconds.Int64()) * time.Second)
	return randomTime, nil
}
