package transformer_utils

import (
	"errors"
	"fmt"
	"math/rand"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
)

var alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

/* SLICE MANIPULATION UTILS */

// returns a random index from a one-dimensional slice
func GetRandomValueFromSlice[T any](arr []T) (T, error) {
	if len(arr) == 0 {
		var zeroValue T
		return zeroValue, errors.New("slice is empty")
	}

	//nolint:all
	randomIndex := rand.Intn(len(arr))

	return arr[randomIndex], nil
}

// substrings a string using rune length to account for multi-byte characters
func SliceString(s string, l int) string {

	// use runes instead of strings in order to avoid slicing a multi-byte character and returning invalid UTF-8
	runes := []rune(s)

	if l > len(runes) {
		l = len(runes)
	}

	return string(runes[:l])
}

// converts a slice of int to a slice of strings
func IntSliceToStringSlice(ints []int64) []string {

	var str []string

	if len(ints) == 0 {
		return []string{}
	}

	for i := range ints {
		str = append(str, strconv.Itoa((i)))

	}

	return str
}

// generate a random string of length l
func GenerateRandomString(min, max int64) (string, error) {

	if min < 0 && max < 0 && min > max {
		return "", fmt.Errorf("the min and max can't be less than 0 and the min can't be greater than the max")
	}

	var length int64

	if min == max {
		length = min
	} else {

		randlength, err := GenerateRandomInt64WithInclusiveBounds(min, max)
		if err != nil {
			return "", fmt.Errorf("unable to generate a random length for the string")
		}

		length = randlength
	}

	result := make([]byte, length)

	for i := int64(0); i < length; i++ {
		// Generate a random index in the range [0, len(alphabet))
		//nolint:all
		index := rand.Intn(len(alphanumeric))

		// Get the character at the generated index and append it to the result
		result[i] = alphanumeric[index]
	}

	return strings.ToLower(string(result)), nil
}

func ParseEmail(email string) ([]string, error) {

	inputEmail, err := mail.ParseAddress(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email format: %s", email)
	}

	parsedEmail := strings.Split(inputEmail.Address, "@")

	return parsedEmail, nil
}

func IsValidEmail(email string) bool {
	// Regular expression pattern for a simple email validation
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(emailPattern)
	return regex.MatchString(email)
}

func IsValidDomain(domain string) bool {
	pattern := `^@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	rfcRegex, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	return rfcRegex.MatchString(domain)
}

func IsValidUsername(username string) bool {

	// Regex to match RFC 5322 username
	// Chars allowed: a-z A-Z 0-9 . - _
	// First char must be alphanumeric
	// Last char must be alphanumeric or numeric
	// 63 max chars
	rfcRegex := `^[A-Za-z0-9](?:[A-Za-z0-9!#$%&'*+-/=?^_` +
		`{|}~.]{0,62}[A-Za-z0-9])?$`

	matched, _ := regexp.MatchString(rfcRegex, username)

	return matched
}
