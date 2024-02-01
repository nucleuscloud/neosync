package transformers

import (
	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length"))

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
		return func() (any, error) {
			res, err := GenerateFullName(value, preserveLength)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

func GenerateFullName(name string, pl bool) (*string, error) {

	if name == "" {
		return nil, nil
	}

	if pl {
		res, err := GenerateFullNameWithLength(name)
		if err != nil {
			return nil, err
		}
		return &res, nil

	} else {
		res, err := GenerateFullNameWithRandomLength()
		if err != nil {
			return nil, err
		}
		return &res, nil
	}

}

func GenerateFullNameWithRandomLength() (string, error) {

	// fn, err := GenerateRandomFirstNameWithLength(int64(4), int64(4))
	// if err != nil {
	// 	return "", err
	// }

	// ln, err := GenerateRandomLastName()
	// if err != nil {
	// 	return "", err
	// }

	// returnValue := fn + " " + ln

	// return returnValue, err

	return "fasfadf", nil

}

func GenerateFullNameWithLength(fn string) (string, error) {

	// parsedName := strings.Split(fn, " ")

	// fn, err := GenerateRandomFirstNameWithLength(parsedName[0])
	// if err != nil {
	// 	return "", err
	// }

	// ln, err := GenerateRandomLastNameWithLength(parsedName[1])
	// if err != nil {
	// 	return "", err
	// }

	// returnValue := fn + " " + ln

	// return returnValue, err

	return "fasfadf", nil

}
