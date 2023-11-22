package transformers

import (
	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_random_email", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {

			res, err := GenerateRandomEmail()
			return res, err
		}, nil

	})

	if err != nil {
		panic(err)
	}

}

// Generates a random email comprised of randomly sampled alphanumeric characters and returned in the format <username@domaion.tld>
func GenerateRandomEmail() (string, error) {

	un, err := GenerateRandomUsername()
	if err != nil {
		return "", err
	}

	domain, err := GenerateRandomDomain()
	if err != nil {
		return "", err
	}

	return un + domain, nil
}

// Generates a random username comprised of randomly sampled alphanumeric characters
func GenerateRandomUsername() (string, error) {

	randLength, err := transformer_utils.GenerateRandomIntWithBounds(3, 8)
	if err != nil {
		return "", err
	}

	username, err := transformer_utils.GenerateRandomStringWithLength(int64(randLength))
	if err != nil {
		return "", err
	}

	return username, nil

}

// Generates a random domain comprised of randomly sampled alphanumeric characters in the format <@domain.tld>
func GenerateRandomDomain() (string, error) {

	var result string

	domain, err := transformer_utils.GenerateRandomStringWithLength(6)
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
