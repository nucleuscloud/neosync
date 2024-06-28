//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

const functionTemplate = `
package transformers
var _ = registerVMRunnerFunction("{{.Name}}", ` + "`{{.Description}}`" + `).
	Namespace(neosyncFnCtxName).
	{{range .Params}}Param("{{.Name}}", "{{.Type}}", "{{.Description}}").
	{{end}}Example(` + "`neosync.{{.Name}}(\"example\");`" + `).
	FnCtor(func(r *vmRunner) jsFunction {
		return func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (interface{}, error) {
			var (
				{{range .Params}}{{.GoName}} {{.GoType}}
				{{end}}
			)
			if err := parseArgs(call, &{{.Params | firstGoName}}); err != nil {
				return "", err
			}
			{{range .Params}}
			{{if .IsOptional}}if opts != nil && opts["{{.GoName}}"] != nil {
				{{.GoName}} = opts["{{.GoName}}"].({{.GoType}})
			} else {
				{{if .HasDefault}}{{.GoName}} = {{.Default}}{{end}}
			}
			{{else}}// Handling non-optional parameters
			{{end}}
			{{end}}
			randomizer := rng.New(seed)
			return transformer.{{.TransformerFunc}}(randomizer, name, preserveLength, maxLength)
		}
	})
`

type ParamInfo struct {
	Name        string
	Type        string
	Description string
	GoName      string
	GoType      string
	IsOptional  bool
	HasDefault  bool
	Default     string
}

type FuncInfo struct {
	Name            string
	Description     string
	Params          []ParamInfo
	TransformerFunc string
}

func main() {
	// Create a new FileSet
	fileSet := token.NewFileSet()

	// Slice to store information about the functions to be generated
	// funcs := []FuncInfo{}
	fileNodes := []*ast.File{}

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(path) == ".go" {
			node, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
			if err != nil {
				log.Printf("Failed to parse file %s: %v", path, err)
				return nil
			}
			// Iterate over the declarations in the parsed file
			for _, decl := range node.Decls {

				if fn, isFn := decl.(*ast.FuncDecl); isFn && fn.Doc != nil {
					for _, comment := range fn.Doc.List {
						if strings.HasPrefix(comment.Text, "// +javascriptFncBuilder:") {
							// jsonF, _ := json.MarshalIndent(decl, "", " ")
							// fmt.Printf("%s \n", string(jsonF))
							// parts := strings.Split(comment.Text, ":")
							// if len(parts) < 4 {
							// 	continue
							// }
							// fnType := parts[2]
							// fnName := parts[3]
							fileNodes = append(fileNodes, node)

						}
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("impossible to walk directories: %s", err)
	}

	// bloblang.NewPluginSpec().
	// 	Param(bloblang.NewInt64Param("max_length").Default(10000)).
	// 	Param(bloblang.NewAnyParam("value").Optional()).
	// 	Param(bloblang.NewBoolParam("preserve_length").Default(false)).
	// 	Param(bloblang.NewInt64Param("seed").Optional())
	// var benthosSpec ast.Expr
	// for _, node := range fileNodes {
	// 	ast.Inspect(node, func(n ast.Node) bool {
	// 		// Check if the node is a function call expression
	// 		if callExpr, ok := n.(*ast.CallExpr); ok {
	// 			// jsonF, _ := json.MarshalIndent(callExpr, "", " ")
	// 			// fmt.Printf("%s \n", string(jsonF))
	// 			benthosSpec = inspectChainedCalls(callExpr)
	// 		}
	// 		return true
	// 	})
	// }
	// jsonF, _ := json.MarshalIndent(benthosSpec, "", " ")
	// fmt.Printf("%s \n", string(jsonF))
	// Traverse the AST and look for bloblang.NewPluginSpec calls
	for _, node := range fileNodes {
		ast.Inspect(node, func(n ast.Node) bool {
			// Check if the node is a function call expression
			if callExpr, ok := n.(*ast.CallExpr); ok {
				jsonF, _ := json.MarshalIndent(callExpr, "", " ")
				fmt.Printf("%s \n", string(jsonF))
				fmt.Println()
				fmt.Println()
				fmt.Println()

				// Check if the function being called is bloblang.NewPluginSpec
				if fun, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if pkg, ok := fun.X.(*ast.Ident); ok && fun.Sel.Name == "Param" {
						jsonF, _ := json.MarshalIndent(callExpr, "", " ")
						fmt.Printf("%s \n", string(jsonF))

						fmt.Println("Found call to bloblang.NewPluginSpec")
						fmt.Println(pkg)

						// Print details of the call
						for _, arg := range callExpr.Args {
							fmt.Printf("Argument: %s\n", arg)
						}

						// Look for chained method calls
						inspectChainedCalls(callExpr)
					}
				}
			}
			return true
		})
	}

}

func inspectChainedCalls(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		if fun, ok := e.Fun.(*ast.SelectorExpr); ok {
			fmt.Printf("Method: %s\n", fun.Sel.Name)
			inspectChainedCalls(fun.X)
		}
	}
}

// func inspectChainedCalls(expr ast.Expr) ast.Expr {
// 	switch e := expr.(type) {
// 	case *ast.CallExpr:
// 		if fun, ok := e.Fun.(*ast.SelectorExpr); ok {
// 			if fun.Sel.Name == "Param" {
// 				fmt.Println()
// 				// jsonF, _ := json.MarshalIndent(fun., "", " ")
// 				// fmt.Printf("%s \n", string(jsonF))
// 				fmt.Printf("Method: %s\n", fun.Sel.Name)
// 				return expr
// 			}
// 			return inspectChainedCalls(fun.X)
// 		}
// 	}
// 	return expr
// }

// extractParams extracts the parameters from the function body and returns a slice of ParamInfo
func extractParams(fn *ast.FuncDecl) []ParamInfo {
	params := []ParamInfo{}
	for _, stmt := range fn.Body.List {
		if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
			if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
				for _, arg := range callExpr.Args {
					if call, ok := arg.(*ast.CallExpr); ok {
						paramName := getParamName(call)
						paramType := getParamType(call)
						paramDefault := getParamDefault(call)
						isOptional := paramDefault != ""
						hasDefault := paramDefault != ""

						params = append(params, ParamInfo{
							Name:        paramName,
							Type:        paramType,
							Description: getParamDescription(paramName),
							GoName:      toCamelCase(paramName),
							GoType:      toGoType(paramType),
							IsOptional:  isOptional,
							HasDefault:  hasDefault,
							Default:     paramDefault,
						})
					}
				}
			}
		}
	}
	return params
}

