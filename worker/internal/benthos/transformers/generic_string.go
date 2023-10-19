package neosync_transformers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

const defaultStrLength = 10

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewInt64Param("str_length")).
		Param(bloblang.NewStringParam("case"))
		// Param(bloblang.NewBoolParam("char_set")).

	// register the plugin
	err := bloblang.RegisterMethodV2("genericstringtransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		strLength, err := args.GetInt64("str_length")
		if err != nil {
			return nil, err
		}

		strCase, err := args.GetString("case")
		if err != nil {
			return nil, err
		}

		cased, err := StrCaseFromString(strCase)

		if err != nil {
			return nil, fmt.Errorf("unable to convert the string case to a defined enum value")
		}

		return bloblang.StringMethod(func(s string) (any, error) {
			res, err := ProcessRandomString(s, preserveLength, strLength, cased)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// main transformer logic goes here
func ProcessRandomString(s string, preserveLength bool, strLength int64, strCase mgmtv1alpha1.RandomString_StringCase) (string, error) {
	var returnValue string

	if preserveLength {

		val, err := GenerateRandomStringWithLength(int64(len(s)))

		if err != nil {
			return "", fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	} else if strLength > 0 {

		val, err := GenerateRandomStringWithLength(strLength)

		if err != nil {
			return "", fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	} else if preserveLength && strLength > 0 {

		val, err := GenerateRandomStringWithLength(strLength)

		if err != nil {
			return "", fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	} else {

		val, err := GenerateRandomStringWithLength(defaultStrLength)

		if err != nil {
			return "", fmt.Errorf("unable to generate a random string with length")
		}

		returnValue = val

	}

	switch strCase {
	case mgmtv1alpha1.RandomString_STRING_CASE_LOWER:
		returnValue = strings.ToLower(returnValue)
	case mgmtv1alpha1.RandomString_STRING_CASE_UPPER:
		returnValue = strings.ToUpper(returnValue)
	case mgmtv1alpha1.RandomString_STRING_CASE_TITLE:
		returnValue = strings.ToTitle(returnValue)
	}

	return returnValue, nil
}

func GenerateRandomStringWithLength(l int64) (string, error) {

	const alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

	if l <= 0 {
		return "", fmt.Errorf("the length cannot be zero or negative")
	}

	// Create a random source using crypto/rand
	source := rand.Reader

	// Calculate the max index in the alphabet string
	maxIndex := big.NewInt(int64(len(alphanumeric)))

	result := make([]byte, l)

	for i := int64(0); i < l; i++ {
		// Generate a random index in the range [0, len(alphabet))
		index, err := rand.Int(source, maxIndex)
		if err != nil {
			return "", fmt.Errorf("unable to generate a random index for random string generation")
		}

		// Get the character at the generated index and append it to the result
		result[i] = alphanumeric[index.Int64()]
	}

	return string(result), nil

}

func StrCaseFromString(strCase string) (mgmtv1alpha1.RandomString_StringCase, error) {
	switch strCase {
	case "UPPER":
		return mgmtv1alpha1.RandomString_STRING_CASE_UPPER, nil
	case "LOWER":
		return mgmtv1alpha1.RandomString_STRING_CASE_LOWER, nil
	case "TITLE":
		return mgmtv1alpha1.RandomString_STRING_CASE_TITLE, nil
	default:
		return mgmtv1alpha1.RandomString_STRING_CASE_UPPER, fmt.Errorf("invalid string case: %s", strCase)
	}
}
