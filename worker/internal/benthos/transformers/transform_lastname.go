package transformers

import (
	_ "embed"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformers_dataset "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/data-sets"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

var lastNames = transformers_dataset.LastNames

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("max_length")).
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length"))

	err := bloblang.RegisterFunctionV2("transform_last_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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

		return func() (any, error) {
			res, err := TransformLastName(value, preserveLength, maxLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

// Generates a random last name which can be of either random length between [2,12] characters or as long as the input name
func TransformLastName(name string, preserveLength bool, maxLength int64) (*string, error) {
	if name == "" {
		return nil, nil
	}

	nameLength := int64(len(name))

	if preserveLength {
		// assume that if pl is true than it already meets the maxCharacterLimit constraint
		res, err := GenerateRandomLastNameInLengthRange(nameLength, nameLength)
		if err != nil {
			return nil, err
		}
		return &res, nil
	} else {
		res, err := GenerateRandomLastNameInLengthRange(minNameLength, maxLength)
		if err != nil {
			return nil, err
		}
		return &res, nil
	}
}

// Generates a random last name with length [min, max]. If the length is greater than 12, a last name of length 12 will be returned.
func GenerateRandomLastNameInLengthRange(minLength, maxLength int64) (string, error) {
	if minLength == maxLength {
		if minLength > 12 {
			names := lastNames[12]
			res, err := transformer_utils.GetRandomValueFromSlice[string](names)
			if err != nil {
				return "", err
			}
			return res, nil
		} else if minLength < minNameLength {
			names := lastNames[2]
			res, err := transformer_utils.GetRandomValueFromSlice[string](names)
			if err != nil {
				return "", err
			}
			return res, nil
		} else {
			names := lastNames[minLength]
			res, err := transformer_utils.GetRandomValueFromSlice[string](names)
			if err != nil {
				return "", err
			}
			return res, nil
		}
	} else {
		if maxLength < 12 && maxLength >= 2 {
			names := lastNames[maxLength]
			res, err := transformer_utils.GetRandomValueFromSlice[string](names)
			if err != nil {
				return "", err
			}
			return res, nil
		} else if maxLength < 2 {
			res, err := transformer_utils.GenerateRandomStringWithDefinedLength(1)
			if err != nil {
				return "", err
			}
			return res, nil
		} else {
			randInd, err := transformer_utils.GenerateRandomInt64InValueRange(2, 12)
			if err != nil {
				return "", err
			}

			names := lastNames[randInd]
			res, err := transformer_utils.GetRandomValueFromSlice[string](names)
			if err != nil {
				return "", err
			}
			return res, nil
		}
	}
}
