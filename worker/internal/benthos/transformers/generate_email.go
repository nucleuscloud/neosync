package transformers

import (
	"math/rand"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

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

	err := bloblang.RegisterFunctionV2("generate_email", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		return func() (any, error) {

			res, err := GenerateEmail()
			return res, err
		}, nil

	})

	if err != nil {
		panic(err)
	}

}

/* Generates an email in the format <username@domaion.tld> such as jdoe@gmail.com */
func GenerateEmail() (string, error) {

	un, err := GenerateEmailUsername()
	if err != nil {
		return "", err
	}

	domain, err := GenerateEmailDomain()
	if err != nil {
		return "", err
	}

	return un + domain, nil
}

// Generates an email username for an email address either as <firstinitial><lastName> for ex. jdoe or <firstname>.<lastname> such as john.doe
func GenerateEmailUsername() (string, error) {

	//nolint
	// randomly generate a 0 or 1 in order to pick an email username format
	randValue := rand.Intn(2)

	if randValue == 1 {
		val, err := GenerateUsername()
		if err != nil {
			return "", err
		}

		return val, nil
	} else {
		fn, err := GenerateRandomFirstName(int64(23))
		if err != nil {
			return "", err
		}
		ln, err := GenerateRandomLastName(int64(4))
		if err != nil {
			return "", err
		}
		return fn + "." + ln, nil
	}

}

// Generates a realistic looking domain such as @gmail.com
func GenerateEmailDomain() (string, error) {
	//nolint
	randValue := rand.Intn(len(emailDomains))

	return "@" + emailDomains[randValue], nil

}
