package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

type ParamInfo struct {
	Name       string
	TypeStr    string
	IsOptional bool
	HasDefault bool
	Default    string
}

func main() {
	// Define the input string
	src := `package main

func main() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewAnyParam("email").Optional()).
		Param(bloblang.NewBoolParam("preserve_length").Default(false)).
		Param(bloblang.NewBoolParam("preserve_domain").Default(false)).
		Param(bloblang.NewAnyParam("excluded_domains").Default([]any{})).
		Param(bloblang.NewInt64Param("max_length").Default(10000)).
		Param(bloblang.NewInt64Param("seed").Optional()).
		Param(bloblang.NewStringParam("email_type").Default(GenerateEmailType_UuidV4.String())).
		Param(bloblang.NewStringParam("invalid_email_action").Default(InvalidEmailAction_Reject.String()))
}
`

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	var params []ParamInfo

	// Inspect the AST
	ast.Inspect(node, func(n ast.Node) bool {
		// Find CallExpr nodes
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if x, ok := sel.X.(*ast.CallExpr); ok {
					if fun, ok := x.Fun.(*ast.SelectorExpr); ok && fun.Sel.Name == "Param" {
						// Extract parameter type and name
						paramType := x.Fun.(*ast.SelectorExpr).Sel.Name[:len(fun.Sel.Name)-5]
						if len(x.Args) > 0 {
							if basicLit, ok := x.Args[0].(*ast.BasicLit); ok {
								paramName, _ := strconv.Unquote(basicLit.Value)
								param := ParamInfo{
									TypeStr: paramType,
									Name:    paramName,
								}

								// Check for Optional and Default methods
								for _, arg := range callExpr.Args {
									if sel, ok := arg.(*ast.SelectorExpr); ok {
										switch sel.Sel.Name {
										case "Optional":
											param.IsOptional = true
										case "Default":
											param.HasDefault = true
											if len(sel.X.(*ast.CallExpr).Args) > 0 {
												defaultVal := sel.X.(*ast.CallExpr).Args[0]
												param.Default = fmt.Sprintf("%s", defaultVal)
											}
										}
									}
								}

								params = append(params, param)
							}
						}
					}
				}
			}
		}
		return true
	})

	// Print the extracted parameters
	for _, param := range params {
		fmt.Printf("%+v\n", param)
	}
}
