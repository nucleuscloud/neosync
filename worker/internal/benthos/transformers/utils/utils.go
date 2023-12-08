package transformer_utils

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
)

var alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

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

// generates a random int between two numbers inclusive of the boundaries
func GenerateRandomIntWithInclusiveBounds(min, max int) (int, error) {

	if min > max {
		return 0, errors.New("min cannot be greater than max")
	}

	if min == max {
		return min, nil
	}

	// Generate a random number in the range [0, max-min]
	// the + 1 allows us to make the max inclusive
	//nolint:all
	randValue := rand.Intn(max - min + 1)

	// Shift the range to [min, max]
	return randValue + min, nil
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

// generates a random integer of length l that is passed in as a int64 param i.e. an l of 3 will generate
// an int64 of 3 digits such as 123 or 789.
func GenerateRandomInt(l int) (int, error) {
	if l <= 0 {
		return 0, errors.New("the length has to be greater than zero") // Or handle this case as an error
	}

	// Calculate the range
	min := int(math.Pow10(l - 1))
	max := int(math.Pow10(l)) - 1

	// Generate a random number in the range
	//nolint:all
	return rand.Intn(max-min+1) + min, nil
}

func FirstDigitIsNine(n int64) bool {
	// Convert the int64 to a string
	str := strconv.FormatInt(n, 10)

	// Check if the string is empty or if the first character is '9'
	if len(str) > 0 && str[0] == '9' {
		return true
	}

	return false
}

// gets the number of digits in an int64
func GetIntLength(i int64) int64 {
	// Convert the int64 to a string
	str := strconv.FormatInt(i, 10)

	length := int64(len(str))

	return length
}

func IsLastDigitZero(n int64) bool {
	// Convert the int64 to a string
	str := strconv.FormatInt(n, 10)

	// Check if the string is empty or if the last character is '0'
	if len(str) > 0 && str[len(str)-1] == '0' {
		return true
	}

	return false
}

// generate a random string of length l
func GenerateRandomStringWithLength(l int64) (string, error) {

	if l <= 0 {
		return "", fmt.Errorf("the length cannot be zero or negative")
	}

	result := make([]byte, l)

	for i := int64(0); i < l; i++ {
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
