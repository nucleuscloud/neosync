package transformers

import (
	"errors"
	"fmt"
	"time"

	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateUnixTimestamp

func init() {
	spec := bloblang.NewPluginSpec().Description("Randomly generates a Unix timestamp.").
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_unixtimestamp", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		seedArg, err := args.GetOptionalInt64("seed")
		if err != nil {
			return nil, err
		}

		seed, err := transformer_utils.GetSeedOrDefault(seedArg)
		if err != nil {
			return nil, err
		}
		randomizer := rng.New(seed)

		return func() (any, error) {
			val, err := generateRandomUnixTimestamp(randomizer)
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
	parsedOpts, ok := opts.(GenerateUnixTimestampOpts)
	if !ok {
		return nil, errors.New("invalid parsed opts")
	}
	return generateRandomUnixTimestamp(parsedOpts.randomizer)
}

const (
	secondsInYear = int64(365 * 24 * 60 * 60) // Max seconds in a year
)

func generateRandomUnixTimestamp(randomizer rng.Rand) (int64, error) {
	// get the current UTC time
	currentTime := time.Now().Unix()
	randomSeconds := randomizer.Int63n(secondsInYear + 1)
	// subtract the random number of seconds from the current time
	randomUnixTimestamp := currentTime - randomSeconds
	return randomUnixTimestamp, nil
}
