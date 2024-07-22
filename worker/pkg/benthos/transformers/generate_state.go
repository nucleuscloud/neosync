package transformers

import (
	"errors"
	"math/rand"

	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateState

func init() {
	spec := bloblang.NewPluginSpec().Param(bloblang.NewBoolParam("state_code").Optional()).Description("Randomly selects a US state and either returns the two character state code or the full state name.")
	err := bloblang.RegisterFunctionV2("generate_state", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		codePtr, err := args.GetOptionalBool("state_code")
		if err != nil {
			return nil, err
		}

		defaultCode := false
		if codePtr != nil {
			defaultCode = *codePtr
		}

		return func() (any, error) {
			return generateRandomState(defaultCode), nil
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
