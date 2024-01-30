package transformer_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomStringWithDefinedLength(t *testing.T) {

	val := int64(6)

	res, err := GenerateRandomStringWithDefinedLength(val)
	assert.NoError(t, err)

	assert.Equal(t, val, int64(len(res)), "The output string should be the same length as the input length")
}

func Test_GenerateRandomStringWithDefinedLengthError(t *testing.T) {

	val := int64(0)

	_, err := GenerateRandomStringWithDefinedLength(val)
	assert.Error(t, err)

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

func Test_GenerateRandomStringEqualMinMax(t *testing.T) {

	min := int64(4)
	max := int64(4)
	res, err := GenerateRandomStringWithInclusiveBounds(min, max)

	assert.NoError(t, err)
	assert.Equal(t, int64(len(res)), min, "The output string should be as long as the min or max since they're equal")

}

func Test_GenerateRandomStringRange(t *testing.T) {

	min := int64(2)
	max := int64(4)

	res, err := GenerateRandomStringWithInclusiveBounds(min, max)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, int64(len(res)), min, "the string should be greater than or equal to the min value")
	assert.LessOrEqual(t, int64(len(res)), max, "the string should be less than or equal to the max value")

}

func Test_GenerateRandomStringError(t *testing.T) {

	min := int64(-2)
	max := int64(4)

	_, err := GenerateRandomStringWithInclusiveBounds(min, max)
	assert.Error(t, err, "The min or max cannot be less than 0")
}

func Test_GenerateRandomStringErrorMinGreaterThanMax(t *testing.T) {

	min := int64(5)
	max := int64(4)

	_, err := GenerateRandomStringWithInclusiveBounds(min, max)
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

func Test_isValidCharTrue(t *testing.T) {

	val := "12wefg w1231"

	res := IsValidChar(val)

	assert.True(t, res)

}

func Test_isValidCharFalse(t *testing.T) {

	val := "ij諏計"

	res := IsValidChar(val)

	assert.False(t, res)

}

func Test_IsAllowedSpecialCharTrue(t *testing.T) {
	val := "$*#))"

	for _, r := range val {
		assert.True(t, IsAllowedSpecialChar(r), "Expected true for rune: %v", r)
	}
}

func Test_IsAllowedSpecialCharFalse(t *testing.T) {
	val := "諏計飯利"

	for _, r := range val {
		assert.False(t, IsAllowedSpecialChar(r), "Expected false for rune: %v", r)
	}
}
