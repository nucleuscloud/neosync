package transformers

import (
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("abbreviate").Default(false)).
		Param(bloblang.NewInt64Param("max_length").Default(10000)).
		Param(bloblang.NewInt64Param("seed").Default(time.Now().UnixNano()))

	err := bloblang.RegisterFunctionV2("generate_gender", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		shouldAbbreviate, err := args.GetBool("abbreviate")
		if err != nil {
			return nil, err
		}

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		seed, err := args.GetInt64("seed")
		if err != nil {
			return nil, err
		}
		randomizer := rng.New(seed)

		return func() (any, error) {
			res := generateRandomGender(randomizer, shouldAbbreviate, maxLength)
			return res, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

var genders = []string{"undefined", "nonbinary", "female", "male"}

func generateRandomGender(randomizer rng.Rand, shouldAbbreviate bool, maxLength int64) string {
	genderIdx := randomizer.Intn(len(genders))
	gender := transformer_utils.TrimStringIfExceeds(genders[genderIdx], maxLength)
	if shouldAbbreviate {
		gender = transformer_utils.TrimStringIfExceeds(gender, 1)
	}
	return gender
}
