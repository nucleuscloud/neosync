package neosync_transformers

import (
	"strconv"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	"github.com/bxcodec/faker/v4"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewBoolParam("e164_format")).
		Param(bloblang.NewBoolParam("include_hyphens"))

	// register the plugin
	err := bloblang.RegisterMethodV2("phonetransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		includeHyphens, err := args.GetBool("include_hyphens")
		if err != nil {
			return nil, err
		}

		e164, err := args.GetBool("e164_format")
		if err != nil {
			return nil, err
		}

		return bloblang.StringMethod(func(s string) (any, error) {
			res, err := ProcessPhoneNumber(s, preserveLength, e164, includeHyphens)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// main transformer logic goes here
func ProcessPhoneNumber(pn string, preserveLength, e164, includeHyphens bool) (string, error) {

	var returnValue string

	if preserveLength && !includeHyphens && !e164 {

		if strings.Contains(pn, "-") { // checks if input phone number has hyphens
			pn = strings.ReplaceAll(pn, "-", "")
		}

		const maxPhoneNum = 9

		val, err := GenerateRandomInt(0, maxPhoneNum, len(pn)) // generates len(pn) random numbers from 0 -> 9

		if err != nil {
			return "", nil
		}

		returnValue = strings.Join(IntArrToStringArr(val), "")

	} else if !preserveLength && includeHyphens && !e164 {
		// only works with 10 digit-based phone numbers like in the US
		returnValue = faker.Phonenumber()

	} else if !preserveLength && !includeHyphens && e164 {

		/* outputs in e164 format -> for ex. +873104859612, regex: ^\+[1-9]\d{1,14}$ */
		returnValue = faker.E164PhoneNumber()

	} else if e164 && preserveLength && !includeHyphens {

		val := strings.Split(pn, "+")

		p, err := ProcessPhoneNumber(val[1], true, false, false)
		if err != nil {
			return "", err
		}

		returnValue = "+" + p

	} else {

		// returns a phone number with no hyphens
		returnValue = strings.ReplaceAll(faker.Phonenumber(), "-", "")

		return returnValue, nil

	}

	return returnValue, nil
}

func IntArrToStringArr(ints []int) []string {

	var str []string

	for _, k := range ints {
		str = append(str, strconv.Itoa((k)))

	}

	return str
}
