package transformers

import (
	"fmt"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("value").Optional()). // this needs to be an AnyParam or else if it's null, benthos will throw an error expecting a type
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewBoolParam("preserve_domain"))

	err := bloblang.RegisterFunctionV2("transform_email", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

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

		preserveDomain, err := args.GetBool("preserve_domain")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {

			res, err := TransformEmail(value, preserveLength, preserveDomain)
			return res, err
		}, nil

	})

	if err != nil {
		panic(err)
	}

}

// return a string pointer so if the email field is empty then it returns a null value
func TransformEmail(email string, preserveLength, preserveDomain bool) (*string, error) {

	var returnValue string
	var err error

	if email == "" {
		return nil, nil
	}

	if !preserveLength && preserveDomain {

		returnValue, err = TransformEmailPreserveDomain(email, true)
		if err != nil {
			return nil, err
		}

	} else if preserveLength && !preserveDomain {

		returnValue, err = TransformEmailPreserveLength(email, true)
		if err != nil {
			return nil, err
		}

	} else if preserveLength && preserveDomain {

		returnValue, err = TransformEmailPreserveDomainAndLength(email, true, true)
		if err != nil {
			return nil, err
		}

	} else {

		min := int64(2)
		max := int64(5)
		e, err := GenerateRandomEmail(min, max)
		if err != nil {
			return nil, nil
		}

		returnValue = e
	}

	return &returnValue, nil
}

// Generate a random email and preserve the input email's domain
func TransformEmailPreserveDomain(e string, pd bool) (string, error) {

	parsedEmail, err := transformer_utils.ParseEmail(e)
	if err != nil {
		return "", fmt.Errorf("invalid email: %s", e)
	}

	min := int64(2)
	max := int64(5)

	un, err := GenerateRandomUsername(min, max)
	if err != nil {
		return "", nil
	}

	return strings.ToLower(un) + "@" + parsedEmail[1], err
}

// Preserve the length of email but not the domain name
func TransformEmailPreserveLength(e string, pl bool) (string, error) {

	var res string

	parsedEmail, err := transformer_utils.ParseEmail(e)
	if err != nil {
		return "", fmt.Errorf("invalid email: %s", e)
	}

	// split the domain to account for different domain name lengths
	splitDomain := strings.Split(parsedEmail[1], ".")

	min := int64(2)
	max := int64(5)

	domain, err := GenerateRandomDomain(min, max)
	if err != nil {
		return "", err
	}

	splitGeneratedDomain := strings.Split(domain, ".")

	// the +1 is because we include an @ sign we include in the domain and we want to keep that
	domainName := transformer_utils.SliceString(splitGeneratedDomain[0], len(splitDomain[0])+1)

	tld := transformer_utils.SliceString(splitGeneratedDomain[1], len(splitDomain[1]))

	//un, err := transformer_utils.GenerateRandomString(int64(len(parsedEmail[0])))

	un, err := transformer_utils.GenerateRandomString(min, max)
	if err != nil {
		return "", nil
	}

	res = transformer_utils.SliceString(un, len(parsedEmail[0])) + domainName + "." + tld

	return res, err

}

// preserve domain and length of the email -> keep the domain the same but slice the username to be the same length as the input username
func TransformEmailPreserveDomainAndLength(e string, pd, pl bool) (string, error) {

	parsedEmail, err := transformer_utils.ParseEmail(e)
	if err != nil {
		return "", fmt.Errorf("invalid email: %s", e)
	}

	unLength := len(parsedEmail[0])

	//un, err := transformer_utils.GenerateRandomString(int64(len(parsedEmail[0])))

	min := int64(2)
	max := int64(5)
	un, err := transformer_utils.GenerateRandomString(min, max)
	if err != nil {
		return "", err
	}

	res := transformer_utils.SliceString(un, unLength) + "@" + parsedEmail[1]

	return res, err
}
