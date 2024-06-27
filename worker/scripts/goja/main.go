package main

import (
	"fmt"
	"io"
	"os"

	"github.com/dop251/goja"
	wasmer "github.com/wasmerio/wasmer-go/wasmer"
)

func main() {
	// goja
	vm := goja.New()

	jsCode := `
        function callWasm() {
            return addFunction("s",3);
        }
    `

	_, err := vm.RunString(jsCode)
	if err != nil {
		panic(err)
	}

	// wasm
	file, err := os.Open("../wasm/simple.wasm")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	wasmBytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	module, err := wasmer.NewModule(store, wasmBytes)
	if err != nil {
		panic(err)
	}

	importObject := wasmer.NewImportObject()
	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		panic(err)
	}

	wasmFunction, err := instance.Exports.GetFunction("add") // Replace with your actual Wasm function name
	if err != nil {
		panic(err)
	}
	result, err := wasmFunction(2, 2)
	if err != nil {
		panic(err)
	}
	fmt.Println("Result from Wasm function:", result)

	// goja
	vm.Set("addFunction", func(call goja.FunctionCall) goja.Value {
		// result, err := wasmFunction(2, 3)
		// if err != nil {
		// 	panic(err)
		// }
		result := sum(int(call.Arguments[0].ToInteger()), int(call.Arguments[1].ToInteger()))
		return vm.ToValue(result)
	})

	result, err = vm.RunString("callWasm()")
	if err != nil {
		panic(err)
	}

	fmt.Println("Result from goja Wasm function:", result)
}

func sum(x int, y int) int {
	return x + y
}
