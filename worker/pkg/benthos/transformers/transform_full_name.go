package transformers

import (
	"errors"
	"fmt"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformFullName

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Transforms an existing full name.").
		Category("string").
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Specifies the maximum length for the transformed data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false).Description("Whether the original length of the input data should be preserved during transformation. If set to true, the transformation logic will ensure that the output data has the same length as the input data.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used for generating deterministic transformations."))

	err := bloblang.RegisterFunctionV2(
		"transform_full_name",
		spec,
		func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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

			seed, err := transformer_utils.GetSeedOrDefault(seedArg)
			if err != nil {
				return nil, err
			}

			randomizer := rng.New(seed)

			return func() (any, error) {
				res, err := transformFullName(randomizer, value, preserveLength, maxLength)
				if err != nil {
					return nil, fmt.Errorf("unable to run transform_full_name: %w", err)
				}
				return res, nil
			}, nil
		},
	)

	if err != nil {
		panic(err)
	}
}

func NewTransformFullNameOptsFromConfig(
	config *mgmtv1alpha1.TransformFullName,
	maxLength *int64,
) (*TransformFullNameOpts, error) {
	if config == nil {
		return NewTransformFullNameOpts(nil, nil, nil)
	}
	return NewTransformFullNameOpts(
		maxLength,
		config.PreserveLength,
		nil,
	)
}

func (t *TransformFullName) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformFullNameOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	return transformFullName(
		parsedOpts.randomizer,
		valueStr,
		parsedOpts.preserveLength,
		parsedOpts.maxLength,
	)
}

func transformFullName(
	randomizer rng.Rand,
	name string,
	preserveLength bool,
	maxLength int64,
) (*string, error) {
	if name == "" {
		return nil, nil
	}

	maxValue := maxLength

	if preserveLength {
		if strings.TrimSpace(name) == "" {
			// If the input is all spaces, return a string of spaces of the same length
			output := strings.Repeat(" ", len(name))
			return &output, nil
		}

		firstname, lastname := splitEvenly(name)
		if lastname == "" {
			// Single name: generate two names, join with space, pad/truncate
			minFirst := int64(1)
			minLast := int64(1)
			newfirstname, _ := generateRandomFirstName(randomizer, &minFirst, minFirst)
			newlastname, _ := generateRandomLastName(randomizer, &minLast, minLast)
			fullname := fmt.Sprintf("%s %s", newfirstname, newlastname)
			if len(fullname) < len(name) {
				fullname += transformer_utils.GetRandomCharacterString(randomizer, int64(len(name)-len(fullname)))
			} else if len(fullname) > len(name) {
				fullname = fullname[:len(name)]
			}
			return &fullname, nil
		}

		minFirst := int64(len(firstname))
		newfirstname, _ := generateRandomFirstName(randomizer, &minFirst, minFirst)
		if newfirstname == "" {
			newfirstname, _ = generateRandomFirstName(randomizer, nil, minFirst)
			if int64(len(newfirstname)) != minFirst {
				newfirstname += transformer_utils.GetRandomCharacterString(
					randomizer,
					minFirst-int64(len(newfirstname)),
				)
			}
		}
		// Calculate remaining length for last name (accounting for space)
		remainingLength := int64(len(name)) - int64(len(newfirstname)) - 1 // -1 for space
		if remainingLength < 0 {
			remainingLength = 0
		}
		minLast := int64(len(lastname))
		if minLast == 0 {
			minLast = remainingLength
		}
		newlastname, _ := generateRandomLastName(randomizer, &minLast, minLast)
		if newlastname == "" {
			newlastname, _ = generateRandomLastName(randomizer, nil, minLast)
			if int64(len(newlastname)) != minLast {
				newlastname += transformer_utils.GetRandomCharacterString(
					randomizer,
					minLast-int64(len(newlastname)),
				)
			}
		}
		if newfirstname != "" && newlastname != "" {
			fullname := fmt.Sprintf("%s %s", newfirstname, newlastname)
			if int64(len(fullname)) != int64(len(name)) {
				if int64(len(fullname)) < int64(len(name)) {
					fullname += transformer_utils.GetRandomCharacterString(
						randomizer,
						int64(len(name))-int64(len(fullname)),
					)
				} else {
					fullname = fullname[:len(name)]
				}
			}
			return &fullname, nil
		}
	}

	output, err := generateRandomFullName(randomizer, maxValue)
	if err != nil {
		return nil, err
	}
	if preserveLength && len(output) != int(maxLength) {
		output += transformer_utils.GetRandomCharacterString(
			randomizer,
			maxLength-int64(len(output)),
		)
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
