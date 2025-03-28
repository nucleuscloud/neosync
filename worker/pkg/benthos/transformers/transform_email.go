package transformers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformEmail

type InvalidEmailAction string

const (
	InvalidEmailAction_Reject      InvalidEmailAction = "reject"
	InvalidEmailAction_Passthrough InvalidEmailAction = "passthrough"
	InvalidEmailAction_Null        InvalidEmailAction = "null"
	InvalidEmailAction_Generate    InvalidEmailAction = "generate"
)

func (g InvalidEmailAction) String() string {
	return string(g)
}

func isValidInvalidEmailAction(action string) bool {
	return action == string(InvalidEmailAction_Reject) ||
		action == string(InvalidEmailAction_Passthrough) ||
		action == string(InvalidEmailAction_Null) ||
		action == string(InvalidEmailAction_Generate)
}

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Anonymizes and transforms an existing email address.").
		Category("email").
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false).Description("Specifies the maximum length for the transformed data. This field ensures that the output does not exceed a certain number of characters.")).
		Param(bloblang.NewBoolParam("preserve_domain").Default(false).Description("A boolean indicating whether the domain part of the email should be preserved.")).
		Param(bloblang.NewAnyParam("excluded_domains").Default([]any{}).Description("A list of domains that should be excluded from the transformation")).
		Param(bloblang.NewInt64Param("max_length").Default(100).Description("Whether the original length of the input data should be preserved during transformation. If set to true, the transformation logic will ensure that the output data has the same length as the input data.")).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used for generating deterministic transformations.")).
		Param(bloblang.NewStringParam("email_type").Default(GenerateEmailType_UuidV4.String()).Description("Specifies the type of email to transform, with options including `uuidv4`, `fullname`, or `any`.")).
		Param(bloblang.NewStringParam("invalid_email_action").Default(InvalidEmailAction_Reject.String()).Description("Specifies the action to take when an invalid email is encountered, with options including `reject`, `passthrough`, `null`, or `generate`."))

	err := bloblang.RegisterFunctionV2(
		"transform_email",
		spec,
		func(args *bloblang.ParsedParams) (bloblang.Function, error) {
			emailPtr, err := args.GetOptionalString("value")
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

			invalidEmailActionArg, err := args.GetString("invalid_email_action")
			if err != nil {
				return nil, err
			}
			if !isValidInvalidEmailAction(invalidEmailActionArg) {
				return nil, errors.New("not a valid invalid_email_action argument")
			}

			invalidEmailAction := InvalidEmailAction(invalidEmailActionArg)

			seedArg, err := args.GetOptionalInt64("seed")
			if err != nil {
				return nil, err
			}

			seed, err := transformer_utils.GetSeedOrDefault(seedArg)
			if err != nil {
				return nil, err
			}
			randomizer := rng.New(seed)
			return func() (any, error) {
				output, err := transformEmail(randomizer, email, transformeEmailOptions{
					PreserveLength:     preserveLength,
					PreserveDomain:     preserveDomain,
					MaxLength:          maxLength,
					ExcludedDomains:    excludedDomains,
					EmailType:          emailType,
					InvalidEmailAction: invalidEmailAction,
				})
				if err != nil {
					return nil, fmt.Errorf("unable to run transform_email: %w", err)
				}
				return output, nil
			}, nil
		},
	)
	if err != nil {
		panic(err)
	}
}

func NewTransformEmailOptsFromConfig(
	config *mgmtv1alpha1.TransformEmail,
	maxLength *int64,
) (*TransformEmailOpts, error) {
	if config == nil {
		var excludedDomains any = "[]"
		return NewTransformEmailOpts(nil, nil, &excludedDomains, nil, nil, nil, nil)
	}
	var emailType *string
	if config.EmailType != nil {
		emailTypeStr := dtoEmailTypeToTransformerEmailType(config.GetEmailType()).String()
		emailType = &emailTypeStr
	}
	var invalidEmailAction *string
	if config.InvalidEmailAction != nil {
		invalidEmailActionStr := dtoInvalidEmailActionToTransformerInvalidEmailAction(
			config.GetInvalidEmailAction(),
		).String()
		invalidEmailAction = &invalidEmailActionStr
	}
	excludedDomainsStr, err := convertStringSliceToString(config.GetExcludedDomains())
	if err != nil {
		return nil, err
	}
	var excludedDomains any = excludedDomainsStr
	return NewTransformEmailOpts(
		config.PreserveLength,
		config.PreserveDomain,
		&excludedDomains,
		maxLength,
		nil,
		emailType,
		invalidEmailAction,
	)
}

