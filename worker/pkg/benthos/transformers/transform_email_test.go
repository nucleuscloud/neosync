package transformers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/mail"
	"strings"
	"testing"
	"time"

	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

var email = "evis@gmail.com"
var maxEmailCharLimit = int64(40)
var excludedDomains = []string{"gmail.com", "hotmail.com"}

func Test_TransformEmail_Empty_Email(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := transformEmail(randomizer, "", transformeEmailOptions{})
	require.NoError(t, err)
	require.Nil(t, res)
}

func Test_TransformEmail_Empty_Options(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotEmpty(t, res)

	_, err = mail.ParseAddress(*res)
	require.NoError(t, err)
}

func Test_TransformEmail_Seed_1711240985047220000_Specific_Options(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1711240985047220000))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{MaxLength: 40, EmailType: GenerateEmailType_FullName, ExcludedDomains: excludedDomains, PreserveLength: true, PreserveDomain: true})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotEmpty(t, res)

	_, err = mail.ParseAddress(*res)
	require.NoError(t, err)
}

func Test_TransformEmail_Invalid_Email_Input(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := transformEmail(randomizer, "bademail", transformeEmailOptions{})
	require.Error(t, err)
	require.Nil(t, res)
}

func Test_TransformEmail_Random_Seed(t *testing.T) {
	seed := time.Now().UnixNano()
	randomizer := rand.New(rand.NewSource(seed))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{})
	require.NoError(t, err, "failed with seed", "seed", seed)
	require.NotNil(t, res)
	require.NotEmpty(t, res)

	_, err = mail.ParseAddress(*res)
	require.NoError(t, err)
}

