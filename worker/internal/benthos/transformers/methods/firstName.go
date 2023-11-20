package neosync_benthos_transformers_methods

import (
	_ "embed"
	"encoding/json"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

type Names struct {
	Names []NameGroup `json:"names"`
}

type NameGroup struct {
	NameLength int      `json:"name_length"`
	Names      []string `json:"names"`
}

var (
	//go:embed data-sets/first_names.json
	firstNameBytes []byte
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("preserve_length"))

	// register the plugin
	err := bloblang.RegisterMethodV2("firstnametransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}
		return bloblang.StringMethod(func(s string) (any, error) {
			res, err := GenerateFirstName(s, preserveLength)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// Generates a random first name
func GenerateFirstName(name string, preserveLength bool) (string, error) {

	if !preserveLength {
		res, err := GenerateFirstNameWithRandomLength()
		return res, err
	} else {
		res, err := GenerateFirstNameWithLength(name)
		return res, err
	}
}

func GenerateFirstNameWithRandomLength() (string, error) {

	var returnValue string

	data := struct {
		Names []NameGroup `json:"names"`
	}{}
	if err := json.Unmarshal(firstNameBytes, &data); err != nil {
		panic(err)
	}

	names := data.Names

	// get a random length from the first_names.json file
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
func GenerateFirstNameWithLength(fn string) (string, error) {

	var returnValue string

	data := struct {
		Names []NameGroup `json:"names"`
	}{}
	if err := json.Unmarshal(firstNameBytes, &data); err != nil {
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
			res, err := transformer_utils.GenerateRandomStringWithLength(int64(len(fn)))
			if err != nil {
				return "", err
			}
			returnValue = res
		}
	}

	return returnValue, nil
}
