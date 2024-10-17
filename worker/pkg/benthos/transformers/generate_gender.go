package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateGender

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Randomly generates one of the following genders: female (f), male (m), undefined (u), nonbinary (n).").
		Param(bloblang.NewBoolParam("abbreviate").Default(false).Description("Shortens length of generated value to 1.")).
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length for the generated data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_gender", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		shouldAbbreviate, err := args.GetBool("abbreviate")
		if err != nil {
			return nil, err
		}

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

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
			res := generateRandomGender(randomizer, shouldAbbreviate, maxLength)
			return res, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func NewGenerateGenderOptsFromConfig(config *mgmtv1alpha1.GenerateGender, maxLength *int64) (*GenerateGenderOpts, error) {
	if config == nil {
		return NewGenerateGenderOpts(
			nil,
			nil,
			nil,
		)
	}
	return NewGenerateGenderOpts(
		config.Abbreviate,
		maxLength, nil,
	)
}

func (t *GenerateGender) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateGenderOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
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
