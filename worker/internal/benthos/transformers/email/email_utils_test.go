package transformers_email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseEmail(t *testing.T) {
	test := "evis@gmail.com"

	val, err := parseEmail(test)

	assert.NoError(t, err)
	assert.Equal(t, []string{"evis", "gmail.com"}, val, "Email should have a username and a domain name and tld as entries in the slice")
}

func Test_ParseEmailError(t *testing.T) {
	test := "ehiu.com"

	_, err := parseEmail(test)
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
