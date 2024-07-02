package transformers

import (
	"errors"
	"fmt"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
)

// +javascriptFncBuilder:transform:TransformFirstName

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("max_length").Default(10000)).
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false)).
		Param(bloblang.NewInt64Param("seed").Optional())

	err := bloblang.RegisterFunctionV2("transform_first_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		valuePtr, err := args.GetOptionalString("value")
		if err != nil {
			return nil, err
		}

		var value string
		if valuePtr != nil {
			value = *valuePtr
		}

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		seedArg, err := args.GetOptionalInt64("seed")
		if err != nil {
			return nil, err
		}
		var seed int64
		if seedArg != nil {
			seed = *seedArg
		} else {
			// we want a bit more randomness here with generate_email so using something that isn't time based
			var err error
			seed, err = transformer_utils.GenerateCryptoSeed()
			if err != nil {
				return nil, err
			}
		}

		randomizer := rng.New(seed)

		return func() (any, error) {
			res, err := transformFirstName(randomizer, value, preserveLength, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_first_name: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

type TransformFirstNameOpts struct {
	randomizer     rng.Rand
	maxLength      int64
	preserveLength bool
}
type TransformFirstName struct {
}

func NewTransformFirstName() *TransformFirstName {
	return &TransformFirstName{}
}

func (t *TransformFirstName) GetJsTemplateData() (*TemplateData, error) {
	return &TemplateData{
		Name:        "TransformFirstName",
		Description: "Takes value and transforms it to a first name",
		Params: []*Param{
			{Name: "maxLength", TypeStr: "int64", What: "Max character length of first name"},
			{Name: "preserveLength", TypeStr: "bool", What: "Whether to keep name length the same as original value"},
			{Name: "seed", TypeStr: "int64", What: "Randomzer seed"},
			{Name: "value", TypeStr: "string", What: "'value to transform"},
		},
	}, nil
}
func (t *TransformFirstName) ParseOptions(opts map[string]any) (any, error) {
	var seed int64
	seedArg, ok := opts["seed"].(int64)
	if ok {
		seed = seedArg
	} else {
		var err error
		seed, err = transformer_utils.GenerateCryptoSeed()
		if err != nil {
			return nil, err
		}
	}

	preserveLength, ok := opts["preserveLength"].(bool)
	if !ok {
		preserveLength = false
	}

	maxLength, ok := opts["maxLength"].(int64)
	if !ok {
		maxLength = 10000
	}

	return &TransformFirstNameOpts{
		randomizer:     rng.New(seed),
		preserveLength: preserveLength,
		maxLength:      maxLength,
	}, nil
}

func (t *TransformFirstName) Transform(value any, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformFirstNameOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	return transformFirstName(parsedOpts.randomizer, valueStr, parsedOpts.preserveLength, parsedOpts.maxLength)
}

// Generates a random first name which can be of either random length or as long as the input name
func transformFirstName(randomizer rng.Rand, value string, preserveLength bool, maxLength int64) (*string, error) {
	if value == "" {
		return nil, nil
	}

	maxValue := maxLength

	// unable to generate a random name of this fixed size
	// we may want to change this to just use the below algorithm and pad so that it is more unique
	// as with this algorithm, it will only ever use values from the underlying map that are that specific size
	if preserveLength {
		maxValue = int64(len(value))
		output, err := GenerateRandomFirstName(randomizer, &maxValue, maxValue)
		if err == nil {
			return &output, nil
		}
	}

	output, err := GenerateRandomFirstName(randomizer, nil, maxValue)
	if err != nil {
		return nil, err
	}

	// pad the string so that we can get the correct value
	if preserveLength && int64(len(output)) != maxValue {
		output += transformer_utils.GetRandomCharacterString(randomizer, maxValue-int64(len(output)))
	}
	return &output, nil
}
