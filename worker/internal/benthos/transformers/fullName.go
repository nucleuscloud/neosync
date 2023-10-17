package neosync_transformers

import (
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	"github.com/bxcodec/faker/v4"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("preserve_length"))

	// register the plugin
	err := bloblang.RegisterMethodV2("fullnametransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}
		return bloblang.StringMethod(func(s string) (any, error) {
			res, err := ProcessFullName(s, preserveLength)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// main plugin logic goes here
func ProcessFullName(fn string, preserveLength bool) (string, error) {

	var returnValue string

	parsedName := strings.Split(fn, " ")

	if preserveLength {

		fn, err := ProcessFirstName(parsedName[0], preserveLength)
		if err != nil {
			return "", err
		}

		ln, err := ProcessLastName(parsedName[1], preserveLength)
		if err != nil {
			return "", err
		}

		returnValue = fn + " " + ln

		return returnValue, err

	} else {

		// generate random first name
		returnValue = faker.Name()
	}

	return returnValue, nil
}
