package transformer_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// returns a random index from a one-dimensional slice
func Test_GetRandomValueFromSliceEmptySlice(t *testing.T) {

	arr := []string{}
	_, err := GetRandomValueFromSlice(arr)
	assert.Error(t, err, "Expected an error for the empty slice")

}

func Test_GetRandomValueFromSliceNonEmptySlice(t *testing.T) {

	arr := []string{"a", "b", "c"}
	res, err := GetRandomValueFromSlice(arr)
	assert.NoError(t, err)
	assert.Contains(t, arr, res, "Expected the response to be included in the input array")

}

func Test_SliceStringEmptyString(t *testing.T) {

	res := SliceString("", 10)
	assert.Empty(t, res, "Expected result to be an empty string")
}

func Test_SliceStringShortString(t *testing.T) {

	s := "short"
	res := SliceString(s, 10)
	assert.Equal(t, s, res, "Expected result to be equal to the input string")

}

func Test_SliceStringValidSlice(t *testing.T) {

	s := "hello, world"
	length := 5
	expected := "hello"
	res := SliceString(s, length)
	assert.Equal(t, expected, res, "Expected result to be a substring of the input string with the specified length")
}

func Test_IntArryToStringArr(t *testing.T) {

	val := []int64{1, 2, 3, 4}

	res := IntSliceToStringSlice(val)

	assert.IsType(t, res, []string{})
	assert.Equal(t, len(res), len(val), "The slices should be the same length")

}

func Test_IntArryToStringArrEmptySlice(t *testing.T) {

	val := []int64{}

	res := IntSliceToStringSlice(val)

	assert.IsType(t, res, []string{})
	assert.Equal(t, len(res), len(val), "The slices should be the same length")
}

func Test_GenerateRandomStringEqualMinMax(t *testing.T) {

	min := int64(4)
	max := int64(4)
	res, err := GenerateRandomString(min, max)

	assert.NoError(t, err)
	assert.Equal(t, len(res), min, "The output string should be as long as the min or max since they're equal")

}

func Test_GenerateRandomStringRange(t *testing.T) {

	min := int64(2)
	max := int64(4)

	res, err := GenerateRandomString(min, max)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(res), min, "the string should be greater than or equal to the min value")
	assert.LessOrEqual(t, len(res), max, "the string should be less than or equal to the max value")

}

func Test_GenerateRandomStringError(t *testing.T) {

	min := int64(-2)
	max := int64(4)

	_, err := GenerateRandomString(min, max)
	assert.Error(t, err, "The min or max cannot be less than 0")

}

func Test_GenerateRandomStringErrorMinGreaterThanMax(t *testing.T) {

	min := int64(5)
	max := int64(4)

	_, err := GenerateRandomString(min, max)
	assert.Error(t, err, "The min cannot be greater than the max")

}

func Test_ParseEmail(t *testing.T) {
	test := "evis@gmail.com"

	val, err := ParseEmail(test)

	assert.NoError(t, err)
	assert.Equal(t, []string{"evis", "gmail.com"}, val, "Email should have a username and a domain name and tld as entries in the slice")
}

func Test_ParseEmailError(t *testing.T) {
	test := "ehiu.com"

	_, err := ParseEmail(test)
	assert.Error(t, err, "Email doesn't have a valid email format")
}

func Test_IsValidEmail(t *testing.T) {

	assert.True(t, IsValidEmail("test@example.com"), "Email follows the valid email format")
	assert.False(t, IsValidEmail("invalid"), "Email doesn't have a valid email format")

}

func Test_IsValidDomain(t *testing.T) {

	assert.True(t, IsValidDomain("@example.com"), "Domain should have an @ sign and then a domain and top level domain")
	assert.False(t, IsValidDomain("invalid"), "Domain doesn't contain an @ sign or a top level domain")
}

func Test_IsValidUsername(t *testing.T) {

	assert.True(t, IsValidUsername("test"), "Username should be an alphanumeric value comprised of  a-z A-Z 0-9 . - _ and starting and ending in alphanumeric chars with a max length of 63")
	assert.True(t, IsValidUsername("test-test"), "Username should be an alphanumeric value comprised of  a-z A-Z 0-9 . - _ and starting and ending in alphanumeric chars with a max length of 63")
	assert.True(t, IsValidUsername("test-TEST"), "Username should be an alphanumeric value comprised of  a-z A-Z 0-9 . - _ and starting and ending in alphanumeric chars with a max length of 63")
	assert.False(t, IsValidUsername("eger?45//"), "Username contains non-alphanumeric characters")
}
