package transformers_email

import (
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	"github.com/nucleuscloud/neosync/worker/internal/benthos/transformers"
)

var alphabet = "abcdefghijklmnopqrstuvwxyz"

var emailDomains = []string{
	"gmail.com",
	"yahoo.com",
	"hotmail.com",
	"aol.com",
	"hotmail.co",
	"hotmail.fr",
	"msn.com",
	"yahoo.fr",
	"wanadoo.fr",
	"orange.fr",
	"comcast.net",
	"yahoo.co.uk",
	"yahoo.com.br",
	"yahoo.co.in",
	"live.com",
	"rediffmail.com",
	"free.fr",
	"gmx.de",
	"web.de",
	"yandex.ru",
}

func init() {

	spec := bloblang.NewPluginSpec()

	err := bloblang.RegisterFunctionV2("generate_realistic_email", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {

			res, err := GenerateRealisticEmail()
			return res, err
		}, nil

	})

	if err != nil {
		panic(err)
	}

}

// Generates a realistic email in the format <username@domaion.tld> such as jdoe@gmail.com
func GenerateRealisticEmail() (string, error) {

	un, err := GenerateRealisticUsername()
	if err != nil {
		return "", err
	}

	domain, err := GenerateRealisticDomain()
	if err != nil {
		return "", err
	}

	return un + domain, nil
}

// Generates a realistic looking username for an email address either as firstinitial then lastName for ex. jdoe or firstname.lastname such as john.doe
func GenerateRealisticUsername() (string, error) {

	// randomly generate a 0 or 1 in order to pick an email format
	randValue := rand.Intn(2)

	if randValue == 1 {
		val, err := transformers.GenerateUsername()
		if err != nil {
			return "", err
		}

		return val, nil
	} else {
		fn, err := transformers.GenerateFirstNameWithRandomLength()
		if err != nil {
			return "", err
		}
		ln, err := transformers.GenerateLastNameWithRandomLength()
		if err != nil {
			return "", err
		}
		return fn + "." + ln, nil
	}

}

// Generates a realistic looking domain such as @gmail.com
func GenerateRealisticDomain() (string, error) {

	randValue := rand.Intn(len(emailDomains))

	return "@" + emailDomains[randValue], nil

}
