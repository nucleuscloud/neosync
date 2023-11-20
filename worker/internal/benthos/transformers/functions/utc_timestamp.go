package neosync_benthos_transformers_functions

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec()

	// register the function
	err := bloblang.RegisterFunctionV2("utctimestamptransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {

			val, err := GenerateRandomUTCTimestamp()

			if err != nil {
				return false, fmt.Errorf("unable to generate random utc timestamp")
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
