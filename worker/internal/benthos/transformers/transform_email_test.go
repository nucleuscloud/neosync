package transformers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/stretchr/testify/assert"
)

var email = "evis@gmail.com"
var maxEmailCharLimit = int64(40)
var exclusionList = []string{"gmail.com", "hotmail.com"}
var emptyExclusionList = []string{}

func Test_TransformEmailPreserveLengthFalsePreserveDomainTrue(t *testing.T) {

	res, err := TransformEmail(email, false, true, maxEmailCharLimit, emptyExclusionList)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")
	assert.Equal(t, "gmail.com", strings.Split(*res, "@")[1])
	pEmail, _ := transformer_utils.ParseEmail(email)
	assert.Equal(t, pEmail[1], strings.Split(*res, "@")[1], "The domains should be the same because preserveDomain is true and the exclusion list is empty")

}

func Test_TransformEmailPreserveLengthFalsePreserveDomainFalse(t *testing.T) {

	res, err := TransformEmail(email, false, false, maxEmailCharLimit, emptyExclusionList)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")

}

func Test_TransformEmailPreserveLengthTruePreserveDomainFalse(t *testing.T) {

	res, err := TransformEmail(email, true, false, maxEmailCharLimit, emptyExclusionList)

	assert.NoError(t, err)
	assert.Equal(t, len(email), len(*res), "The expected email should be have a valid email structure")
}

func Test_TransformEmail(t *testing.T) {

	res, err := TransformEmailPreserveDomain(email, true, maxEmailCharLimit, emptyExclusionList)

	assert.NoError(t, err)
	/* There is a very small chance that the randomly generated email address actually matches
	the input email address which is why can't do an assert.NoEqual() but instead just have to check
	that the email has the correct structrue */
	assert.Equal(t, true, transformer_utils.IsValidEmail(res), "true", "The domain should not explicitly be preserved but randomly generated.")
}

func Test_TransformmailPreserveDomain(t *testing.T) {

	res, err := TransformEmailPreserveDomain(email, true, maxEmailCharLimit, emptyExclusionList)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(res), "true", "The domain should not explicitly be preserved but randomly generated.")
}

func Test_TransformEmailPreserveLength(t *testing.T) {

	res, err := TransformEmailPreserveLength(email, emptyExclusionList)

	assert.NoError(t, err)
	assert.Equal(t, len(email), len(res), "The length of the emails should be the same")
}

func Test_TransformEmailPreserveLengthTruePreserveDomainTrue(t *testing.T) {

	res, err := TransformEmailPreserveDomainAndLength(email, maxEmailCharLimit, emptyExclusionList)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(res), "The expected email should be have a valid email structure")

}

func Test_TransformEmailUsername(t *testing.T) {

	res, err := GenerateUsername(int64(13))
	assert.NoError(t, err)

	assert.Equal(t, true, transformer_utils.IsValidUsername(res), "The expected email should have a valid username")

}

func Test_TransformEmailPreserveDomainTrueExclusionListTrue(t *testing.T) {

	res, err := TransformEmailPreserveDomain(email, true, 40, exclusionList)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(res), "The expected email should be have a valid email structure")
	pEmail, _ := transformer_utils.ParseEmail(email)
	assert.NotEqual(t, pEmail[1], strings.Split(res, "@")[1], "The domains should be different")

}

func Test_TransformEmailPreserveDomainTrueExclusionListEmpty(t *testing.T) {

	elEmpty := []string{}

	res, err := TransformEmailPreserveDomain(email, true, 40, elEmpty)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(res), "The expected email should be have a valid email structure")
	pEmail, _ := transformer_utils.ParseEmail(email)
	assert.Equal(t, pEmail[1], strings.Split(res, "@")[1], "The domains should be the same")

}

