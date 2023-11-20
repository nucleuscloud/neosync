package neosync_transformers

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

var tld = []string{"com", "org", "net", "edu", "gov", "app", "dev"}

func init() {

	spec := bloblang.NewPluginSpec().Param(bloblang.NewStringParam("email")).Param(bloblang.NewBoolParam("preserve_length")).Param(bloblang.NewBoolParam("preserve_domain"))

	// register the plugin
	err := bloblang.RegisterFunctionV2("emailtransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		email, err := args.GetString("email")
		if err != nil {
			return nil, err
		}

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		preserveDomain, err := args.GetBool("preserve_domain")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateEmail(email, preserveLength, preserveDomain)
			return res, err
		}, nil

	})

	if err != nil {
		panic(err)
	}

}

// generates a random email address
func GenerateEmail(email string, preserveLength, preserveDomain bool) (string, error) {

	var returnValue string
	var err error

	if email != "" {
		if !preserveLength && preserveDomain {

			returnValue, err = GenerateEmailPreserveDomain(email, true)
			if err != nil {
				return "", err
			}

		} else if preserveLength && !preserveDomain {

			returnValue, err = GenerateEmailPreserveLength(email, true)
			if err != nil {
				return "", err
			}

		} else if preserveLength && preserveDomain {

			returnValue, err = GenerateEmailPreserveDomainAndLength(email, true, true)
			if err != nil {
				return "", err
			}

		} else {
			e, err := GenerateRandomEmail()
			if err != nil {
				return "", nil
			}

			returnValue = e
		}

	} else {

		e, err := GenerateRandomEmail()
		if err != nil {
			return "", nil
		}

		returnValue = e
	}

	return returnValue, nil
}

func GenerateRandomEmail() (string, error) {
	un, err := GenerateRandomUsername()
	if err != nil {
		return "", nil
	}

	domain, err := GenerateDomain()
	if err != nil {
		return "", nil
	}

	// generate random email
	return un + domain, err
}

// Generate a random email and preserve the input email's domain
func GenerateEmailPreserveDomain(e string, pd bool) (string, error) {

	parsedEmail, err := parseEmail(e)
	if err != nil {
		return "", fmt.Errorf("invalid email: %s", e)
	}

	un, err := GenerateRandomUsername()
	if err != nil {
		return "", nil
	}

	return strings.ToLower(un) + "@" + parsedEmail[1], err
}

// Preserve the length of email but not the domain name
func GenerateEmailPreserveLength(e string, pl bool) (string, error) {

	var res string

	parsedEmail, err := parseEmail(e)
	if err != nil {
		return "", fmt.Errorf("invalid email: %s", e)
	}

	// split the domain to account for different domain name lengths
	splitDomain := strings.Split(parsedEmail[1], ".")

	domain, err := GenerateDomain()
	if err != nil {
		return "", err
	}

	splitGeneratedDomain := strings.Split(domain, ".")

	// the +1 is because we include an @ sign we include in the domain and we want to keep that
	domainName := transformer_utils.SliceString(splitGeneratedDomain[0], len(splitDomain[0])+1)

	tld := transformer_utils.SliceString(splitGeneratedDomain[1], len(splitDomain[1]))

	un, err := transformer_utils.GenerateRandomStringWithLength(int64(len(parsedEmail[0])))
	if err != nil {
		return "", nil
	}

	res = transformer_utils.SliceString(un, len(parsedEmail[0])) + domainName + "." + tld

	return res, err

}

// preserve domain and length of the email -> keep the domain the same but slice the username to be the same length as the input username
func GenerateEmailPreserveDomainAndLength(e string, pd, pl bool) (string, error) {

	parsedEmail, err := parseEmail(e)
	if err != nil {
		return "", fmt.Errorf("invalid email: %s", e)
	}

	unLength := len(parsedEmail[0])

	un, err := transformer_utils.GenerateRandomStringWithLength(int64(len(parsedEmail[0])))
	if err != nil {
		return "", err
	}

	res := transformer_utils.SliceString(un, unLength) + "@" + parsedEmail[1]

	return res, err
}

func GenerateDomain() (string, error) {

	var result string

	domain, err := transformer_utils.GenerateRandomStringWithLength(6)
	if err != nil {
		return "", fmt.Errorf("unable to generate random domain name")
	}

	tld, err := transformer_utils.GetRandomValueFromSlice(tld)
	if err != nil {
		return "", err
	}

	result = "@" + domain + "." + tld

	return result, err

}

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

func parseEmail(email string) ([]string, error) {

	inputEmail, err := mail.ParseAddress(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email format: %s", email)
	}

	parsedEmail := strings.Split(inputEmail.Address, "@")

	return parsedEmail, nil
}
