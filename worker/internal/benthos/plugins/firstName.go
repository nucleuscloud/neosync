package neosync_plugins

import (
	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	"github.com/bxcodec/faker/v4"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("preserve_length"))

	// register the plugin
	err := bloblang.RegisterMethodV2("firstnametransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}
		return bloblang.StringMethod(func(s string) (any, error) {
			res, err := ProcessFirstName(s, preserveLength)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// main plugin logic goes here
func ProcessFirstName(fn string, preserveLength bool) (string, error) {

	var returnValue string

	if preserveLength {

		for {
			returnValue = faker.LastName()
			if len(returnValue) >= len(fn) {
				return returnValue[:len(fn)], nil

			}
		}

	} else {

		// generate random first name
		returnValue = faker.FirstName()
	}

	return returnValue, nil
}