func (t *TransformEmail) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformEmailOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	excludedDomains := []string{}
	if parsedOpts.excludedDomains != nil {
		switch v := parsedOpts.excludedDomains.(type) {
		case []any:
			exDomainsStrs, err := fromAnyToStringSlice(excludedDomains)
			if err != nil {
				return nil, errors.New("excludedDomains is not a []string")
			}
			excludedDomains = exDomainsStrs
		case []string:
			excludedDomains = v
		case string:
			css := strings.TrimSuffix(strings.TrimPrefix(v, "["), "]")
			excludedDomains = strings.Split(css, ",")
		default:
			return nil, fmt.Errorf("excludedDomains is of type %T, not []any or []string", v)
		}
	}

	return transformEmail(parsedOpts.randomizer, valueStr, transformeEmailOptions{
		PreserveLength:     parsedOpts.preserveLength,
		PreserveDomain:     parsedOpts.preserveDomain,
		MaxLength:          parsedOpts.maxLength,
		ExcludedDomains:    excludedDomains,
		EmailType:          GenerateEmailType(parsedOpts.emailType),
		InvalidEmailAction: InvalidEmailAction(parsedOpts.invalidEmailAction),
	})
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
	PreserveLength     bool
	PreserveDomain     bool
	MaxLength          int64
	ExcludedDomains    []string
	EmailType          GenerateEmailType
	InvalidEmailAction InvalidEmailAction
}

// Anonymizes an existing email address. This function returns a string pointer to handle nullable email columns where an input email value may not exist.
func transformEmail(
	randomizer rng.Rand,
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
	if emailType == GenerateEmailType_Any {
		emailType = getRandomEmailType(randomizer)
	}

	parsedInputEmail, err := mail.ParseAddress(email)
	if err != nil {
		switch opts.InvalidEmailAction {
		case InvalidEmailAction_Passthrough:
			return &email, nil
		case InvalidEmailAction_Null:
			return nil, nil
		case InvalidEmailAction_Generate:
			newEmail, err := generateRandomEmail(
				randomizer,
				opts.MaxLength,
				opts.EmailType,
				opts.ExcludedDomains,
			)
			if err != nil {
				return nil, err
			}
			return &newEmail, nil
		default: // Default or Reject
			return nil, fmt.Errorf("input email was not a valid email address: %w", err)
		}
	}

	_, domain, found := strings.Cut(parsedInputEmail.Address, "@")
	if !found {
		return nil, errors.New("did not find @ when parsing email address")
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
		return nil, fmt.Errorf(
			"for the given max length, unable to generate an email of sufficient length: %d",
			maxLength,
		)
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
	if emailType == GenerateEmailType_UuidV4 {
		newuuid := strings.ReplaceAll(uuid.NewString(), "-", "")
		trimmeduuid := transformer_utils.TrimStringIfExceeds(newuuid, maxNameLength)
		if trimmeduuid == "" {
			return nil, fmt.Errorf(
				"for the given max length, unable to use uuid to generate transformed email: %d",
				maxNameLength,
			)
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

func dtoEmailTypeToTransformerEmailType(dto mgmtv1alpha1.GenerateEmailType) GenerateEmailType {
	switch dto {
	case mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_FULLNAME:
		return GenerateEmailType_FullName
	default:
		return GenerateEmailType_UuidV4
	}
}

func dtoInvalidEmailActionToTransformerInvalidEmailAction(
	dto mgmtv1alpha1.InvalidEmailAction,
) InvalidEmailAction {
	switch dto {
	case mgmtv1alpha1.InvalidEmailAction_INVALID_EMAIL_ACTION_GENERATE:
		return InvalidEmailAction_Generate
	case mgmtv1alpha1.InvalidEmailAction_INVALID_EMAIL_ACTION_NULL:
		return InvalidEmailAction_Null
	case mgmtv1alpha1.InvalidEmailAction_INVALID_EMAIL_ACTION_PASSTHROUGH:
		return InvalidEmailAction_Passthrough
	default:
		return InvalidEmailAction_Reject
	}
}

func convertStringSliceToString(slc []string) (string, error) {
	var returnStr string

	if len(slc) == 0 {
		returnStr = "[]"
	} else {
		sliceBytes, err := json.Marshal(slc)
		if err != nil {
			return "", err
		}
		returnStr = string(sliceBytes)
	}
	return returnStr, nil
}
