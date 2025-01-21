package javascript_userland

import (
	"fmt"
	"strings"
	"unicode"
)

// GetGenerateJavascriptFunction returns a Javascript function that takes no inputs and generates a value
func GetGenerateJavascriptFunction(jsCode, fnNameSuffix string) string {
	return fmt.Sprintf(`
function fn_%s(){
  %s
};
`, sanitizeFunctionName(fnNameSuffix), jsCode)
}

// GetTransformJavascriptFunction returns a Javascript function that takes a value and input and returns a transformed value
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

// Takes all of the built userland functions and output setters and stuffs them into a single function that can be invoked by the JS VM
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

func BuildOutputSetter(propertyPath string, includeInput bool) string {
	if includeInput {
		return fmt.Sprintf(
			`updatedValues[%q] = fn_%s(%s, input)`,
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