func Test_TransformEmail_Any_EmailType(t *testing.T) {
	randomizer := rand.New(rand.NewSource(time.Now().UnixMicro()))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{
		EmailType: GenerateEmailType_Any,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotEmpty(t, res)

	_, err = mail.ParseAddress(*res)
	require.NoError(t, err)
}

func Test_TransformEmail_Uuid_EmailType(t *testing.T) {
	randomizer := rand.New(rand.NewSource(time.Now().UnixMicro()))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{
		EmailType: GenerateEmailType_UuidV4,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotEmpty(t, res)

	_, err = mail.ParseAddress(*res)
	require.NoError(t, err)
}

func Test_TransformEmail_Fullname_EmailType(t *testing.T) {
	randomizer := rand.New(rand.NewSource(time.Now().UnixMicro()))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{
		EmailType: GenerateEmailType_FullName,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotEmpty(t, res)

	_, err = mail.ParseAddress(*res)
	require.NoError(t, err)
}

func Test_TransformEmail_PreserveLength_False_PreserveDomain_False(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{
		PreserveLength: false,
		PreserveDomain: false,
		MaxLength:      maxEmailCharLimit,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")

	_, err = mail.ParseAddress(*res)
	require.NoError(t, err)
}

func Test_TransformEmail_PreserveLength_False_PreserveDomain_False_Excluded(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{
		PreserveLength:  false,
		PreserveDomain:  false,
		MaxLength:       maxEmailCharLimit,
		ExcludedDomains: excludedDomains,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")

	address, err := mail.ParseAddress(*res)
	require.NoError(t, err)

	_, domain, found := strings.Cut(address.Address, "@")
	require.True(t, found)
	require.Equal(t, "gmail.com", domain)
}

func Test_TransformEmail_PreserveLength_False_PreserveDomain_True(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{
		PreserveLength: false,
		PreserveDomain: true,
		MaxLength:      maxEmailCharLimit,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")

	address, err := mail.ParseAddress(*res)
	require.NoError(t, err)

	_, domain, found := strings.Cut(address.Address, "@")
	require.True(t, found)
	require.Equal(t, "gmail.com", domain)
}

func Test_TransformEmail_PreserveLength_False_PreserveDomain_True_Excluded(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{
		PreserveLength:  false,
		PreserveDomain:  true,
		MaxLength:       maxEmailCharLimit,
		ExcludedDomains: excludedDomains,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")

	address, err := mail.ParseAddress(*res)
	require.NoError(t, err)

	_, domain, found := strings.Cut(address.Address, "@")
	require.True(t, found)
	require.NotEqual(t, "gmail.com", domain)
}

func Test_TransformEmail_PreserveLength_True_PreserveDomain_False(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{
		PreserveLength: true,
		PreserveDomain: false,
		MaxLength:      maxEmailCharLimit,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")

	_, err = mail.ParseAddress(*res)
	require.NoError(t, err)

	require.Len(t, *res, len(email))
}

func Test_TransformEmail_PreserveLength_True_PreserveDomain_False_Excluded(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{
		PreserveLength:  true,
		PreserveDomain:  false,
		MaxLength:       maxEmailCharLimit,
		ExcludedDomains: excludedDomains,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")

	address, err := mail.ParseAddress(*res)
	require.NoError(t, err)

	require.Len(t, *res, len(email))

	_, domain, found := strings.Cut(address.Address, "@")
	require.True(t, found)
	require.Equal(t, "gmail.com", domain)
}

func Test_TransformEmail_PreserveLength_True_PreserveDomain_True(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{
		PreserveLength: true,
		PreserveDomain: true,
		MaxLength:      maxEmailCharLimit,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")

	address, err := mail.ParseAddress(*res)
	require.NoError(t, err)

	require.Len(t, *res, len(email))

	_, domain, found := strings.Cut(address.Address, "@")
	require.True(t, found)
	require.Equal(t, "gmail.com", domain)
}

func Test_TransformEmail_PreserveLength_True_PreserveDomain_True_Excluded(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))
	res, err := transformEmail(randomizer, email, transformeEmailOptions{
		PreserveLength:  true,
		PreserveDomain:  true,
		MaxLength:       maxEmailCharLimit,
		ExcludedDomains: excludedDomains,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")

	address, err := mail.ParseAddress(*res)
	require.NoError(t, err)

	require.Len(t, *res, len(email))

	_, domain, found := strings.Cut(address.Address, "@")
	require.True(t, found)
	require.NotEqual(t, "gmail.com", domain)
}

func Test_TransformEmail_InvalidEmailArg(t *testing.T) {
	randomizer := rand.New(rand.NewSource(1))

	invalidemail := "invalid@gmail..com"

	output, err := transformEmail(randomizer, invalidemail, transformeEmailOptions{})
	require.Error(t, err)
	require.Nil(t, output)

	output, err = transformEmail(randomizer, invalidemail, transformeEmailOptions{
		InvalidEmailAction: InvalidEmailAction_Reject,
	})
	require.Error(t, err)
	require.Nil(t, output)

	output, err = transformEmail(randomizer, invalidemail, transformeEmailOptions{
		InvalidEmailAction: InvalidEmailAction_Passthrough,
	})
	require.NoError(t, err)
	require.NotNil(t, output)
	require.Equal(t, invalidemail, *output)

	output, err = transformEmail(randomizer, invalidemail, transformeEmailOptions{
		InvalidEmailAction: InvalidEmailAction_Generate,
	})
	require.NoError(t, err)
	require.NotNil(t, output)

	output, err = transformEmail(randomizer, invalidemail, transformeEmailOptions{
		InvalidEmailAction: InvalidEmailAction_Null,
	})
	require.NoError(t, err)
	require.Nil(t, output)
}

func Test_Bloblang_transform_email_empty_opts(t *testing.T) {
	mapping := `root = transform_email()`
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	require.NoError(t, err)
	require.Nil(t, res)
}

func Test_Bloblang_transform_email(t *testing.T) {
	sliceBytes, err := json.Marshal(excludedDomains)
	require.NoError(t, err)

	sliceString := string(sliceBytes)

	mapping := fmt.Sprintf(`root = transform_email(value:%q,preserve_domain:true,preserve_length:true,excluded_domains:%v,max_length:%d,seed:1711240985047220000)`, email, sliceString, maxEmailCharLimit)
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	require.NoError(t, err)
	require.NotNil(t, res, "The response shouldn't be nil.")

	resStr, ok := res.(*string)
	require.True(t, ok)
	require.NotNil(t, resStr)
	require.Equal(t, len(*resStr), len(email), "Transformd email must be the same length as the input email")
	require.NotEqual(t, strings.Split(*resStr, "@")[1], "gmail.com", "The actual value should be have gmail.com as the domain")
}

// the case where the input value is null
func Test_TransformEmailTransformerWithEmptyValue(t *testing.T) {
	nilEmail := ""
	mapping := fmt.Sprintf(`root = transform_email(value:%q,preserve_domain:true,preserve_length:true,excluded_domains:%v,max_length:%d)`, nilEmail, []string{}, maxEmailCharLimit)
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the email transformer")

	_, err = ex.Query(nil)
	require.NoError(t, err)
}

func Test_TransformEmailTransformerWithEmptyValuePassNull(t *testing.T) {
	nilEmail := ""
	mapping := fmt.Sprintf(`root = transform_email(value:%q,preserve_domain:true,preserve_length:true,excluded_domains:null,max_length:%d)`, nilEmail, maxEmailCharLimit)
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the email transformer")

	_, err = ex.Query(nil)
	require.NoError(t, err)
}

func Test_TransformEmailTransformerWithEmptyValueNilDomains(t *testing.T) {
	nilEmail := "evis@gmail.com"

	mapping := fmt.Sprintf(`root = transform_email(value:%q,preserve_domain:true,preserve_length:true,excluded_domains:[],max_length:%d)`, nilEmail, maxEmailCharLimit)
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the email transformer")

	_, err = ex.Query(nil)
	require.NoError(t, err)
}

func Test_TransformEmailTransformerWithEmptyValueNilDomainsNoSliceDomains(t *testing.T) {
	nilEmail := "evis@gmail.com"

	mapping := fmt.Sprintf(`root = transform_email(value:%q,preserve_domain:true,preserve_length:true,excluded_domains:joiej,max_length:%d)`, nilEmail, maxEmailCharLimit)
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the email transformer")

	_, err = ex.Query(nil)
	require.NoError(t, err)
}

func Test_TransformEmailTransformerWithEmptyValueNilDomainsIntegerDomains(t *testing.T) {
	nilEmail := "evis@gmail.com"

	mapping := fmt.Sprintf(`root = transform_email(value:%q,preserve_domain:true,preserve_length:true,excluded_domains:132412,max_length:%d)`, nilEmail, maxEmailCharLimit)
	_, err := bloblang.Parse(mapping)
	require.Error(t, err, "The excluded domains must be strings")
}

func Test_TransformEmailTransformerWithEmptyValueNilDomainsIntegerSliceDomains(t *testing.T) {
	nilEmail := "evis@gmail.com"

	mapping := fmt.Sprintf(`root = transform_email(value:%q,preserve_domain:true,preserve_length:true,excluded_domains:[132,412],max_length:%d)`, nilEmail, maxEmailCharLimit)
	_, err := bloblang.Parse(mapping)
	require.Error(t, err, "The excluded domains must be strings")
}

func Test_TransformEmailTransformer_InvalidEmailArg(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_email(value:%q,invalid_email_action:"passthrough")`, "nick@neosync.dev")
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the email transformer")

	_, err = ex.Query(nil)
	require.NoError(t, err)
}

func Test_TransformEmailTransformer_NoOptions(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_email(value:%q)`, "nick@neosync.dev")
	ex, err := bloblang.Parse(mapping)
	require.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	require.NoError(t, err)
	require.NotEmpty(t, res)
}

func Test_fromAnyToStringSlice(t *testing.T) {
	var foo any = []any{"123", "456"}
	output, err := fromAnyToStringSlice(foo)
	require.NoError(t, err)
	require.Equal(t, []string{"123", "456"}, output)

	foo = []string{"123", "456"}
	output, err = fromAnyToStringSlice(foo)
	require.NoError(t, err)
	require.Equal(t, []string{"123", "456"}, output)

	foo = []any{"123", 456}
	output, err = fromAnyToStringSlice(foo)
	require.Error(t, err)
	require.Nil(t, output)

	foo = []int{123, 456}
	output, err = fromAnyToStringSlice(foo)
	require.Error(t, err)
	require.Nil(t, output)

	foo = nil
	output, err = fromAnyToStringSlice(foo)
	require.NoError(t, err)
	require.Empty(t, output)
}

func Test_ConverStringSliceToStringEmptySlice(t *testing.T) {
	slc := []string{}

	res, err := convertStringSliceToString(slc)
	require.NoError(t, err)
	require.Equal(t, "[]", res)
}

func Test_ConverStringSliceToStringNotEmptySlice(t *testing.T) {
	slc := []string{"gmail.com", "yahoo.com"}

	res, err := convertStringSliceToString(slc)
	require.NoError(t, err)
	require.Equal(t, `["gmail.com","yahoo.com"]`, res)
}
