package neosync_benthos_transformers_functions

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("include_hyphen"))

	// register the plugin
	err := bloblang.RegisterFunctionV2("uuidtransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		include_hyphen, err := args.GetBool("include_hyphen")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {

			val, err := GenerateUuid(include_hyphen)

			if err != nil {
				return false, fmt.Errorf("unable to generate random utc timestamp")
			}
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

// main transformer logic goes here
func GenerateUuid(include_hyphen bool) (string, error) {

	if include_hyphen {

		// generate uuid with hyphens
		return uuid.NewString(), nil

	} else {

		/*generates uuid with no hyphens
		for postgres, if the dest column is defined as a UUID column then it will automatically
		convert the UUID with no hyphens to having hyphens
		so this is more useful for string columns or other dbs that won't do the automatic
		conversion if you want don't want your UUIDs to have hyphens on purpose */

		newUUID := uuid.New()
		uuidWithHyphens := newUUID.String()
		return strings.ReplaceAll(uuidWithHyphens, "-", ""), nil

	}
}
