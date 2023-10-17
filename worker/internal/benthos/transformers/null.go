package neosync_transformers

import (
	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec()

	// register the function

	err := bloblang.RegisterFunctionV2("transformernull", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (interface{}, error) {
			return "null", nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}
