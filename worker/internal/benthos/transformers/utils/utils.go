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
func GenerateRandomIntWithInclusiveBounds(min, max int64) (int64, error) {

	if !IsNegativeInt64(min) && !IsNegativeInt64(max) && min > max {
		return 0, fmt.Errorf("the min can't be greater than the max if both the min and max are positive")
	}
	// If min is numerically larger (but less negative) than max, swap them
	if min < max {
		min, max = max, min
	}

	if min == max {
		return min, nil
	}

	// Calculate the range. Since we are dealing with negative numbers,
	// min is less negative than max.
	rangeVal := min - max + 1

	// Generate a random value within the range and subtract it from min
	// This keeps the result within the original [max, min] bounds
	return min - rand.Int63n(rangeVal), nil
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
func GetInt64Length(i int64) int64 {
	// Convert the int64 to a string
	str := strconv.FormatInt(i, 10)

	length := int64(len(str))

	return length
}

// GetFloatLength gets the number of digits in a float64
func GetFloat64Length(i float64) int64 {
	// Convert the float64 to a string with a specific format and precision
	// Using 'g' format and a precision of -1 to automatically determine the best format
	str := strconv.FormatFloat(i, 'g', -1, 64)

	// Remove the minus sign if the number is negative
	str = strings.Replace(str, "-", "", 1)

	// Remove the decimal point
	str = strings.Replace(str, ".", "", 1)

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

// Generates a random float64 in the range of the min and max float64 values
func GenerateRandomFloat64InRange(min, max float64) (float64, error) {

	if !IsNegativeFloat64(min) && !IsNegativeFloat64(max) && min > max {
		return 0, fmt.Errorf("the min can't be greater than the max if both the min and max are positive")
	}

	// generates a rand float64 value from [0.0,1.0)
	//nolint:all
	randValue := rand.Float64()

	// Scale and shift the value to the range
	returnValue := min + randValue*(max-min)

	return returnValue, nil
}

// Returns the float64 range between the min and max
func GetFloat64Range(min, max float64) (float64, error) {

	if min > max {
		return 0, fmt.Errorf("min cannot be greater than max")
	}

	if min == max {
		return min, nil
	}

	return max - min, nil

}

func IsNegativeFloat64(val float64) bool {
	if (val * -1) < 0 {
		return false
	} else {
		return true
	}
}

func IsNegativeInt64(val int64) bool {
	if (val * -1) < 0 {
		return false
	} else {
		return true
	}
}

func AbsInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// Returns the int64 range between the min and max
func GetInt64Range(min, max int64) (int64, error) {

	if min > max {
		return 0, fmt.Errorf("min cannot be greater than max")
	}

	if min == max {
		return min, nil
	}

	return max - min, nil

}
