package transformers

import (
	"errors"
	"time"

	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateUTCTimestamp

func init() {
	spec := bloblang.NewPluginSpec().Description("Randomly generates a UTC timestamp.").
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_utctimestamp", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			return generateRandomUTCTimestamp(randomizer), nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func (t *GenerateUTCTimestamp) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(GenerateUTCTimestampOpts)
	if !ok {
		return nil, errors.New("invalid parsed opts")
	}
	return generateRandomUTCTimestamp(parsedOpts.randomizer), nil
}

func generateRandomUTCTimestamp(randomizer rng.Rand) time.Time {
	// get the current UTC time
	currentTime := time.Now().UTC()
	randomSeconds := randomizer.Int63n(secondsInYear + 1)
	// subtract the random number of seconds from the current time
	randomTime := currentTime.Add(-time.Duration(randomSeconds) * time.Second)
	return randomTime
}
