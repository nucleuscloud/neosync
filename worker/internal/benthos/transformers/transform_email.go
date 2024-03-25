package transformers

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/mail"
	"strings"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/google/uuid"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("email").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false)).
		Param(bloblang.NewBoolParam("preserve_domain").Default(false)).
		Param(bloblang.NewAnyParam("excluded_domains").Default([]any{})).
		Param(bloblang.NewInt64Param("max_length").Default(10000)).
		Param(bloblang.NewInt64Param("seed").Optional()).
		Param(bloblang.NewStringParam("email_type").Default(fullNameEmailType))

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

		excludedDomainsArg, err := args.Get("excluded_domains")
		if err != nil {
			return nil, err
		}

		excludedDomains, err := fromAnyToStringSlice(excludedDomainsArg)
		if err != nil {
			return nil, err
		}

		emailTypeArg, err := args.GetString("email_type")
		if err != nil {
			return nil, err
		}
		emailType := getEmailTypeOrDefault(emailTypeArg)

		seedArg, err := args.GetOptionalInt64("seed")
		if err != nil {
			return nil, err
		}
		var seed int64
		if seedArg != nil {
			seed = *seedArg
		} else {
			// we want a bit more randomness here with generate_email so using something that isn't time based
			var err error
			seed, err = transformer_utils.GenerateCryptoSeed()
			if err != nil {
				return nil, err
			}
		}
		randomizer := rand.New(rand.NewSource(seed)) //nolint:gosec
		return func() (any, error) {
			output, err := transformEmail(randomizer, email, transformeEmailOptions{
				PreserveLength:  preserveLength,
				PreserveDomain:  preserveDomain,
				MaxLength:       maxLength,
				ExcludedDomains: excludedDomains,
				EmailType:       emailType,
			})
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_email: %w", err)
			}
			return output, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func fromAnyToStringSlice(input any) ([]string, error) {
	var output []string
	if input == nil {
		return output, nil
	}
	anySlice, ok := input.([]any)
	if ok {
		for _, anyValue := range anySlice {
			value, ok := anyValue.(string)
			if !ok {
				return nil, fmt.Errorf("expected string, got %T", anyValue)
			}
			output = append(output, value)
		}
		return output, nil
	}
	stringSlice, ok := input.([]string)
	if ok {
		return stringSlice, nil
	}
	return nil, fmt.Errorf("unable to cast arg to []any or []string, got %T", input)
}

type transformeEmailOptions struct {
	PreserveLength  bool
	PreserveDomain  bool
	MaxLength       int64
	ExcludedDomains []string
	EmailType       generateEmailType
}

// Anonymizes an existing email address. This function returns a string pointer to handle nullable email columns where an input email value may not exist.
func transformEmail(
	randomizer *rand.Rand,
	email string,
	opts transformeEmailOptions,
) (*string, error) {
	if email == "" {
		return nil, nil
	}
	if opts.MaxLength <= 0 {
		opts.MaxLength = math.MaxInt64
	}
	emailType := opts.EmailType
	if emailType == anyEmailType {
		emailType = getRandomEmailType(randomizer)
	}

	parsedInputEmail, err := mail.ParseAddress(email)
	if err != nil {
		return nil, fmt.Errorf("input email was not a valid email address: %w", err)
	}

	_, domain, found := strings.Cut(parsedInputEmail.Address, "@")
	if !found {
		return nil, errors.New("did not found @ when parsed email address")
	}

	excludedDomainsSet := transformer_utils.ToSet(opts.ExcludedDomains)

	preserveDomain := opts.PreserveDomain
	_, isDomainExcluded := excludedDomainsSet[domain]
	if preserveDomain && isDomainExcluded {
		// preserve, but domain is excluded, so for this input it should be false
		preserveDomain = false
	} else if !preserveDomain && isDomainExcluded {
		// not preserve, but domain is excluded, so for this input it should be true
		preserveDomain = true
	}
	maxLength := opts.MaxLength

	domainMaxLength := maxLength - 3 // is there enough room for at least two characters and an @ sign
	if opts.PreserveLength {
		domainMaxLength = int64(len(email)) - 3
	}
	if (domainMaxLength) <= 0 {
		return nil, fmt.Errorf("for the given max length, unable to generate an email of sufficient length: %d", maxLength)
	}

	newdomain := domain
	if !preserveDomain {
		// generate a new domain, but do not generate any that are in the excluded domains list
		randomdomain, err := getRandomEmailDomain(randomizer, domainMaxLength, opts.ExcludedDomains)
		if err != nil {
			return nil, err
		}
		newdomain = randomdomain
	}

	maxNameLength := maxLength - int64(len(newdomain)) - 1
	var minNameLength *int64
	if opts.PreserveLength {
		minLength := int64(len(email)) - int64(len(newdomain)) - 1
		maxNameLength = minLength
		minNameLength = &minLength
	}

	var newname string
	if emailType == uuidV4EmailType {
		newuuid := strings.ReplaceAll(uuid.NewString(), "-", "")
		trimmeduuid := transformer_utils.TrimStringIfExceeds(newuuid, maxNameLength)
		if trimmeduuid == "" {
			return nil, fmt.Errorf("for the given max length, unable to use uuid to generate transformed email: %d", maxNameLength)
		}
		newname = trimmeduuid
	} else {
		name, err := generateNameForEmail(randomizer, minNameLength, maxNameLength)
		if err != nil {
			return nil, fmt.Errorf("for the given max length, unable to generate a full name to generate transformed email: %d", maxNameLength)
		}
		newname = name
	}

	generatedemail := fmt.Sprintf("%s@%s", newname, newdomain)
	return &generatedemail, nil
}
