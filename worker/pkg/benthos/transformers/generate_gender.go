package transformers

import (
	"errors"
	"time"

	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateGender

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Randomly generates one of the following genders: female, male, undefined, nonbinary.").
		Param(bloblang.NewBoolParam("abbreviate").Default(false).Description("Shortens length of generated value to 1.")).
		Param(bloblang.NewInt64Param("max_length").Default(10000).Description("Specifies the maximum length for the generated data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewInt64Param("seed").Default(time.Now().UnixNano()).Description("An optional seed value used to generate deterministic outputs."))

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

func (t *GenerateGender) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateGenderOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	return generateRandomGender(parsedOpts.randomizer, parsedOpts.abbreviate, parsedOpts.maxLength), nil
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
