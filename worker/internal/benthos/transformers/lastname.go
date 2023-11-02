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

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("preserve_length"))

	// register the plugin
	err := bloblang.RegisterMethodV2("lastnametransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}
		return bloblang.StringMethod(func(s string) (any, error) {
			res, err := GenerateLastName(s, preserveLength)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// Generates a random last name
func GenerateLastName(name string, preserveLength bool) (string, error) {

	if !preserveLength {
		res, err := GenerateLastNameWithRandomLength()
		return res, err
	} else {
		res, err := GenerateLastNameWithLength(name)
		return res, err
	}
}

func GenerateLastNameWithRandomLength() (string, error) {

	var returnValue string

	data := struct {
		Names []NameGroup `json:"names"`
	}{}
	if err := json.Unmarshal(lastNameBytes, &data); err != nil {
		panic(err)
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
			} else {

			}
			returnValue = res
		}
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
		panic(err)
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
			res, err := GenerateRandomStringWithLength(int64(len(fn)))
			if err != nil {
				return "", err
			}
			returnValue = res
		}
	}

	return returnValue, nil
}
