package transformers

import (
	"fmt"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateCategorical

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Randomly selects a value from a defined set of categorical values.").
		Param(bloblang.NewStringParam("categories").Default("ultimo,proximo,semper").Description("A list of comma-separated string values to randomly select from.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used to generate deterministic outputs."))

	err := bloblang.RegisterFunctionV2("generate_categorical", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		// get stringified categories
		catString, err := args.GetString("categories")
		if err != nil {
			return nil, err
		}
		categories := strings.Split(catString, ",")

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
			res := generateCategorical(randomizer, categories)
			return res, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func NewGenerateCategoricalOptsFromConfig(config *mgmtv1alpha1.GenerateCategorical) (*GenerateCategoricalOpts, error) {
	if config == nil {
		return NewGenerateCategoricalOpts(nil, nil)
	}
	return NewGenerateCategoricalOpts(config.Categories, nil)
}

func (t *GenerateCategorical) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateCategoricalOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return generateCategorical(parsedOpts.randomizer, strings.Split(parsedOpts.categories, ",")), nil
}

// Generates a randomly selected value from the user-provided list of categories. We don't account for the maxLength param here because the input is user-provided. We assume that they values they provide in the set abide by the maxCharacterLength constraint.
func generateCategorical(randomizer rng.Rand, categories []string) string {
	if len(categories) == 0 {
		return ""
	}
	randomIndex := randomizer.Intn(len(categories))
	return categories[randomIndex]
}
