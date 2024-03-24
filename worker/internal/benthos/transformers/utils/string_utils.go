package transformer_utils

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"unicode"
)

var SpecialCharsSet = map[rune]struct{}{
	'!': {}, '@': {}, '#': {}, '$': {}, '%': {}, '^': {}, '&': {}, '*': {}, '(': {}, ')': {},
	'-': {}, '+': {}, '=': {}, '_': {}, '[': {}, ']': {}, '{': {}, '}': {}, '|': {}, '\\': {},
	' ': {}, ';': {}, '"': {}, '<': {}, '>': {}, ',': {}, '.': {}, '/': {}, '?': {},
}
var SpecialChars []rune

func init() {
	SpecialChars = SetToSlice(SpecialCharsSet)
}

func SetToSlice[T rune | string](input map[T]struct{}) []T {
	slice := make([]T, 0, len(input))
	for val := range input {
		slice = append(slice, val)
	}
	return slice
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
		return "", fmt.Errorf("invalid bounds when attempting to generate random string: [%d:%d]", min, max)
	}

	var length int64
	if min == max {
		length = min
	} else {
		randlength, err := GenerateRandomInt64InValueRange(min, max)
		if err != nil {
			return "", fmt.Errorf("unable to generate a random length for the string within range [%d:%d]: %w", min, max, err)
		}
		length = randlength
	}

	result := make([]byte, length)
	for i := int64(0); i < length; i++ {
		// Generate a random index in the range [0, len(alphabet))
		index := rand.Intn(len(alphanumeric)) //nolint:gosec
		// Get the character at the generated index and append it to the result
		result[i] = alphanumeric[index]
	}
	return strings.ToLower(string(result)), nil
}

const (
	emailRegexPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z0-9]{2,}$`
)

var (
	emailRegex = regexp.MustCompile(emailRegexPattern)
)

func IsValidEmail(email string) bool {
	// Regular expression pattern for a simple email validation
	return emailRegex.MatchString(email)
}

const (
	domainPattern = `^@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
)

var (
	domainRegex = regexp.MustCompile(domainPattern)
)

func IsValidDomain(domain string) bool {
	return domainRegex.MatchString(domain)
}

const (
	// Regex to match RFC 5322 username
	// Chars allowed: a-z A-Z 0-9 . - _
	// First char must be alphanumeric
	// Last char must be alphanumeric or numeric
	// 63 max char
	usernamePattern = `^[A-Za-z0-9](?:[A-Za-z0-9!#$%&'*+\/=?^_{|}~.-]{0,62}[A-Za-z0-9])?$`
)

var (
	usernameRegex = regexp.MustCompile(usernamePattern)
)

func IsValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
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
	_, ok := SpecialCharsSet[r]
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

func TrimStringIfExceeds(s string, maxLength int64) string {
	if int64(len(s)) > maxLength {
		return s[:maxLength]
	}
	return s
}

func GetSmallerOrEqualNumbers(nums []int64, val int64) []int64 {
	candidates := []int64{}

	for _, num := range nums {
		if num <= val {
			candidates = append(candidates, num)
		}
	}
	return candidates
}

func ToSet[T string | int64](input []T) map[T]struct{} {
	unique := map[T]struct{}{}

	for _, val := range input {
		unique[val] = struct{}{}
	}

	return unique
}

func WithoutCharacters(input string, invalidChars []rune) string {
	invalid := make(map[rune]bool)
	for _, ch := range invalidChars {
		invalid[ch] = true
	}

	var builder strings.Builder
	for _, ch := range input {
		if !invalid[ch] {
			builder.WriteRune(ch)
		}
	}

	// Return the cleaned string.
	return builder.String()
}

func GetRandomCharacterString(randomizer *rand.Rand, size int64) string {
	var stringBuilder []rune = make([]rune, size)
	for i := int64(0); i < size; i++ {
		num := randomizer.Intn(26)
		stringBuilder[i] = rune('a' + num)
	}
	return string(stringBuilder)
}

// For the given map and list of keys along with bounds, will generate a random value from the corpus
// stringMap is expected to be in the format of key: size, value: values of size
// sizeIndices is expected to be a slice of the stringMap, preferably sorted
func GenerateStringFromCorpus(
	randomizer *rand.Rand,
	stringMap map[int64][]string,
	sizeIndices []int64,
	minLength *int64,
	maxLength int64,
) (string, error) {
	idxCandidates := ClampInts(sizeIndices, minLength, &maxLength)
	if len(idxCandidates) == 0 {
		return "", fmt.Errorf("unable to find first name with range %s", getRangeText(minLength, maxLength))
	}
	mapKey := idxCandidates[randomizer.Intn(len(idxCandidates))]
	values, ok := stringMap[mapKey]
	if !ok {
		return "", fmt.Errorf("when generating string from corpus, the generated index was not present in map: %d", mapKey)
	}
	return values[randomizer.Intn(len(values))], nil
}

func getRangeText(minLength *int64, maxLength int64) string {
	if minLength != nil {
		return fmt.Sprintf("[%d:%d]", *minLength, maxLength)
	}
	return fmt.Sprintf("[-:%d]", maxLength)
}
