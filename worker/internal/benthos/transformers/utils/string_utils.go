package transformer_utils

import (
	"fmt"
	"math/rand"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

var allowedSpecialChars = map[rune]struct{}{
	'!': {}, '@': {}, '#': {}, '$': {}, '%': {}, '^': {}, '&': {}, '*': {}, '(': {}, ')': {},
	'-': {}, '+': {}, '=': {}, '_': {}, '[': {}, ']': {}, '{': {}, '}': {}, '|': {}, '\\': {},
	' ': {}, ';': {}, '"': {}, '<': {}, '>': {}, ',': {}, '.': {}, '/': {}, '?': {},
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

// Generate a random alphanumeric string of length l
func GenerateRandomStringWithDefinedLength(length int64) (string, error) {
	if length < 1 {
		return "", fmt.Errorf("the length of the string can't be less than 1")
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

// Generate a random alphanumeric string within the interval [min, max]
func GenerateRandomStringWithInclusiveBounds(min, max int64) (string, error) {
	if min < 0 || max < 0 || min > max {
		return "", fmt.Errorf("the min and max can't be less than 0 and the min can't be greater than the max")
	}

	var length int64

	if min == max {
		length = min
	} else {
		randlength, err := GenerateRandomInt64InValueRange(min, max)
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
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z0-9]{2,}$`
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

// use MaxASCII to ensure that the unicode value is only within the ASCII block which only contains latin numbers, letters and characters.
func IsValidChar(s string) bool {
	for _, r := range s {
		if !(r <= unicode.MaxASCII && (unicode.IsNumber(r) || unicode.IsLetter(r) || unicode.IsSpace(r) || IsAllowedSpecialChar(r))) {
			return false
		}
	}
	return true
}

func IsAllowedSpecialChar(r rune) bool {
	_, ok := allowedSpecialChars[r]
	return ok
}

// stringInSlice checks if a string is present in a slice of strings.
// It returns true if the string is found, and false otherwise.
func StringInSlice(str string, list []string) bool {
	for _, item := range list {
		if item == str {
			return true
		}
	}
	return false
}
