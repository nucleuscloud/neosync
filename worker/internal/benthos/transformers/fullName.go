package neosync_transformers

import (
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("preserve_length"))

	// register the plugin
	err := bloblang.RegisterMethodV2("fullnametransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}
		return bloblang.StringMethod(func(s string) (any, error) {
			res, err := GenerateFullName(s, preserveLength)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// main transformer logic goes here
func GenerateFullName(fn string, pl bool) (string, error) {

	if !pl {
		res, err := GenerateFullNameWithRandomLength()
		return res, err
	} else {
		res, err := GenerateFullNameWithLength(fn)
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
