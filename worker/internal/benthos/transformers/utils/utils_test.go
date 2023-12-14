package transformer_utils

import (
	"strconv"
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

func Test_GenerateRandomNumberWithBoundsMinError(t *testing.T) {

	_, err := GenerateRandomIntWithInclusiveBounds(10, 1)
	assert.Error(t, err, "Expected an error such that the min is greated than the max")
}

func Test_GenerateRandomNumberWithBoundsMinEqualMax(t *testing.T) {

	minMax := int64(5)
	val, err := GenerateRandomIntWithInclusiveBounds(minMax, minMax)
	assert.NoError(t, err, "Did not expect an error when min == max")
	assert.Equal(t, minMax, val, "actual value to be equal to min/max")

}

func Test_GenerateRandomNumberWithBoundsValid(t *testing.T) {

	min := int64(2)
	max := int64(9)
	val, err := GenerateRandomIntWithInclusiveBounds(min, max)
	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val >= min && val <= max, "actual value to be within the range")
}

func Test_GenerateRandomNumberWithBoundsNegative(t *testing.T) {

	min := int64(-2)
	max := int64(-9)

	val, err := GenerateRandomIntWithInclusiveBounds(min, max)
	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val <= min && val >= max, "actual value to be within the range")
}

func Test_GenerateRandomNumberWithBoundsNegativeToPositive(t *testing.T) {

	min := int64(-2)
	max := int64(9)

	val, err := GenerateRandomIntWithInclusiveBounds(min, max)
	assert.NoError(t, err, "Did not expect an error for valid range")
	assert.True(t, val >= min && val <= max, "actual value to be within the range")
}

func Test_GenerateRandomIntError(t *testing.T) {

	_, err := GenerateRandomInt(-2)
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

func Test_GenerateRandomInt(t *testing.T) {

	expectedLength := 9

	res, err := GenerateRandomInt(expectedLength)

	assert.NoError(t, err)
	numStr := strconv.FormatInt(int64(res), 10)
	assert.Equal(t, len(numStr), expectedLength, "The length of the generated random int should be the same as the expectedLength")

}

func Test_FirstDigitIsNineTrue(t *testing.T) {

	value := int64(9546789)

	res := FirstDigitIsNine(value)
	assert.Equal(t, res, true, "The first digit is nine.")
}

func Test_FirstDigitIsNineFalse(t *testing.T) {

	value := int64(23546789)

	res := FirstDigitIsNine(value)
	assert.Equal(t, res, false, "The first digit is not nine.")
}

func Test_GetInt64Legth(t *testing.T) {

	expected := 3

	val := GetInt64Length(782)

	assert.Equal(t, int64(expected), val, "The calculated length should match the expected length.")
}

func Test_GetFloat64Legth(t *testing.T) {

	expected := 8

	val := GetFloat64Length(7823.2332)

	assert.Equal(t, int64(expected), val, "The calculated length should match the expected length.")
}

func Test_IsLastDigitZeroTrue(t *testing.T) {

	value := int64(954670)

	res := IsLastDigitZero(value)
	assert.Equal(t, res, true, "The last digit is zero.")
}

func Test_IsLastDigitZeroFalse(t *testing.T) {

	value := int64(23546789)

	res := IsLastDigitZero(value)
	assert.Equal(t, res, false, "The last digit is not zero.")
}

func Test_RandomStringGeneration(t *testing.T) {

	expectedLength := 5
	res, err := GenerateRandomStringWithLength(int64(expectedLength))

	assert.NoError(t, err)
	assert.Equal(t, len(res), expectedLength, "The output string should be as long as the input string")

}

func Test_RandomStringGenerationError(t *testing.T) {

	_, err := GenerateRandomStringWithLength(int64(-2))
	assert.Error(t, err, "The length cannot be")

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

func Test_GetInt64Range(t *testing.T) {

	min := int64(2)
	max := int64(4)

	val, err := GetInt64Range(min, max)
	assert.NoError(t, err)

	assert.Equal(t, max-min, val)

}

func Test_GetInt64RangeError(t *testing.T) {

	min := int64(6)
	max := int64(2)

	_, err := GetInt64Range(min, max)
	assert.Error(t, err)

}

func Test_GetInt64RangeMinEqualMax(t *testing.T) {

	min := int64(2)
	max := int64(2)

	val, err := GetInt64Range(min, max)
	assert.NoError(t, err)

	assert.Equal(t, min, val)

}

func Test_GetFloat64Range(t *testing.T) {

	min := float64(2.2)
	max := float64(4.2)

	val, err := GetFloat64Range(min, max)
	assert.NoError(t, err)

	assert.Equal(t, max-min, val)

}

func Test_GetFloat64RangeError(t *testing.T) {

	min := float64(6.9)
	max := float64(2.2)

	_, err := GetFloat64Range(min, max)
	assert.Error(t, err)

}

func Test_GetFloat64RangeMinEqualMax(t *testing.T) {

	min := float64(2.2)
	max := float64(2.2)

	val, err := GetFloat64Range(min, max)
	assert.NoError(t, err)

	assert.Equal(t, min, val)
}

func Test_IsNegativeFloatTrue(t *testing.T) {

	val := IsNegativeFloat64(-1.63)

	assert.True(t, val, "The value should be negative")
}

func Test_IsNegativeFloatFalse(t *testing.T) {

	val := IsNegativeFloat64(324.435)

	assert.False(t, val, "The value should be positive")
}

func Test_AbsInt64Positive(t *testing.T) {

	val := int64(7)

	res := AbsInt64(val)
	assert.Equal(t, int64(7), res)

}

func Test_AbsInt64Negative(t *testing.T) {

	val := int64(-7)

	res := AbsInt64(val)
	assert.Equal(t, int64(7), res)

}

func Test_IsNegativeIntTrue(t *testing.T) {

	val := IsNegativeInt64(-1)

	assert.True(t, val, "The value should be negative")
}

func Test_IsNegativeIntFalse(t *testing.T) {

	val := IsNegativeInt64(1)

	assert.False(t, val, "The value should be positive")
}
