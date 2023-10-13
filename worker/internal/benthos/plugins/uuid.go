package neosync_plugins

import (
	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	"github.com/bxcodec/faker/v4"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("include_hyphen"))

	// register the plugin
	err := bloblang.RegisterMethodV2("uuid", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		include_hyphen, err := args.GetBool("include_hyphen")
		if err != nil {
			return nil, err
		}

		return bloblang.BoolMethod(func(b bool) (any, error) {
			res, err := ProcessUuid(include_hyphen)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// main plugin logic goes here
func ProcessUuid(include_hyphen bool) (string, error) {

	var returnValue string

	if include_hyphen {

		//generate uuid with hyphens
		returnValue = faker.UUIDHyphenated()

	} else {

		//generates uuid with no hyphens
		returnValue = faker.UUIDDigit()
	}

	return returnValue, nil
}
