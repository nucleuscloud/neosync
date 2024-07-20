package transformers

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

func toCamelCase(snake string) string {
	var result string
	upperNext := false
	for i, char := range snake {
		if char == '_' {
			upperNext = true
		} else {
			if upperNext {
				result += strings.ToUpper(string(char))
				upperNext = false
			} else {
				if i == 0 {
					result += strings.ToLower(string(char))
				} else {
					result += string(char)
				}
			}
		}
	}
	return result
}

type BenthosSpecParam struct {
	Name        string
	TypeStr     string
	IsOptional  bool
	HasDefault  bool
	Default     string
	Description string
}

type BenthosSpec struct {
	Name        string
	Description string
	Example     string
	Params      []*BenthosSpecParam
	Type        string
	SourceFile  string
}

type ParsedBenthosSpec struct {
	Params          []*BenthosSpecParam
	SpecDescription string
}

func ParseBloblangSpec(benthosSpec *BenthosSpec) (*ParsedBenthosSpec, error) {
	paramRegex := regexp.MustCompile(`bloblang\.New(\w+)Param\("(\w+)"\)(?:\.Optional\(\))?(?:\.Default\(([^()]*(?:\([^()]*\))?[^()]*)\))?(?:\.Description\("([^"]*)"\))?`)
	specDescriptionRegex := regexp.MustCompile(`\.Description\("([^"]*)"\)`)
	params := []*BenthosSpecParam{}
	readFile, err := os.Open(benthosSpec.SourceFile)
	if err != nil {
		return nil, err
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	var benthosSpecStr string
	start := false
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if strings.Contains(line, "bloblang.NewPluginSpec") {
			start = true
			benthosSpecStr += strings.TrimSpace(fileScanner.Text())
		} else if start {
			if strings.Contains(line, ":=") {
				break
			}
			benthosSpecStr += strings.TrimSpace(fileScanner.Text())
		}
	}

	var specDescription string
	parsedSpec := strings.Split(benthosSpecStr, ".Param")
	for _, line := range parsedSpec {
		if strings.Contains(line, "bloblang.NewPluginSpec()") {
			if specMatches := specDescriptionRegex.FindStringSubmatch(line); len(specMatches) > 0 {
				specDescription = specMatches[1]
			}
		}
		if strings.HasPrefix(line, "(") {
			matches := paramRegex.FindStringSubmatch(line)
			if len(matches) > 0 {
				defaultVal := matches[3]
				description := matches[4]
				// seed hack
				if strings.Contains(line, "Default(time.Now().UnixNano())") {
					defaultVal = "time.Now().UnixNano()"
					if specMatches := specDescriptionRegex.FindStringSubmatch(line); len(specMatches) > 0 {
						description = specMatches[1]
					}
				}
				param := &BenthosSpecParam{
					TypeStr:     lowercaseFirst(matches[1]),
					Name:        toCamelCase(matches[2]),
					IsOptional:  strings.Contains(line, ".Optional()"),
					HasDefault:  defaultVal != "",
					Default:     defaultVal,
					Description: description,
				}
				params = append(params, param)
			}
		}
	}

	return &ParsedBenthosSpec{
		Params:          params,
		SpecDescription: specDescription,
	}, nil
}

func lowercaseFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(string(s[0])) + s[1:]
}
