package main

import (
	"fmt"
	"reflect"
)

func createMultiDimMapArray(depth int) interface{} {
	if depth <= 0 {
		return map[string]any{"key": "value"}
	}

	return []interface{}{createMultiDimMapArray(depth - 1)}
}

func main() {
	// Example usage
	result1 := createMultiDimMapArray(0)
	result2 := createMultiDimMapArray(1)
	result3 := createMultiDimMapArray(2)
	result4 := createMultiDimMapArray(3)

	fmt.Printf("Result 1 (depth 0): %v, Type: %v\n", result1, reflect.TypeOf(result1))
	fmt.Printf("Result 2 (depth 1): %v, Type: %v\n", result2, reflect.TypeOf(result2))
	fmt.Printf("Result 3 (depth 2): %v, Type: %v\n", result3, reflect.TypeOf(result3))
	fmt.Printf("Result 4 (depth 3): %v, Type: %v\n", result4, reflect.TypeOf(result4))
}
