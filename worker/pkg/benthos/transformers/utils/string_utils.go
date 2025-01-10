package transformer_utils

import (
	"errors"
	"fmt"
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
	SpecialChars = setToSlice(SpecialCharsSet)
}

func setToSlice[T rune | string](input map[T]struct{}) []T {
	slice := make([]T, 0, len(input))
	for val := range input {
		slice = append(slice, val)
	}
	return slice
}

// Generate a random alphanumeric string of length l
func GenerateRandomStringWithDefinedLength(randomizer rng.Rand, length int64) (string, error) {
	if length < 1 {
		return "", fmt.Errorf("the length of the string can't be less than 1")
	}

	result := make([]byte, length)
	for i := int64(0); i < length; i++ {
		// Generate a random index in the range [0, len(alphabet))
		index := randomizer.Intn(len(alphanumeric))
		// Get the character at the generated index and append it to the result
		result[i] = alphanumeric[index]
	}
	return strings.ToLower(string(result)), nil
}

// Generate a random alphanumeric string within the interval [min, max]
func GenerateRandomStringWithInclusiveBounds(randomizer rng.Rand, minValue, maxValue int64) (string, error) {
	if minValue < 0 || maxValue < 0 || minValue > maxValue {
		return "", fmt.Errorf("invalid bounds when attempting to generate random string: [%d:%d]", minValue, maxValue)
	}

	var length int64
	if minValue == maxValue {
		length = minValue
	} else {
		randlength, err := GenerateRandomInt64InValueRange(randomizer, minValue, maxValue)
		if err != nil {
			return "", fmt.Errorf("unable to generate a random length for the string within range [%d:%d]: %w", minValue, maxValue, err)
		}
		length = randlength
	}

	result := make([]byte, length)
	for i := int64(0); i < length; i++ {
		// Generate a random index in the range [0, len(alphabet))
		index := randomizer.Intn(len(alphanumeric))
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

func TrimStringIfExceeds(s string, maxLength int64) string {
	if int64(len(s)) > maxLength {
		return s[:maxLength]
	}
	return s
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
// TODO: there may be some optimizations we can do here to reduce the number of attempts to generate a string by currying away the maxLength and exclusions
// If we know these ahead of time, we can pre-compute the actual candidates and then just randomly select from the candidates.
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
