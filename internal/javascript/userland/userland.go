package javascript_userland

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

// GetGenerateJavascriptFunction returns a Javascript function that takes no inputs and generates a value
// fnNameSuffix is the suffix of the function name
func GetGenerateJavascriptFunction(jsCode, fnNameSuffix string) string {
	return fmt.Sprintf(`
function fn_%s(){
  %s
};
`, sanitizeFunctionName(fnNameSuffix), jsCode)
}

// GetTransformJavascriptFunction returns a Javascript function that takes a value and input and returns a transformed value
// fnNameSuffix is the suffix of the function name
// includeRecord is true if the function should take in the input record
func GetTransformJavascriptFunction(jsCode, fnNameSuffix string, includeRecord bool) string {
	if includeRecord {
		return fmt.Sprintf(`
function fn_%s(value, input){
  %s
};
`, sanitizeFunctionName(fnNameSuffix), jsCode)
	}

	return fmt.Sprintf(`
function fn_%s(value){
  %s
};
`, sanitizeFunctionName(fnNameSuffix), jsCode)
}

func sanitizeFunctionName(input string) string {
	var result strings.Builder

	for i, r := range input {
		if unicode.IsLetter(r) || r == '_' || r == '$' || (unicode.IsDigit(r) && i > 0) {
			result.WriteRune(r)
		} else if unicode.IsDigit(r) && i == 0 {
			result.WriteRune('_')
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}

	return result.String()
}

// Takes a userland function and returns a single function that can be invoked by the JS VM
// Returns the property key that the output will be set to
func GetSingleGenerateFunction(userCode string) (code, propertyPath string) {
	propertyPath = uuid.NewString()
	fn := GetGenerateJavascriptFunction(userCode, propertyPath)
	outputSetter := BuildOutputSetter(propertyPath, false, false)
	return GetFunction([]string{fn}, []string{outputSetter}), propertyPath
}

// Takes a userland function and returns a single function that can be invoked by the JS VM
// Returns the property key that the output will be set to
func GetSingleTransformFunction(userCode string) (code, propertyPath string) {
	propertyPath = uuid.NewString()
	fn := GetTransformJavascriptFunction(userCode, propertyPath, false)
	outputSetter := BuildOutputSetter(propertyPath, true, false)
	return GetFunction([]string{fn}, []string{outputSetter}), propertyPath
}

// Takes all of the built userland functions and output setters and stuffs them into a single function that can be invoked by the JS VM
// Calling the resulting program expects benthos.v0_msg_as_structured() and neosync.patchStructuredMessage() to be defined in the JS VM
func GetFunction(jsFuncs, outputSetters []string) string {
	jsFunctionStrings := strings.Join(jsFuncs, "\n")

	benthosOutputString := strings.Join(outputSetters, "\n")

	jsCode := fmt.Sprintf(`
(() => {
%s
const input = benthos.v0_msg_as_structured();
const updatedValues = {}
%s
neosync.patchStructuredMessage(updatedValues)
})();`, jsFunctionStrings, benthosOutputString)
	return jsCode
}

// BuildOutputSetter builds a string that sets the output of the function to the property path on the "updatedValues" object
// includeInput is true if the propertyPath's value should be passed to the function
// includeInputRecord is true if the entire "input" object should be passed to the function as the second argument
func BuildOutputSetter(propertyPath string, includeInput, includeInputRecord bool) string {
	if includeInput {
		var strTemplate string
		if includeInputRecord {
			strTemplate = `updatedValues[%q] = fn_%s(%s, input)`
		} else {
			strTemplate = `updatedValues[%q] = fn_%s(%s)`
		}
		return fmt.Sprintf(
			strTemplate,
			propertyPath,
			sanitizeFunctionName(propertyPath),
			convertJsObjPathToOptionalChain(fmt.Sprintf("input.%s", propertyPath)),
		)
	}
	return fmt.Sprintf(
		`updatedValues[%q] = fn_%s()`,
		propertyPath,
		sanitizeFunctionName(propertyPath),
	)
}

func convertJsObjPathToOptionalChain(inputPath string) string {
	parts := strings.Split(inputPath, ".")
	for i := 1; i < len(parts); i++ {
		parts[i] = fmt.Sprintf("['%s']", parts[i])
	}
	return strings.Join(parts, "?.")
}
