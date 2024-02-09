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
		Param(bloblang.NewAnyParam("email").Optional()).
		Param(bloblang.NewBoolParam("preserve_length")).
		Param(bloblang.NewBoolParam("preserve_domain")).
		Param(bloblang.NewAnyParam("exclusion_list")).
		Param(bloblang.NewInt64Param("max_length"))

	err := bloblang.RegisterFunctionV2("transform_email", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		emailPtr, err := args.GetOptionalString("email")
		if err != nil {
			return nil, err
		}

		var email string
		if emailPtr != nil {
			email = *emailPtr
		}

		preserveLength, err := args.GetBool("preserve_length")
		if err != nil {
			return nil, err
		}

		preserveDomain, err := args.GetBool("preserve_domain")
		if err != nil {
			return nil, err
		}

		maxLength, err := args.GetInt64("max_length")
		if err != nil {
			return nil, err
		}

		eL, err := args.Get("exclusion_list")
		if err != nil {
			return nil, err
		}

		excl, ok := eL.([]any)
		if !ok {
			return nil, fmt.Errorf("unable to cast arg to any slice")
		}

		var excludeStringSlice []string

		for _, str := range excl {
			val, ok := str.(string)
			if !ok {
				return nil, fmt.Errorf("expected string, got :%T", str)

			}

			excludeStringSlice = append(excludeStringSlice, val)
		}

		return func() (any, error) {

			res, err := TransformEmail(email, preserveLength, preserveDomain, maxLength, excludeStringSlice)
			return res, err
		}, nil

	})

	if err != nil {
		panic(err)
	}

}

// Anonymizes an existing email address. This function returns a string pointer to handle nullable email columns where an input email value may not exist.
func TransformEmail(email string, preserveLength, preserveDomain bool, maxLength int64, excludeList []string) (*string, error) {

	var returnValue string
	var err error

	if email == "" {
		return nil, nil
	}

	if !preserveLength && preserveDomain {

		returnValue, err = TransformEmailPreserveDomain(email, true, maxLength, excludeList)
		if err != nil {
			return nil, err
		}

	} else if preserveLength && !preserveDomain {

		returnValue, err = TransformEmailPreserveLength(email, excludeList)
		if err != nil {
			return nil, err
		}

	} else if preserveLength && preserveDomain {

		returnValue, err = TransformEmailPreserveDomainAndLength(email, maxLength, excludeList)
		if err != nil {
			return nil, err
		}

	} else {

		randLength, err := transformer_utils.GenerateRandomInt64InValueRange(10, maxLength)
		if err != nil {
			return nil, err
		}

		randomEmail, err := GenerateRandomEmail(randLength)
		if err != nil {
			return nil, err
		}

		parsedInputEmail, err := transformer_utils.ParseEmail(email)
		if err != nil {
			return nil, fmt.Errorf("invalid email: %s", email)
		}

		parsedGeneratedEmail, err := transformer_utils.ParseEmail(randomEmail)
		if err != nil {
			return nil, fmt.Errorf("invalid email: %s", email)
		}

		un := parsedGeneratedEmail[0]

		// handle exclusion list
		if transformer_utils.StringInSlice(parsedInputEmail[1], excludeList) {

			domain := parsedInputEmail[1]

			returnValue = un + "@" + domain

		} else {
			splitDomain := strings.Split(parsedInputEmail[1], ".")

			// generate a random domain
			dom, err := transformer_utils.GenerateRandomStringWithDefinedLength(int64(len(splitDomain[0])))
			if err != nil {
				return nil, err
			}

			// generate a random top level domain
			tld, err := transformer_utils.GenerateRandomStringWithDefinedLength(int64(len(splitDomain[1])))
			if err != nil {
				return nil, err
			}
			domain := dom + "." + tld

			returnValue = un + "@" + domain
		}

	}

	return &returnValue, nil
}

// Generate a random email and preserve the input email's domain
func TransformEmailPreserveDomain(email string, pd bool, maxLength int64, excludeList []string) (string, error) {

	parsedEmail, err := transformer_utils.ParseEmail(email)
	if err != nil {
		return "", fmt.Errorf("invalid email: %s", email)
	}

	// handle exclusion list
	if transformer_utils.StringInSlice(parsedEmail[1], excludeList) {

		randLength, err := transformer_utils.GenerateRandomInt64InValueRange(10, maxLength)
		if err != nil {
			return "", err
		}

		e, err := GenerateRandomEmail(randLength)
		if err != nil {
			return "", err
		}

		return e, nil
	} else {
		// generate a random username and preserve the domain
		un, err := GenerateUsername(int64(len(parsedEmail[0])))
		if err != nil {
			return "", nil
		}
		return strings.ToLower(un) + "@" + parsedEmail[1], err
	}

}

// Preserve the length of email but not the domain name. If domain is in the exclusion list, then it will preserve the domain
func TransformEmailPreserveLength(email string, excludeList []string) (string, error) {

	var res string

	parsedEmail, err := transformer_utils.ParseEmail(email)
	if err != nil {
		return "", fmt.Errorf("invalid email: %s", email)
	}

	// generate a random username for the email address
	un, err := transformer_utils.GenerateRandomStringWithDefinedLength(int64(len(parsedEmail[0])))
	if err != nil {
		return "", nil
	}

	// split the domain to account for different domain name lengths
	splitDomain := strings.Split(parsedEmail[1], ".")

	var domain string

	// handle exclusion list
	if transformer_utils.StringInSlice(parsedEmail[1], excludeList) {

		domain = parsedEmail[1]

	} else {

		// generate a random domain
		dom, err := transformer_utils.GenerateRandomStringWithDefinedLength(int64(len(splitDomain[0])))
		if err != nil {
			return "", nil
		}

		// generate a random top level domain
		tld, err := transformer_utils.GenerateRandomStringWithDefinedLength(int64(len(splitDomain[1])))
		if err != nil {
			return "", nil
		}
		domain = dom + "." + tld

	}

	res = transformer_utils.SliceString(un, len(parsedEmail[0])) + "@" + domain

	return res, err

}

// preserve domain and length of the email -> keep the domain the same but slice the username to be the same length as the input username
func TransformEmailPreserveDomainAndLength(e string, maxLength int64, excludeList []string) (string, error) {

	parsedEmail, err := transformer_utils.ParseEmail(e)
	if err != nil {
		return "", fmt.Errorf("invalid email: %s", e)
	}

	unLength := len(parsedEmail[0])

	// generate a random username for the email address
	un, err := transformer_utils.GenerateRandomStringWithDefinedLength(int64(unLength))
	if err != nil {
		return "", nil
	}

	// handle exclusion list
	if transformer_utils.StringInSlice(parsedEmail[1], excludeList) {

		splitDomain := strings.Split(parsedEmail[1], ".")

		// generate a random domain
		dom, err := transformer_utils.GenerateRandomStringWithDefinedLength(int64(len(splitDomain[0])))
		if err != nil {
			return "", nil
		}

		// generate a random top level domain
		tld, err := transformer_utils.GenerateRandomStringWithDefinedLength(int64(len(splitDomain[1])))
		if err != nil {
			return "", nil
		}
		domain := dom + "." + tld

		res := un + "@" + domain

		return res, nil

	} else {
		// generate a random username and preserve the domain
		un, err := GenerateUsername(int64(len(parsedEmail[0])))
		if err != nil {
			return "", nil
		}
		return strings.ToLower(un) + "@" + parsedEmail[1], err
	}
}
