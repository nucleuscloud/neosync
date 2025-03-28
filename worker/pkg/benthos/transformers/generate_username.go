package transformers

import (
	"fmt"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateUsername

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Randomly generates a username").
		Category("string").
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length for the generated data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2(
		"generate_username",
		spec,
		func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
				res, err := generateUsername(randomizer, maxLength)
				if err != nil {
					return nil, fmt.Errorf("unable to run generate_username: %w", err)
				}
				return res, nil
			}, nil
		},
	)

	if err != nil {
		panic(err)
	}
}

func NewGenerateUsernameOptsFromConfig(
	config *mgmtv1alpha1.GenerateUsername,
	maxLength *int64,
) (*GenerateUsernameOpts, error) {
	if config == nil {
		return NewGenerateUsernameOpts(nil, nil)
	}
	return NewGenerateUsernameOpts(
		maxLength, nil,
	)
}

func (t *GenerateUsername) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateUsernameOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return generateUsername(parsedOpts.randomizer, parsedOpts.maxLength)
}

// Generates a username with a lowercase first initial and titlecase lastname
func generateUsername(randomizer rng.Rand, maxLength int64) (string, error) {
	fn := transformer_utils.GetRandomCharacterString(randomizer, 1)

	ln, err := generateRandomLastName(randomizer, nil, maxLength-1)
	if err != nil {
		return "", err
	}
	return fn + strings.ToLower(ln), nil
}
