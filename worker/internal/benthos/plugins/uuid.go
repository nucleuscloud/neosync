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
	err := bloblang.RegisterMethodV2("uuidtransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		include_hyphen, err := args.GetBool("include_hyphen")
		if err != nil {
			return nil, err
		}

		/*we set this to a string method because even though we want to
		ignore the input uuid string (because we're changing it)
		benthos still makes us handle it or it throws an error that is expecting a string input
		so we just ignore it and don't pass it into our ProcessUuid function*/
		return bloblang.StringMethod(func(b string) (any, error) {
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
		//for postgres, if the dest column is defined as a UUID column then it will automatically
		//convert the UUID with no hyphens to having hyphens
		//so this is more useful for string columns or other dbs that won't do the automatic
		//convesion if you want don't want your UUIDs to have hyphens on purpose
		returnValue = faker.UUIDDigit()
	}

	return returnValue, nil
}
