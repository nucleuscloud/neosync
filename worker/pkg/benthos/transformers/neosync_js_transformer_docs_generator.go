//go:build ignore

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ParamInfo struct {
	Name        string
	TypeStr     string
	IsOptional  bool
	HasDefault  bool
	Default     string
	Description string
}

type FuncInfo struct {
	Name        string
	Description string
	Params      []*ParamInfo
	Type        string
	SourceFile  string
}

func main() {
	args := os.Args
	if len(args) < 1 {
		panic("must provide necessary args")
	}

	// packageName := args[1]
	fileSet := token.NewFileSet()
	transformerFuncs := []*FuncInfo{}

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(path) == ".go" {
			node, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
			if err != nil {
				log.Printf("Failed to parse file %s: %v", path, err)
				return nil
			}
			for _, cgroup := range node.Comments {
				for _, comment := range cgroup.List {
					if strings.HasPrefix(comment.Text, "// +neosyncTransformerBuilder:") {
						parts := strings.Split(comment.Text, ":")
						if len(parts) < 3 {
							continue
						}
						transformerFuncs = append(transformerFuncs, &FuncInfo{
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
		log.Fatalf("impossible to walk directories: %s", err)
	}

	for _, tf := range transformerFuncs {
		p, err := parseBloblangSpec(tf)
		if err != nil {
			fmt.Println("Error parsing bloblang params:", err)
		}
		tf.Params = p
		if tf.Name == "generateUUID" {
			jsonF, _ := json.MarshalIndent(tf, "", " ")
			fmt.Printf("%s \n", string(jsonF))

		}
	}

	// for _, tf := range transformerFuncs {
	// 	codeStr, err := generateCode(packageName, tf)
	// 	if err != nil {
	// 		fmt.Println("Error writing to output file:", err)
	// 		return
	// 	}
	// 	output := fmt.Sprintf("gen_%s", tf.SourceFile)
	// 	outputFile, err := os.Create(output)
	// 	if err != nil {
	// 		fmt.Println("Error creating output file:", err)
	// 		return
	// 	}

	// 	_, err = outputFile.WriteString(codeStr)
	// 	if err != nil {
	// 		fmt.Println("Error writing to output file:", err)
	// 		return
	// 	}
	// 	outputFile.Close()
	// }
}

// func parseBloblangSpec(funcInfo *FuncInfo) ([]*ParamInfo, error) {
// 	paramRegex := regexp.MustCompile(`bloblang\.New(\w+)Param\("(\w+)"\)(?:\.Optional\(\))?(?:\.Default\(([^()]*(?:\([^()]*\))?[^()]*)\))?`)
// 	params := []*ParamInfo{}
// 	readFile, err := os.Open(funcInfo.SourceFile)
// 	if err != nil {
// 		return nil, err
// 	}
// 	fileScanner := bufio.NewScanner(readFile)
// 	fileScanner.Split(bufio.ScanLines)

// 	for fileScanner.Scan() {
// 		line := fileScanner.Text()
// 		line = strings.TrimSpace(line)
// 		// parse bloblang spec
// 		if strings.HasPrefix(line, "Param(bloblang.") {
// 			matches := paramRegex.FindStringSubmatch(line)
// 			if len(matches) > 0 {
// 				param := &ParamInfo{
// 					TypeStr:    lowercaseFirst(matches[1]),
// 					Name:       toCamelCase(matches[2]),
// 					IsOptional: strings.Contains(line, ".Optional()"),
// 					HasDefault: matches[3] != "",
// 					Default:    matches[3],
// 				}
// 				params = append(params, param)
// 			}
// 		}
// 	}
// 	readFile.Close()
// 	return params, nil
// }

func parseBloblangSpec(funcInfo *FuncInfo) ([]*ParamInfo, error) {
	paramRegex := regexp.MustCompile(`bloblang\.New(\w+)Param\("(\w+)"\)(?:\.Optional\(\))?(?:\.Default\(([^()]*(?:\([^()]*\))?[^()]*)\))?(?:\.Description\("([^"]*)"\))?`)
	specDescriptionRegex := regexp.MustCompile(`\.Description\("([^"]*)"\)`)
	params := []*ParamInfo{}
	readFile, err := os.Open(funcInfo.SourceFile)
	if err != nil {
		return nil, err
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	var benthosSpec string
	start := false
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if strings.Contains(line, "bloblang.NewPluginSpec") {
			start = true
			benthosSpec += strings.TrimSpace(fileScanner.Text())
		} else if start {
			if strings.Contains(line, ":=") {
				break
			}
			benthosSpec += strings.TrimSpace(fileScanner.Text())
		}
	}

	parsedSpec := strings.Split(benthosSpec, ".Param")
	for _, line := range parsedSpec {
		if strings.Contains(line, "bloblang.NewPluginSpec()") {
			if specMatches := specDescriptionRegex.FindStringSubmatch(line); len(specMatches) > 0 {
				funcInfo.Description = specMatches[1]
			}
		}
		// parse param level description
		if strings.HasPrefix(line, "(") {
			matches := paramRegex.FindStringSubmatch(line)
			if len(matches) > 0 {
				param := &ParamInfo{
					TypeStr:     lowercaseFirst(matches[1]),
					Name:        toCamelCase(matches[2]),
					IsOptional:  strings.Contains(line, ".Optional()"),
					HasDefault:  matches[3] != "",
					Default:     matches[3],
					Description: matches[4],
				}
				params = append(params, param)
			}
		}
	}

	return params, nil
}

// const codeTemplate = `
// // Code generated by Neosync neosync_transformer_generator.go. DO NOT EDIT.
// // source: {{.SourceFile}}

// package {{.PackageName}}

// import (
// 	{{- if eq .ImportFmt true}}
// 	"fmt"
// 	{{ end }}
// 	{{- if eq .HasSeedParam true}}
// 	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
// 	"github.com/nucleuscloud/neosync/worker/pkg/rng"
// 	{{ end }}
// )

// type {{.StructName}} struct{}

// type {{.StructName}}Opts struct {
// 	{{- if eq .HasSeedParam true}}
// 	randomizer     rng.Rand
// 	{{ end }}
// 	{{- range $index, $param := .FunctInfo.Params }}
// 	{{- if eq $param.Name "value" }}{{ continue }}{{ end }}
// 	{{- if eq $param.Name "seed" }}{{ continue }}{{ end }}
// 	{{- if $param.IsOptional }}
// 	{{$param.Name}} *{{$param.TypeStr}}
// 	{{- else }}
// 	{{$param.Name}} {{$param.TypeStr}}
// 	{{- end }}
// 	{{- end }}
// }

// func New{{.StructName}}() *{{.StructName}} {
// 	return &{{.StructName}}{}
// }

// func (t *{{.StructName}}) GetJsTemplateData() (*TemplateData, error) {
// 	return &TemplateData{
// 		Name: "{{.FunctInfo.Name}}",
// 		Description: "{{.FunctInfo.Description}}",
// 	}, nil
// }

// func (t *{{.StructName}}) ParseOptions(opts map[string]any) (any, error) {
// 	transformerOpts := &{{.StructName}}Opts{}
// 	{{- range $index, $param := .FunctInfo.Params }}
// 	{{- if eq $param.Name "value" }}{{ continue }}{{ end }}

// 	{{- if eq $param.Name "seed" }}

// 	var seed int64
// 	seedArg, ok := opts["seed"].(int64)
// 	if ok {
// 		seed = seedArg
// 	} else {
// 		var err error
// 		seed, err = transformer_utils.GenerateCryptoSeed()
// 		if err != nil {
// 			return nil, fmt.Errorf("unable to generate seed: %w", err)
// 		}
// 	}
// 	transformerOpts.randomizer = rng.New(seed)

// 	{{- continue }}
// 	{{ end }}
// 	{{- if $param.HasDefault }}

// 	{{$param.Name}}, ok := opts["{{$param.Name}}"].({{$param.TypeStr}})
// 	if !ok {
// 		{{$param.Name}} = {{$param.Default}}
// 	}

// 	{{- else if $param.IsOptional }}

// 	var {{$param.Name}} *{{$param.TypeStr}}
// 	if arg, ok := opts["{{$param.Name}}"].({{$param.TypeStr}}); ok {
// 		{{$param.Name}} = &arg
// 	}

// 	{{- else }}

// 	if _, ok := opts["{{$param.Name}}"].({{$param.TypeStr}}); !ok {
// 		return nil, fmt.Errorf("missing required argument. function: %s argument: %s", "{{ $.FunctInfo.Name }}", "{{$param.Name}}")
// 	}
// 	{{$param.Name}} := opts["{{$param.Name}}"].({{$param.TypeStr}})

// 	{{- end }}
// 	transformerOpts.{{$param.Name}} = {{$param.Name}}
// 	{{- end }}

// 	return transformerOpts, nil
// }
// `

// type TemplateData struct {
// 	SourceFile   string
// 	PackageName  string
// 	FunctInfo    FuncInfo
// 	StructName   string
// 	ImportFmt    bool
// 	HasSeedParam bool
// }

// func generateCode(pkgName string, funcInfo *FuncInfo) (string, error) {
// 	data := TemplateData{
// 		SourceFile:  funcInfo.SourceFile,
// 		PackageName: pkgName,
// 		FunctInfo:   *funcInfo,
// 		StructName:  capitalizeFirst(funcInfo.Name),
// 	}
// 	for _, p := range funcInfo.Params {
// 		if p.Name == "seed" {
// 			data.HasSeedParam = true
// 			data.ImportFmt = true
// 		}
// 		if !p.IsOptional && !p.HasDefault {
// 			data.ImportFmt = true
// 		}
// 	}
// 	t := template.Must(template.New("neosyncTransformerImpl").Parse(codeTemplate))
// 	var out bytes.Buffer
// 	err := t.Execute(&out, data)
// 	if err != nil {
// 		return "", err
// 	}
// 	return out.String(), nil
// }

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

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func lowercaseFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(string(s[0])) + s[1:]
}
