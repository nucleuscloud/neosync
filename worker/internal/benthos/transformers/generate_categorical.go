package transformers

import (
	"math/rand"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec().Param(bloblang.NewStringParam("categories"))

	err := bloblang.RegisterFunctionV2("generate_categorical", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		// get stringified categories
		catString, err := args.GetString("categories")
		if err != nil {
			return nil, err
		}

		// convert stringified categories to array
		catArray := strings.Split(catString, ",")

		return func() (any, error) {
			res := GenerateCategorical(catArray)
			return res, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

// Generates a randomly selected value from the user-provided list of categories. We don't account for the maxLength param here because the input is user-provided. We assume that they values they provide in the set abide by the maxCharacterLength constraint.
func GenerateCategorical(categories []string) string {

	//nolint:all
	randomIndex := rand.Intn(len(categories))

	return categories[randomIndex]
}
