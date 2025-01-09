package main

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	transformers "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
)

func main() {
	args := os.Args
	if len(args) < 1 {
		panic("must provide necessary args")
	}

	docsPath := args[1]

	fileSet := token.NewFileSet()
	transformerFuncs := []*transformers.BenthosSpec{}

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
						transformerFuncs = append(transformerFuncs, &transformers.BenthosSpec{
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
		parsedSpec, err := transformers.ParseBloblangSpec(tf)
		if err != nil {
			fmt.Println("Error parsing bloblang params:", err) //nolint:forbidigo
		}
		tf.Params = sanitizeParamDefaults(parsedSpec.Params)
		tf.Description = parsedSpec.SpecDescription
		exampleStr, err := generateExample(tf)
		if err != nil {
			fmt.Println("Error generating example:", err) //nolint:forbidigo
			return
		}
		tf.Example = exampleStr
	}

	codeStr, err := generateCode(transformerFuncs)
	if err != nil {
		fmt.Println("Error writing to output file:", err) //nolint:forbidigo
		return
	}

	outputFile, err := os.Create(docsPath)
	if err != nil {
		fmt.Println("Error creating output file:", err) //nolint:forbidigo
		return
	}

	_, err = outputFile.WriteString(codeStr)
	if err != nil {
		fmt.Println("Error writing to output file:", err) //nolint:forbidigo
		return
	}
	outputFile.Close()

}

// makes defaults docs friendly
func sanitizeParamDefaults(params []*transformers.BenthosSpecParam) []*transformers.BenthosSpecParam {
	newParams := []*transformers.BenthosSpecParam{}
	for _, p := range params {
		var newDefault string
		if p.HasDefault && p.Default == "GenerateEmailType_UuidV4.String()" {
			newDefault = "'uuidv4'"
		} else if p.HasDefault && p.Default == "InvalidEmailAction_Reject.String()" {
			newDefault = "'reject'"
		} else if p.HasDefault && p.Default == "[]any{}" {
			newDefault = "[]"
		} else if p.HasDefault && p.Default == "time.Now().UnixNano()" {
			newDefault = "Unix timestamp in nanoseconds"
		} else {
			newDefault = p.Default
		}
		newParams = append(newParams, &transformers.BenthosSpecParam{
			Name:        p.Name,
			TypeStr:     p.TypeStr,
			IsOptional:  p.IsOptional || p.HasDefault,
			HasDefault:  p.HasDefault,
			Default:     newDefault,
			Description: p.Description,
		})
	}
	return newParams
}

const exampleTemplate = `{{- if eq .BenthosSpec.Type "transform" -}}
{{if eq (len .BenthosSpec.Params) 0}}
const newValue = neosync.{{.BenthosSpec.Name}}(value, {});
{{- else }}
const newValue = neosync.{{.BenthosSpec.Name}}(value, {
{{- range $i, $param := .BenthosSpec.Params -}}
{{- if eq $param.Name "value" -}}{{ continue }}{{- end -}}
	{{ if $param.HasDefault }}
	{{ if eq $param.Name "seed" -}}
	{{$param.Name}}: 1,
	{{- else -}}
	{{$param.Name}}: {{$param.Default}},
	{{- end }}
	{{- else }}
	{{ if eq $param.TypeStr "string"}}{{$param.Name}}: "", {{ end -}}
	{{ if eq $param.TypeStr "int64"}}{{$param.Name}}: 1, {{ end -}}
	{{ if eq $param.TypeStr "float64"}}{{$param.Name}}: 1.12, {{ end -}}
	{{ if eq $param.TypeStr "bool"}}{{$param.Name}}: false, {{ end -}}
	{{ if eq $param.TypeStr "any"}}{{$param.Name}}: "", {{ end -}}
	{{ end }}
{{- end }}
});
{{- end }}
{{- else if eq .BenthosSpec.Type "generate" -}}
{{if eq (len .BenthosSpec.Params) 0}}
const newValue = neosync.{{.BenthosSpec.Name}}({});
{{- else }}
const newValue = neosync.{{.BenthosSpec.Name}}({
	{{- range $i, $param := .BenthosSpec.Params -}}
	{{ if $param.HasDefault }}
	{{ if eq $param.Name "seed" -}}
	{{$param.Name}}: 1,
	{{- else -}}
	{{$param.Name}}: {{$param.Default}},
	{{- end }}
	{{- else }}
	{{ if eq $param.TypeStr "string"}}{{$param.Name}}: "", {{ end -}}
	{{ if eq $param.TypeStr "int64"}}{{$param.Name}}: 1, {{ end -}}
	{{ if eq $param.TypeStr "float64"}}{{$param.Name}}: 1.12, {{ end -}}
	{{ if eq $param.TypeStr "bool"}}{{$param.Name}}: false, {{ end -}}
	{{ if eq $param.TypeStr "any"}}{{$param.Name}}: "", {{ end -}}
	{{ end }}
{{- end }}
});
{{- end -}}
{{ end }}
`

type ExampleTemplateData struct {
	BenthosSpec transformers.BenthosSpec
}

func generateExample(bs *transformers.BenthosSpec) (string, error) {
	if bs == nil {
		return "", nil
	}
	data := ExampleTemplateData{
		BenthosSpec: *bs,
	}
	t := template.Must(template.New("neosyncTransformerExample").Parse(exampleTemplate))
	var out bytes.Buffer
	err := t.Execute(&out, data)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

const docTemplate = `---
title: Javascript Transformer
slug: /transformers/javascript
hide_title: false
id: javascript
description: Learn about Neosync's javascript transformer
---
<!-- prettier-ignore-start -->
<!--
	Code generated by Neosync neosync_js_transformer_docs_generator.go. DO NOT EDIT.
-->

# Neosync Javascript Transformer Functions

Learn about Neosync's Javascript transformer and generator functions, which provide a wide range of capabilities for data transformation and
generation within the Javascript Transformer and Generator. Explore detailed descriptions and examples to effectively utilize these functions in your jobs.

## Transformers

Neosync's transformer functions allow you to manipulate and transform data values with ease.
These functions are designed to provide powerful and flexible data transformation capabilities within your jobs.
Each transformer function accepts a value and a configuration object as arguments.
The source column value is accessible via the ` + "`value`" + ` keyword, while additional columns can be referenced using ` + "`input.{column_name}`" + `.
<br/>

{{range $i, $bs := .TransformerSpecs }}

<!--
source: {{$bs.SourceFile}}
-->

### {{$bs.Name}}

{{$bs.Description}}

**Parameters**

**Value**
Type: Any
Description: Value that will be transformed

**Config**

| Field    | Type | Default | Required | Description |
| -------- | ---- | ------- | -------- | ----------- |
{{- range $i, $param := $bs.Params}}
{{- if eq $param.Name "value" }}{{ continue }}{{- end }}
| {{$param.Name}} | {{$param.TypeStr}} | {{$param.Default}} | {{ if $param.IsOptional -}} false {{- else -}} true {{- end }} | {{$param.Description}}
{{- end -}}
<br/>

**Example**

` + "```javascript" + `
{{$bs.Example}}
` + "```" + `
<br/>
{{end }}

## Generators

Neosync's generator functions enable the creation of various data values, facilitating the generation of realistic and diverse data for
testing and development purposes. These functions are designed to provide robust and versatile data generation capabilities within your jobs.
Each generator function accepts a configuration object as an argument.

<br/>
{{range $i, $bs := .GeneratorSpecs }}

<!--
source: {{$bs.SourceFile}}
-->

### {{$bs.Name}}

{{$bs.Description}}

**Parameters**

**Config**

| Field    | Type | Default | Required | Description |
| -------- | ---- | ------- | -------- | ----------- |
{{range $i, $param := $bs.Params -}}
| {{$param.Name}} | {{$param.TypeStr}} | {{$param.Default}} | {{ if $param.IsOptional -}} false {{- else -}} true {{- end }} | {{ $param.Description }}
{{ end -}}
<br/>

**Example**

` + "```javascript" + `
{{$bs.Example}}
` + "```" + `
<br/>
{{end }}
<!-- prettier-ignore-end -->`

type TemplateData struct {
	TransformerSpecs []*transformers.BenthosSpec
	GeneratorSpecs   []*transformers.BenthosSpec
}

func generateCode(benthosSpecs []*transformers.BenthosSpec) (string, error) {
	transformerSpecs := []*transformers.BenthosSpec{}
	generatorSpecs := []*transformers.BenthosSpec{}

	for _, spec := range benthosSpecs {
		if spec.Type == "transform" {
			transformerSpecs = append(transformerSpecs, spec)
		} else if spec.Type == "generate" {
			generatorSpecs = append(generatorSpecs, spec)
		}
	}
	data := TemplateData{
		TransformerSpecs: transformerSpecs,
		GeneratorSpecs:   generatorSpecs,
	}
	t := template.Must(template.New("neosyncTransformerDocs").Parse(docTemplate))
	var out bytes.Buffer
	err := t.Execute(&out, data)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}
