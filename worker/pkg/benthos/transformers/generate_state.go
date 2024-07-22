package transformers

import (
	"errors"
	"math/rand"

	transformers_dataset "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/data-sets"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateState

func init() {
	spec := bloblang.NewPluginSpec().Param(bloblang.NewBoolParam("generate_full_name").Default(false).Description("Randomly selects a US state and either returns the two character state code or the full state name."))
	err := bloblang.RegisterFunctionV2("generate_state", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		generateFullName, err := args.GetBool("state_code")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			return generateRandomState(generateFullName), nil
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
	return generateRandomState(parsedOpts.generateFullName), nil
}

/*
Generates a randomly selected state that exists in the United States.

By default, it returns the 2-letter state code i.e. California will return CA. However, this is configurable using the Generate Full Name parameter which, when set to true, will return the full name of the state starting with a capitalized letter.
*/
func generateRandomState(generateFullName bool) string {
	stateData := transformers_dataset.States

	//nolint:all
	randomIndex := rand.Intn(len(stateData))
	state := stateData[randomIndex].Code

	if generateFullName {
		state = stateData[randomIndex].FullName
	}

	return state
}
