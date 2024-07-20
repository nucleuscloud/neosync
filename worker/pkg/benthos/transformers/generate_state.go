package transformers

import (
	"errors"
	"math/rand"

	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateState

func init() {
	spec := bloblang.NewPluginSpec().Param(bloblang.NewBoolParam("state_code"))
	err := bloblang.RegisterFunctionV2("generate_state", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		code, err := args.GetBool("state_code")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			return generateRandomState(code), nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func (t *GenerateState) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateStateOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}
	return generateRandomState(parsedOpts.stateCode), nil
}

// Generates a randomly selected state that exists in the United States
func generateRandomState(state_code bool) string {
	stateData := transformers_dataset.States

	//nolint:all
	randomIndex := rand.Intn(len(stateData))
	gender := stateData[randomIndex].FullName

	if state_code {
		gender = stateData[randomIndex].Code
	}

	return gender
}
