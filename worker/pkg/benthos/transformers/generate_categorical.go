package transformers

import (
	"errors"
	"math/rand"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateCategorical

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewStringParam("categories"))

	err := bloblang.RegisterFunctionV2("generate_categorical", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		// get stringified categories
		catString, err := args.GetString("categories")
		if err != nil {
			return nil, err
		}
		categories := strings.Split(catString, ",")

		return func() (any, error) {
			res := generateCategorical(categories)
			return res, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func (t *GenerateCategorical) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateCategoricalOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	return generateCategorical(strings.Split(parsedOpts.categories, ",")), nil
}

// Generates a randomly selected value from the user-provided list of categories. We don't account for the maxLength param here because the input is user-provided. We assume that they values they provide in the set abide by the maxCharacterLength constraint.
func generateCategorical(categories []string) string {
	if len(categories) == 0 {
		return ""
	}
	//nolint:gosec
	randomIndex := rand.Intn(len(categories))
	return categories[randomIndex]
}
