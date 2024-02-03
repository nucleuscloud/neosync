package transformers

import (
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("min")).
		Param(bloblang.NewInt64Param("max")).
		Param(bloblang.NewInt64Param("max_length"))

	err := bloblang.RegisterFunctionV2("generate_random_string", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		min, err := args.GetInt64("min")
		if err != nil {
			return nil, err
		}

		max, err := args.GetInt64("max")
		if err != nil {
			return nil, err
		}

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateRandomString(min, max, maxLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// generate a random string with length between min and max
func GenerateRandomString(min, max, maxLength int64) (string, error) {

	if max > maxLength && min > maxLength {
		val, err := transformer_utils.GenerateRandomStringWithDefinedLength(maxLength)

		if err != nil {
			return "", fmt.Errorf("unable to generate a random string with length")
		}

		return val, nil
	} else if max > maxLength && min < maxLength {

		val, err := transformer_utils.GenerateRandomStringWithInclusiveBounds(min, maxLength)

		if err != nil {
			return "", fmt.Errorf("unable to generate a random string with length")
		}

		return val, nil

	} else {
		val, err := transformer_utils.GenerateRandomStringWithInclusiveBounds(min, max)

		if err != nil {
			return "", fmt.Errorf("unable to generate a random string with length")
		}
		return val, nil
	}

}