// Helper functions to extract parameter information
func getParamName(call *ast.CallExpr) string {
	if selExpr, ok := call.Fun.(*ast.SelectorExpr); ok {
		return selExpr.Sel.Name
	}
	return ""
}

func getParamType(call *ast.CallExpr) string {
	if selExpr, ok := call.Fun.(*ast.SelectorExpr); ok {
		switch selExpr.Sel.Name {
		case "NewInt64Param":
			return "int64"
		case "NewAnyParam":
			return "any"
		case "NewBoolParam":
			return "bool"
		}
	}
	return ""
}

func getParamDefault(call *ast.CallExpr) string {
	for _, arg := range call.Args {
		if basicLit, ok := arg.(*ast.BasicLit); ok {
			return basicLit.Value
		}
	}
	return ""
}

func getParamDescription(name string) string {
	switch name {
	case "max_length":
		return "Maximum length"
	case "value":
		return "Value to use"
	case "preserve_length":
		return "Preserve the length"
	case "seed":
		return "Seed for randomness"
	default:
		return ""
	}
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

func toGoType(t string) string {
	switch t {
	case "int64":
		return "int64"
	case "any":
		return "interface{}"
	case "bool":
		return "bool"
	default:
		return "interface{}"
	}
}

func getDescription(fnType string) string {
	switch fnType {
	case "transform":
		return "Transforms first name"
	case "generate":
		return "Generates first name"
	default:
		return ""
	}
}

func getTransformerFunc(fnType string) string {
	switch fnType {
	case "transform":
		return "TransformFirstName"
	case "generate":
		return "GenerateRandomFirstName"
	default:
		return ""
	}
}
