package neosync_transformers

import (
	_ "embed"
	"encoding/json"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

var (
	//go:embed data-sets/last_names.json
	lastNameBytes []byte
)

func init() {

	spec := bloblang.NewPluginSpec().Param(bloblang.NewStringParam("name").Optional()).Param(bloblang.NewBoolParam("preserve_length").Optional())

	err := bloblang.RegisterFunctionV2("lastnametransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		namePtr, err := args.GetOptionalString("name")
		if err != nil {
			return nil, err
		}
		var name string
		if namePtr != nil {
			name = *namePtr
		}

		preserveLengthPtr, err := args.GetOptionalBool("preserve_length")
		if err != nil {
			return nil, err
		}
		var preserveLength bool
		if preserveLengthPtr != nil {
			preserveLength = *preserveLengthPtr
		}

		return func() (any, error) {
			res, err := GenerateLastName(name, preserveLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

// Generates a random last name
func GenerateLastName(name string, preserveLength bool) (string, error) {

	if name != "" {
		if !preserveLength {
			res, err := GenerateLastNameWithRandomLength()
			return res, err
		} else {
			res, err := GenerateLastNameWithLength(name)
			return res, err
		}
	} else {
		res, err := GenerateLastNameWithRandomLength()
		return res, err
	}

}

func GenerateLastNameWithRandomLength() (string, error) {

	var returnValue string

	data := struct {
		Names []NameGroup `json:"names"`
	}{}
	if err := json.Unmarshal(lastNameBytes, &data); err != nil {
		return "", err
	}

	names := data.Names

	// get a random length from the last_names.json file
	var nameLengths []int

	for _, v := range names {
		nameLengths = append(nameLengths, v.NameLength)
	}

	randomNameLengthVal, err := transformer_utils.GetRandomValueFromSlice[int](nameLengths)
	if err != nil {
		return "", err
	}

	for _, v := range names {
		if v.NameLength == randomNameLengthVal {
			res, err := transformer_utils.GetRandomValueFromSlice[string](v.Names)
			if err != nil {
				return "", err
			}
			returnValue = res
		}
	}

	// handles the case where the name provided is longer than the longest names in the first_names slice
	if returnValue == "" {
		res, err := transformer_utils.GetRandomValueFromSlice[string](names[3].Names)
		if err != nil {
			return "", err
		}

		returnValue = res
	}

	return returnValue, nil
}

// main transformer logic goes here
func GenerateLastNameWithLength(fn string) (string, error) {

	var returnValue string

	data := struct {
		Names []NameGroup `json:"names"`
	}{}
	if err := json.Unmarshal(lastNameBytes, &data); err != nil {
		return "", err
	}

	names := data.Names

	for _, v := range names {
		if v.NameLength == len(fn) {
			res, err := transformer_utils.GetRandomValueFromSlice[string](v.Names)
			if err != nil {
				return "", err
			}
			returnValue = res
		} else {
			res, err := transformer_utils.GenerateRandomStringWithLength(int64(len(fn)))
			if err != nil {
				return "", err
			}
			returnValue = res
		}
	}

	return returnValue, nil
}
