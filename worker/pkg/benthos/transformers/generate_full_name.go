package transformers

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateFullName

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Generates a new full name consisting of a first and last name.").
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length for the generated data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_full_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			res, err := generateRandomFullName(randomizer, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_full_name: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func NewGenerateFullNameOptsFromConfig(config *mgmtv1alpha1.GenerateFullName, maxLength *int64) (*GenerateFullNameOpts, error) {
	if config == nil {
		return NewGenerateFullNameOpts(
			nil,
			nil,
		)
	}
	return NewGenerateFullNameOpts(
		maxLength, nil,
	)
}

func (t *GenerateFullName) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateFullNameOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return generateRandomFullName(parsedOpts.randomizer, parsedOpts.maxLength)
}

/* Generates a random full name */
func generateRandomFullName(randomizer rng.Rand, maxLength int64) (string, error) {
	maxLengthMinusSpace := maxLength - 1
	if maxLengthMinusSpace <= 0 {
		return "", fmt.Errorf("unable to generate full name including space with provided max length: %d", maxLength)
	}
	maxFirstNameIdx, maxLastNameIdx := transformer_utils.FindClosestPair(
		transformers_dataset.FirstNameIndices, transformers_dataset.LastNameIndices,
		maxLengthMinusSpace,
	)
	if maxFirstNameIdx == -1 || maxLastNameIdx == -1 {
		return "", fmt.Errorf("unable to generate a full name with the provided max length: %d", maxLength)
	}

	maxFirstNameLength := transformers_dataset.FirstNameIndices[maxFirstNameIdx]
	maxLastNameLength := transformers_dataset.LastNameIndices[maxLastNameIdx]
	firstname, err := generateRandomFirstName(randomizer, nil, maxFirstNameLength)
	if err != nil {
		return "", fmt.Errorf("unable to generate random first name with length: %d", maxFirstNameLength)
	}
	lastname, err := generateRandomLastName(randomizer, nil, maxLastNameLength)
	if err != nil {
		return "", fmt.Errorf("unable to generate random last name with length: %d", maxLastNameLength)
	}

	return fmt.Sprintf("%s %s", firstname, lastname), nil
}
