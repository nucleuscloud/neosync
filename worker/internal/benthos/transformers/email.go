package neosync_transformers

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	"github.com/bxcodec/faker/v4"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewBoolParam("preserve_domain"))

	// register the plugin
	err := bloblang.RegisterMethodV2("emailtransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Method, error) {

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		preserveDomain, err := args.GetBool("preserve_domain")
		if err != nil {
			return nil, err
		}

		return bloblang.StringMethod(func(s string) (any, error) {
			res, err := ProcessEmail(s, preserveLength, preserveDomain)
			return res, err
		}), nil
	})

	if err != nil {
		panic(err)
	}

}

// main transformer logic goes here
func ProcessEmail(email string, preserveLength, preserveDomain bool) (string, error) {

	parsedEmail, err := parseEmail(email)
	if err != nil {
		return "", fmt.Errorf("invalid email: %s", email)
	}

	var returnValue string

	if preserveDomain && !preserveLength {

		returnValue = strings.ToLower(faker.Username()) + "@" + parsedEmail[1]

	} else if preserveLength && !preserveDomain {

		// preserve length of email but not the domain

		splitDomain := strings.Split(parsedEmail[1], ".") // split the domain to account for different domain name lengths

		domain := sliceString(faker.DomainName(), len(splitDomain[0]))

		tld := sliceString(faker.DomainName(), len(splitDomain[1]))

		returnValue = sliceString(faker.Username(), len(parsedEmail[0])) + "@" + domain + "." + tld

	} else if preserveDomain && preserveLength {

		// preserve domain and length of the email -> keep the domain the same but slice the username to be the same length as the input username
		unLength := len(parsedEmail[0])

		un := faker.Username()

		returnValue = sliceString(un, unLength) + "@" + parsedEmail[1]

	} else {
		// generate random email

		returnValue = faker.Email()
	}

	return returnValue, nil
}

func parseEmail(email string) ([]string, error) {

	inputEmail, err := mail.ParseAddress(email)
	if err != nil {

		return nil, fmt.Errorf("invalid email format: %s", email)
	}

	parsedEmail := strings.Split(inputEmail.Address, "@")

	return parsedEmail, nil
}

func sliceString(s string, l int) string {

	runes := []rune(s) // use runes instead of strings in order to avoid slicing a multi-byte character and returning invalid UTF-8

	if l > len(runes) {
		l = len(runes)
	}

	return string(runes[:l])
}
