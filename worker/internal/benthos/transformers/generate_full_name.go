package transformers

import (
	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_full_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {
			res, err := GenerateRandomFullName()
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

/* Generates a random full name */
func GenerateRandomFullName() (string, error) {

	fn, err := GenerateRandomFirstName()
	if err != nil {
		return "", err
	}

	ln, err := GenerateRandomLastName()
	if err != nil {
		return "", err
	}

	returnValue := fn + " " + ln

	return returnValue, err

}
