package transformers

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
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
	Name         string
	BloblangName string
	TypeStr      string
	IsOptional   bool
	HasDefault   bool
	Default      string
	Description  string
}

type BenthosSpec struct {
	Name             string
	BloblangFuncName string
	Description      string
	Example          string
	Category         string
	Params           []*BenthosSpecParam
	Type             string // transform or generate
	SourceFile       string
}

type ParsedBenthosSpec struct {
	Params           []*BenthosSpecParam
	BloblangFuncName string
	SpecDescription  string
	Category         string
}

func ExtractBenthosSpec(fileSet *token.FileSet) ([]*BenthosSpec, error) {
	transformerSpecs := []*BenthosSpec{}

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(path) == ".go" {
			node, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
			if err != nil {
				return fmt.Errorf("failed to parse file %s: %v", path, err)
			}
			for _, cgroup := range node.Comments {
				for _, comment := range cgroup.List {
					if strings.HasPrefix(comment.Text, "// +neosyncTransformerBuilder:") {
						parts := strings.Split(comment.Text, ":")
						if len(parts) < 3 {
							continue
						}
						transformerSpecs = append(transformerSpecs, &BenthosSpec{
							SourceFile: path,
							Name:       parts[2],
							Type:       parts[1],
						})
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("impossible to walk directories: %s", err)
	}
	return transformerSpecs, nil
}

func ParseBloblangSpec(benthosSpec *BenthosSpec) (*ParsedBenthosSpec, error) {
	paramRegex := regexp.MustCompile(
		`bloblang\.New(\w+)Param\("(\w+)"\)(?:\.Optional\(\))?(?:\.Default\(([^()]*(?:\([^()]*\))?[^()]*)\))?(?:\.Description\("([^"]*)"\))?`,
	)
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
	foundRegister := false
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if strings.Contains(line, "bloblang.NewPluginSpec") {
			start = true
			benthosSpecStr += strings.TrimSpace(line)
		} else if start {
			benthosSpecStr += strings.TrimSpace(line)
			if foundRegister {
				break // Now we break after capturing one more line after RegisterFunctionV2
			}
			if strings.Contains(line, "RegisterFunctionV2") {
				foundRegister = true
			}
		}
	}
	if !foundRegister {
		return nil, fmt.Errorf("RegisterFunctionV2 not found in file: %s", filepath.Base(benthosSpec.SourceFile))
	}

	categoryRegex := regexp.MustCompile(`\.Category\("([^"]*)"\)`)
	var category string
	if categoryMatches := categoryRegex.FindStringSubmatch(benthosSpecStr); len(
		categoryMatches,
	) > 0 {
		category = categoryMatches[1]
	}
	if category == "" {
		return nil, fmt.Errorf("category not found: %s", benthosSpec.SourceFile)
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
					if specMatches := specDescriptionRegex.FindStringSubmatch(line); len(
						specMatches,
					) > 0 {
						description = specMatches[1]
					}
				}
				param := &BenthosSpecParam{
					TypeStr:      lowercaseFirst(matches[1]),
					Name:         toCamelCase(matches[2]),
					BloblangName: matches[2],
					IsOptional:   strings.Contains(line, ".Optional()"),
					HasDefault:   defaultVal != "",
					Default:      defaultVal,
					Description:  description,
				}
				params = append(params, param)
			}
		}
	}

	bloblangFuncName, err := extractBloblangFunctionName(benthosSpecStr, benthosSpec.SourceFile)
	if err != nil {
		return nil, err
	}

	return &ParsedBenthosSpec{
		BloblangFuncName: bloblangFuncName,
		Params:           params,
		SpecDescription:  specDescription,
		Category:         category,
	}, nil
}

func extractBloblangFunctionName(input, sourceFile string) (string, error) {
	// Looks for bloblang.RegisterFunctionV2 and captures the function name in quotes
	re := regexp.MustCompile(`RegisterFunctionV2\s*\(\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(input)

	if len(matches) == 0 {
		return "", fmt.Errorf("bloblang function name not found: %s", filepath.Base(sourceFile))
	}
	return matches[1], nil
}

func lowercaseFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(string(s[0])) + s[1:]
}

func CapitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}
