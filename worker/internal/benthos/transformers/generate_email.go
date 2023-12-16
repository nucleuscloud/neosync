package transformers

import (
	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

var tld = []string{"com", "org", "net", "edu", "gov", "app", "dev"}

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("min")).
		Param(bloblang.NewInt64Param("max"))

	err := bloblang.RegisterFunctionV2("generate_email", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		min, err := args.GetInt64("min")
		if err != nil {
			return nil, err
		}

		max, err := args.GetInt64("max")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {

			res, err := GenerateRandomEmail(min, max)
			return res, err
		}, nil

	})

	if err != nil {
		panic(err)
	}

}

// Generates a random email comprised of randomly sampled alphanumeric characters and returned in the format <username@domaion.tld>
func GenerateRandomEmail(min, max int64) (string, error) {

	un, err := GenerateRandomUsername(min, max)
	if err != nil {
		return "", err
	}

	domain, err := GenerateRandomDomain(min, max)
	if err != nil {
		return "", err
	}

	return un + domain, nil
}

// Generates a random username comprised of randomly sampled alphanumeric characters
func GenerateRandomUsername(min, max int64) (string, error) {

	username, err := transformer_utils.GenerateRandomString(min, max)
	if err != nil {
		return "", err
	}

	return username, nil

}

// Generates a random domain comprised of randomly sampled alphanumeric characters in the format <@domain.tld>
func GenerateRandomDomain(min, max int64) (string, error) {

	var result string

	domain, err := transformer_utils.GenerateRandomString(min, max)
	if err != nil {
		return "", err
	}

	tld, err := transformer_utils.GetRandomValueFromSlice(tld)
	if err != nil {
		return "", err
	}

	result = "@" + domain + "." + tld

	return result, nil

}
