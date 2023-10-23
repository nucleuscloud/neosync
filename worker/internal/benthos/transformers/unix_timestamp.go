package neosync_transformers

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
	err := bloblang.RegisterFunctionV2("unixtimestamptransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {

			val, err := GenerateRandomUnixTimestamp()

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

func GenerateRandomUnixTimestamp() (int64, error) {
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