func Test_TransformEmailPreserveDomainFalsePreserveLengthTrueExclusionListTrue(t *testing.T) {

	res, err := TransformEmail(email, true, false, 40, exclusionList)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")
	pEmail, _ := transformer_utils.ParseEmail(email)
	assert.Equal(t, pEmail[1], strings.Split(*res, "@")[1], "The domains should be the same")
	assert.Equal(t, len(email), len(*res), "The emails should have the same length")
}

func Test_TransformEmailPreserveDomainFalsePreserveLengthTrueExclusionListEmpty(t *testing.T) {

	elEmpty := []string{}

	res, err := TransformEmail(email, true, false, 40, elEmpty)

	fmt.Println("res", *res)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")
	pEmail, _ := transformer_utils.ParseEmail(email)
	assert.NotEqual(t, pEmail[1], strings.Split(*res, "@")[1], "The domains should be different")
	assert.Equal(t, len(email), len(*res), "The emails should have the same length")

}

func Test_TransformEmailPreserveDomainTruePreserveLengthTrueExclusionList(t *testing.T) {

	res, err := TransformEmail(email, true, true, 40, exclusionList)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")
	pEmail, _ := transformer_utils.ParseEmail(email)
	assert.NotEqual(t, pEmail[1], strings.Split(*res, "@")[1], "The domains should be different")
	assert.Equal(t, len(email), len(*res), "The emails should have the same length")

}

func Test_TransformEmailPreserveDomainTruePreserveLengthTrueExclusionListEmpty(t *testing.T) {

	elEmpty := []string{}

	res, err := TransformEmail(email, true, true, 40, elEmpty)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")
	pEmail, _ := transformer_utils.ParseEmail(email)
	assert.Equal(t, pEmail[1], strings.Split(*res, "@")[1], "The domains should be different")
	assert.Equal(t, len(email), len(*res), "The emails should have the same length")
}

func Test_TransformEmailPreserveDomainFalsePreserveLengthFalseExclusionListTrue(t *testing.T) {

	res, err := TransformEmail(email, false, false, 40, exclusionList)

	fmt.Println("res", *res)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")
	pEmail, _ := transformer_utils.ParseEmail(email)
	assert.Equal(t, pEmail[1], strings.Split(*res, "@")[1], "The domains should be the same")

}

func Test_TransformEmailPreserveDomainFalsePreserveLengthFalseExclusionListEmpty(t *testing.T) {

	elEmpty := []string{}

	res, err := TransformEmail(email, false, false, 40, elEmpty)

	fmt.Println("res", *res)

	assert.NoError(t, err)
	assert.Equal(t, true, transformer_utils.IsValidEmail(*res), "The expected email should be have a valid email structure")
	pEmail, _ := transformer_utils.ParseEmail(email)
	assert.NotEqual(t, pEmail[1], strings.Split(*res, "@")[1], "The domains should be different")

}

func Test_TransformEmailTransformerWithValue(t *testing.T) {
	sliceBytes, err := json.Marshal(exclusionList)
	if err != nil {
		log.Fatalf("json.Marshal failed: %v", err)
	}

	sliceString := string(sliceBytes)

	mapping := fmt.Sprintf(`root = transform_email(email:%q,preserve_domain:true,preserve_length:true,exclusion_list:%v,max_length:%d)`, email, sliceString, maxEmailCharLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resStr, ok := res.(*string)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resStr != nil {
		assert.Equal(t, len(*resStr), len(email), "Transformd email must be the same length as the input email")
		assert.NotEqual(t, strings.Split(*resStr, "@")[1], "gmail.com", "The actual value should be have gmail.com as the domain")
	} else {
		t.Error("Pointer is nil, expected a valid string pointer")
	}
}

// the case where the input value is null
func Test_TransformEmailTransformerWithEmptyValue(t *testing.T) {

	nilEmail := ""
	mapping := fmt.Sprintf(`root = transform_email(email:%q,preserve_domain:true,preserve_length:true,exclusion_list: %s,max_length:%d)`, nilEmail, emptyExclusionList, maxEmailCharLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the email transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}
