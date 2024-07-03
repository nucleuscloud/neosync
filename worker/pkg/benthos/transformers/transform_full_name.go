package transformers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
)

// +neosyncTransformerBuilder:transform:transformFullName

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("max_length").Default(10000)).
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false)).
		Param(bloblang.NewInt64Param("seed").Optional())

	err := bloblang.RegisterFunctionV2("transform_full_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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
			res, err := transformFullName(randomizer, value, preserveLength, maxLength)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_full_name: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func (t *TransformFullName) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformFullNameOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	return transformFullName(parsedOpts.randomizer, valueStr, parsedOpts.preserveLength, parsedOpts.maxLength)
}

func transformFullName(randomizer rng.Rand, name string, preserveLength bool, maxLength int64) (*string, error) {
	if name == "" {
		return nil, nil
	}

	maxValue := maxLength

	if preserveLength {
		firstname, lastname := splitEvenly(name)
		minFirst := int64(len(firstname))
		newfirstname, _ := generateRandomFirstName(randomizer, &minFirst, minFirst)
		if newfirstname == "" {
			newfirstname, _ = generateRandomFirstName(randomizer, nil, minFirst)
			if int64(len(newfirstname)) != minFirst {
				newfirstname += transformer_utils.GetRandomCharacterString(randomizer, minFirst-int64(len(newfirstname)))
			}
		}
		minLast := int64(len(lastname))
		newlastname, _ := generateRandomLastName(randomizer, &minLast, minLast)
		if newlastname == "" {
			newfirstname, _ = generateRandomLastName(randomizer, nil, minLast)
			if int64(len(newlastname)) != minLast {
				newlastname += transformer_utils.GetRandomCharacterString(randomizer, minFirst-int64(len(newlastname)))
			}
		}
		if newfirstname != "" && newlastname != "" {
			fullname := fmt.Sprintf("%s %s", newfirstname, newlastname)
			return &fullname, nil
		}
	}

	output, err := generateRandomFullName(randomizer, maxValue)
	if err != nil {
		return nil, err
	}
	if preserveLength && len(output) != int(maxLength) {
		output += transformer_utils.GetRandomCharacterString(randomizer, maxLength-int64(len(output)))
	}
	return &output, nil
}

func splitEvenly(input string) (first, last string) {
	parts := strings.Fields(input)

	// Calculate the split index. If there are an odd number of parts, the first half will have one more.
	splitIndex := len(parts) / 2
	if len(parts)%2 != 0 { // Adjust split index if number of words is odd.
		splitIndex++
	}

	firstHalf := strings.Join(parts[:splitIndex], " ")
	secondHalf := strings.Join(parts[splitIndex:], " ")
	return firstHalf, secondHalf
}
