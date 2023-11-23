package transformers

import (
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec().Param(bloblang.NewStringParam(("name"))).Param(bloblang.NewBoolParam("preserve_length"))

	err := bloblang.RegisterFunctionV2("transform_full_name", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		name, err := args.GetString("name")
		if err != nil {
			return nil, err
		}

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
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

func GenerateFullName(name string, pl bool) (string, error) {

	if name != "" {
		if pl {
			res, err := GenerateFullNameWithLength(name)
			return res, err

		} else {
			res, err := GenerateFullNameWithRandomLength()
			return res, err
		}
	} else {
		res, err := GenerateFullNameWithRandomLength()
		return res, err
	}

}

func GenerateFullNameWithRandomLength() (string, error) {

	fn, err := GenerateRandomFirstName()
	if err != nil {
		return "", err
	}

	ln, err := GenerateRandomLastName()
	if err != nil {
		return "", err
	}

	returnValue := fn + " " + ln

	return returnValue, err

}

func GenerateFullNameWithLength(fn string) (string, error) {

	parsedName := strings.Split(fn, " ")

	fn, err := GenerateRandomFirstNameWithLength(parsedName[0])
	if err != nil {
		return "", err
	}

	ln, err := GenerateRandomLastNameWithLength(parsedName[1])
	if err != nil {
		return "", err
	}

	returnValue := fn + " " + ln

	return returnValue, err

}
