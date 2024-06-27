package main

import (
	"fmt"

	"github.com/dop251/goja"
)

type jsFunction func(call goja.FunctionCall, rt *goja.Runtime, logger *Logger) (interface{}, error)

type vmRunner struct {
	vm     *goja.Runtime
	logger *Logger
}

type Logger struct{}

func (l *Logger) With(key string, value string) *Logger {
	// Simplified logger implementation for demonstration purposes
	fmt.Printf("%s: %s\n", key, value)
	return l
}

func main() {
	// Simplified example to demonstrate the function
	vm := goja.New()
	vr := &vmRunner{vm: vm, logger: &Logger{}}

	function := func(call goja.FunctionCall, rt *goja.Runtime, logger *Logger) (interface{}, error) {
		logger.With("called", "function")
		return "function result", nil
	}

	err := setFunction(vr, "benthos", "myFunc", function)
	if err != nil {
		fmt.Println("Error:", err)
	}
	err = setFunction(vr, "neosync", "myNeosyncFunc", function)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func setFunction(vr *vmRunner, namespace, name string, function jsFunction) error {
	fmt.Println("-------")

	var targetObj *goja.Object
	if targetObjValue := vr.vm.GlobalObject().Get(namespace); targetObjValue != nil {
		targetObj = targetObjValue.ToObject(vr.vm)
		fmt.Println("Namespace found:", namespace)
	} else {
		fmt.Println("Namespace not found, creating new one:", namespace)
	}

	if targetObj == nil {
		if err := vr.vm.GlobalObject().Set(namespace, map[string]interface{}{}); err != nil {
			return fmt.Errorf("failed to set global %s object: %w", namespace, err)
		}
		targetObj = vr.vm.GlobalObject().Get(namespace).ToObject(vr.vm)
		fmt.Println("New namespace created:", namespace)
	}

	if err := targetObj.Set(name, func(call goja.FunctionCall, rt *goja.Runtime) goja.Value {
		l := vr.logger.With("function", name)
		result, err := function(call, rt, l)
		if err != nil {
			panic(rt.ToValue(err.Error()))
		}
		return rt.ToValue(result)
	}); err != nil {
		return fmt.Errorf("failed to set global function %v: %w", name, err)
	}

	fmt.Println("Global object properties for", "benthos", ":")
	x := vr.vm.GlobalObject().Get("benthos")
	if x != nil {
		for key, value := range x.ToObject(vr.vm).Export().(map[string]interface{}) {
			fmt.Printf("%s: %v\n", key, value)
		}
	} else {
		fmt.Println("No properties found in namespace", "benthos")
	}
	fmt.Println("Global object properties for neosync", ":")
	x = vr.vm.GlobalObject().Get("neosync")
	if x != nil {
		for key, value := range x.ToObject(vr.vm).Export().(map[string]interface{}) {
			fmt.Printf("%s: %v\n", key, value)
		}
	} else {
		fmt.Println("No properties found in namespace", "neosync")
	}

	return nil
}
