package transformer_utils

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"unicode"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
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
func GenerateRandomStringWithInclusiveBounds(minValue, maxValue int64) (string, error) {
	if minValue < 0 || maxValue < 0 || minValue > maxValue {
		return "", fmt.Errorf("invalid bounds when attempting to generate random string: [%d:%d]", minValue, maxValue)
	}

	var length int64
	if minValue == maxValue {
		length = minValue
	} else {
		randlength, err := GenerateRandomInt64InValueRange(minValue, maxValue)
		if err != nil {
			return "", fmt.Errorf("unable to generate a random length for the string within range [%d:%d]: %w", minValue, maxValue, err)
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

func GetRandomCharacterString(randomizer rng.Rand, size int64) string {
	var stringBuilder []rune = make([]rune, size)
	for i := int64(0); i < size; i++ {
		num := randomizer.Intn(26)
		stringBuilder[i] = rune('a' + num)
	}
	return string(stringBuilder)
}

// Generates a random string from the values corpus
// The lengthMap and mapKeys must be derivative of values.
// See the pre-generated values in the data-sets folder
// Eventually this will be abstracted into a Corpus struct for better readability.
// The expectation is the values, lengthMap, and mapKeys are all in their optimal, sorted form.
func GenerateStringFromCorpus(
	randomizer rng.Rand,
	values []string,
	lengthMap map[int64][2]int,
	mapKeys []int64,
	minLength *int64,
	maxLength int64,
	exclusions []string,
) (string, error) {
	excludedset := ToSet(exclusions)
	idxCandidates := ClampInts(mapKeys, minLength, &maxLength)
	if len(idxCandidates) == 0 {
		return "", fmt.Errorf("unable to find candidates with range %s", getRangeText(minLength, maxLength))
	}

	rangeIdxs := getRangeFromCandidates(idxCandidates, lengthMap)
	leftIdx := rangeIdxs[0]
	rightIdx := rangeIdxs[1]

	if leftIdx == -1 || rightIdx == -1 {
		return "", errors.New("unable to generate string from corpus due to invalid dictionary ranges")
	}

	attemptedValues := map[int64]struct{}{}
	totalValues := rightIdx - leftIdx + 1
	for int64(len(attemptedValues)) < totalValues {
		randIdx := randomInt64(randomizer, leftIdx, rightIdx)
		// may need to check to ensure randIdx is not outside of values bounds
		value := values[randIdx]
		if _, ok := attemptedValues[randIdx]; ok {
			continue
		}
		attemptedValues[randIdx] = struct{}{}
		if _, ok := excludedset[value]; ok {
			continue
		}
		return value, nil
	}
	return "", errors.New("unable to generate random value given the max length and excluded values")
}

func getRangeFromCandidates(candidates []int64, lengthMap map[int64][2]int) [2]int64 {
	candidateLen := len(candidates)
	if candidateLen == 0 {
		return [2]int64{-1, -1}
	}
	range1Idx := candidates[0]
	range2Idx := candidates[candidateLen-1]

	range1, ok1 := lengthMap[range1Idx]
	range2, ok2 := lengthMap[range2Idx]

	leftIdx := int64(-1)
	rightIdx := int64(-1)
	if ok1 {
		leftIdx = int64(range1[0])
		rightIdx = int64(range1[1])
	}
	if ok2 {
		leftIdx = int64(range2[0])
		rightIdx = int64(range2[1])
	}
	if ok1 && ok2 {
		leftIdx = int64(range1[0])
		rightIdx = int64(range2[1])
	}

	return [2]int64{leftIdx, rightIdx}
}

func getRangeText(minLength *int64, maxLength int64) string {
	if minLength != nil {
		return fmt.Sprintf("[%d:%d]", *minLength, maxLength)
	}
	return fmt.Sprintf("[-:%d]", maxLength)
}

func randomInt64(randomizer rng.Rand, minValue, maxValue int64) int64 {
	return minValue + randomizer.Int63n(maxValue-minValue+1)
}
