package neosync_transformers

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
	nameBytes []byte
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

// main transformer logic goes here
func GenerateFirstName(fn string, pl bool) (string, error) {

	if !pl {
		res, err := GenerateFirstNameWithRandomLength(fn)
		return res, err
	} else {
		res, err := GenerateFirstNameWithLength(fn, len(fn))
		return res, err
	}
}

func GenerateFirstNameWithRandomLength(fn string) (string, error) {

	var returnValue string

	data := struct {
		Names []NameGroup `json:"names"`
	}{}
	if err := json.Unmarshal(nameBytes, &data); err != nil {
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

	return returnValue, nil
}

// main transformer logic goes here
func GenerateFirstNameWithLength(fn string, l int) (string, error) {

	var returnValue string

	data := struct {
		Names []NameGroup `json:"names"`
	}{}
	if err := json.Unmarshal(nameBytes, &data); err != nil {
		panic(err)
	}

	names := data.Names

	for _, v := range names {
		if v.NameLength == l {
			res, err := transformer_utils.GetRandomValueFromSlice[string](v.Names)
			if err != nil {
				return "", err
			}
			returnValue = res
		}
	}

	return returnValue, nil
}
