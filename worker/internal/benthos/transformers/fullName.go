package neosync_transformers

import (
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec().Param(bloblang.NewStringParam(("name")).Optional()).Param(bloblang.NewBoolParam("preserve_length").Optional())

	// register the plugin
	err := bloblang.RegisterFunctionV2("fullnametransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

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
			res, err := GenerateFullName(name, preserveLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// generates a random full name
func GenerateFullName(name string, pl bool) (string, error) {

	if name != "" {
		if !pl {
			res, err := GenerateFullNameWithRandomLength()
			return res, err
		} else {
			res, err := GenerateFullNameWithLength(name)
			return res, err
		}
	} else {
		res, err := GenerateFullNameWithRandomLength()
		return res, err
	}

}

// main transformer logic goes here
func GenerateFullNameWithRandomLength() (string, error) {

	fn, err := GenerateFirstNameWithRandomLength()
	if err != nil {
		return "", err
	}

	ln, err := GenerateLastNameWithRandomLength()
	if err != nil {
		return "", err
	}

	returnValue := fn + " " + ln

	return returnValue, err

}

func GenerateFullNameWithLength(fn string) (string, error) {

	parsedName := strings.Split(fn, " ")

	fn, err := GenerateFirstName(parsedName[0], true)
	if err != nil {
		return "", err
	}

	ln, err := GenerateLastName(parsedName[1], true)
	if err != nil {
		return "", err
	}

	returnValue := fn + " " + ln

	return returnValue, err

}
