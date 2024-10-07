package jsonanonymizer

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/itchyny/gojq"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
)

type AnonymizeJsonError struct {
	InputIndex int64
	FieldPath  string
	Message    string
}

type JsonAnonymizer struct {
	transformerMappings        map[string]*mgmtv1alpha1.TransformerConfig
	transformerExecutors       map[string]*transformer.TransformerExecutor
	defaultTransformers        *mgmtv1alpha1.DefaultTransformersConfig
	defaultTransformerExecutor *DefaultExecutors
	compiledQuery              *gojq.Code
	haltOnFailure              bool
	skipPaths                  map[string]struct{}
}

// Option is a functional option for configuring the Anonymizer
type Option func(*JsonAnonymizer)

// NewAnonymizer initializes a new Anonymizer with functional options
func NewAnonymizer(opts ...Option) (*JsonAnonymizer, error) {
	a := &JsonAnonymizer{
		transformerMappings: make(map[string]*mgmtv1alpha1.TransformerConfig),
	}
	for _, opt := range opts {
		opt(a)
	}

	if len(a.transformerMappings) == 0 && a.defaultTransformers == nil {
		return nil, fmt.Errorf("failed to initialize JSON anonymizer. must provide either default transformers or transformer mappings.")
	}

	// Initialize transformerExecutors
	var err error
	a.transformerExecutors, err = initTransformerExecutors(a.transformerMappings)
	if err != nil {
		return nil, err
	}

	// Initialize defaultTransformerExecutor if needed
	if a.defaultTransformers != nil {
		a.defaultTransformerExecutor, err = initDefaultTransformerExecutors(a.defaultTransformers)
		if err != nil {
			return nil, err
		}
	}

	// Initialize jq
	if err := a.initializeJq(); err != nil {
		return nil, err
	}
	return a, nil
}

// WithTransformerMappings sets the transformer mappings
func WithTransformerMappings(mappings map[string]*mgmtv1alpha1.TransformerConfig) Option {
	return func(a *JsonAnonymizer) {
		if mappings != nil {
			a.transformerMappings = mappings
		}
	}
}

// WithDefaultTransformers sets the default transformers
func WithDefaultTransformers(defaults *mgmtv1alpha1.DefaultTransformersConfig) Option {
	return func(a *JsonAnonymizer) {
		a.defaultTransformers = defaults
	}
}

// WithHaltOnFailure sets the haltOnFailure flag
func WithHaltOnFailure(halt bool) Option {
	return func(a *JsonAnonymizer) {
		a.haltOnFailure = halt
	}
}

func (a *JsonAnonymizer) initializeJq() error {
	queryString, functionMap, err := a.buildJqQuery()
	if err != nil {
		return err
	}
	fmt.Println(queryString)
	query, err := gojq.Parse(queryString)
	if err != nil {
		return fmt.Errorf("failed to parse jq query: %v", err)
	}

	var compilerOpts []gojq.CompilerOption

	a.skipPaths = map[string]struct{}{}
	for functionName, fieldPath := range functionMap {
		executor := a.transformerExecutors[fieldPath]
		fnName := functionName
		exec := executor
		path := fieldPath
		compilerOpts = append(compilerOpts, gojq.WithFunction(fnName, 1, 1, func(_ any, args []any) any {
			value := args[0]
			result, err := exec.Mutate(value, exec.Opts)
			if err != nil {
				return fmt.Errorf("unable to anonymize value. field_path: %s  error: %w", path, err)
			}
			fmt.Println("derefPointer(result)", derefPointer(result))
			return derefPointer(result)
		}))

		cleanPath := strings.ReplaceAll(fieldPath, "[", ".")
		cleanPath = strings.ReplaceAll(path, "]", "")
		a.skipPaths[cleanPath] = struct{}{}
	}

	// if a.defaultTransformers != nil && a.defaultTransformerExecutor != nil {
	// 	if a.defaultTransformerExecutor.S != nil {
	// 		executor := a.defaultTransformerExecutor.S
	// 		compilerOpts = append(compilerOpts, gojq.WithFunction("anonymizeString", 1, 1, func(_ any, args []any) any {
	// 			value := args[0]
	// 			result, err := executor.Mutate(value, executor.Opts)
	// 			if err != nil {
	// 				return value // what to do here need to return error
	// 			}
	// 			return derefPointer(result)
	// 		}))
	// 	}
	// 	if a.defaultTransformerExecutor.N != nil {
	// 		executor := a.defaultTransformerExecutor.N
	// 		compilerOpts = append(compilerOpts, gojq.WithFunction("anonymizeNumber", 1, 1, func(_ any, args []any) any {
	// 			value := args[0]
	// 			result, err := executor.Mutate(value, executor.Opts)
	// 			if err != nil {
	// 				return value
	// 			}
	// 			return derefPointer(result)
	// 		}))
	// 	}
	// 	if a.defaultTransformerExecutor.Boolean != nil {
	// 		executor := a.defaultTransformerExecutor.Boolean
	// 		compilerOpts = append(compilerOpts, gojq.WithFunction("anonymizeBoolean", 1, 1, func(_ any, args []any) any {
	// 			value := args[0]
	// 			result, err := executor.Mutate(value, executor.Opts)
	// 			if err != nil {
	// 				return value
	// 			}
	// 			return derefPointer(result)
	// 		}))
	// 	}
	// }

	myWalkFunc := func(value any, args []any) gojq.Iter {
		result := a.myWalk(value, nil)
		return gojq.NewIter(result)
	}
	compilerOpts = append(compilerOpts, gojq.WithIterFunction("myWalk", 0, 0, myWalkFunc))

	compiledQuery, err := gojq.Compile(query, compilerOpts...)
	if err != nil {
		return fmt.Errorf("failed to compile jq query: %v", err)
	}

	a.compiledQuery = compiledQuery
	return nil
}

func derefPointer(v any) any {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	return rv.Interface()
}

func customHash(input string) string {
	hasher := sha1.New()
	hasher.Write([]byte(input))
	hash := hex.EncodeToString(hasher.Sum(nil))

	// replace leading digit with a
	if strings.IndexAny(hash[:1], "0123456789") != -1 {
		hash = "a" + hash
	}

	return hash
}

func generateFunctionName(fieldPath string) string {
	return customHash(fieldPath)
}

func (a *JsonAnonymizer) buildJqQuery() (string, map[string]string, error) {
	queryParts := []string{}
	functionMap := make(map[string]string) // functionName -> fieldPath

	for fieldPath := range a.transformerMappings {
		functionName := generateFunctionName(fieldPath)
		functionMap[functionName] = fieldPath
		queryPart := fmt.Sprintf("%s? |= %s(.)", fieldPath, functionName)
		queryParts = append(queryParts, queryPart)
	}
	// // Handle default transformers
	// if a.defaultTransformers != nil {
	// 	walkConditions := []string{}

	// 	if a.defaultTransformers.S != nil {
	// 		walkConditions = append(walkConditions, `type == "string" then anonymizeString(.)`)
	// 	}
	// 	if a.defaultTransformers.N != nil {
	// 		walkConditions = append(walkConditions, `type == "number" then anonymizeNumber(.)`)
	// 	}
	// 	if a.defaultTransformers.Boolean != nil {
	// 		walkConditions = append(walkConditions, `type == "boolean" then anonymizeBoolean(.)`)
	// 	}

	// 	if len(walkConditions) > 0 {
	// 		walkQuery := fmt.Sprintf("walk(if %s else . end)", strings.Join(walkConditions, " elif "))
	// 		queryParts = append(queryParts, walkQuery)
	// 	}
	// }
	if a.defaultTransformers != nil {
		if a.defaultTransformers.S != nil || a.defaultTransformers.N != nil || a.defaultTransformers.Boolean != nil {
			queryParts = append(queryParts, "myWalk")
		}
	}

	queryString := strings.Join(queryParts, " | ")
	fmt.Println("## JQ QUERY")
	fmt.Println(fmt.Sprintf("%s", queryString))
	fmt.Println()
	return queryString, functionMap, nil
}

func (a *JsonAnonymizer) myWalk(value any, path []string) any {
	fmt.Println(
		"value", value,
		"path", path,
	)
	switch v := value.(type) {
	case map[string]any:
		newMap := make(map[string]any)
		for key, val := range v {
			newPath := append(path, key)
			fullPath := strings.Join(newPath, ".")
			if a.isSkipPath(fmt.Sprintf(".%s", fullPath)) {
				newMap[key] = val
			} else {
				newMap[key] = a.myWalk(val, newPath)
			}
		}
		return newMap
	case []any:
		newArray := make([]any, len(v))
		for i, elem := range v {
			indexStr := strconv.Itoa(i)
			newPath := append(path, indexStr)
			fullPath := strings.Join(newPath, ".")
			if a.isSkipPath(fmt.Sprintf(".%s", fullPath)) {
				newArray[i] = elem
			} else {
				newArray[i] = a.myWalk(elem, newPath)
			}
		}
		return newArray
	default:
		// fullPath := strings.Join(path, ".")
		// return a.applyTransformations(value, fullPath)
		fullPath := strings.Join(path, ".")
		if a.isSkipPath(fmt.Sprintf(".%s", fullPath)) {
			return value
		} else {
			return a.applyTransformations(value, fullPath)
		}
	}
}

func (a *JsonAnonymizer) isSkipPath(path string) bool {
	_, exists := a.skipPaths[path]
	return exists
}

func (a *JsonAnonymizer) applyTransformations(value any, fullPath string) any {
	// if executor, exists := a.transformerExecutors[fullPath]; exists {
	// 	// Apply specific transformer for this path
	// 	result, err := executor.Mutate(value, executor.Opts)
	// 	if err != nil {
	// 		return value // Handle error as needed
	// 	}
	// 	return derefPointer(result)
	// } else {
	// Apply default transformers
	switch v := value.(type) {
	case string:
		if a.defaultTransformerExecutor != nil && a.defaultTransformerExecutor.S != nil {
			result, err := a.defaultTransformerExecutor.S.Mutate(v, a.defaultTransformerExecutor.S.Opts)
			if err != nil {
				return v
			}
			return derefPointer(result)
		}
		return v
	case float64, int, int64:
		if a.defaultTransformerExecutor != nil && a.defaultTransformerExecutor.N != nil {
			result, err := a.defaultTransformerExecutor.N.Mutate(v, a.defaultTransformerExecutor.N.Opts)
			if err != nil {
				return v
			}
			return derefPointer(result)
		}
		return v
	case bool:
		if a.defaultTransformerExecutor != nil && a.defaultTransformerExecutor.Boolean != nil {
			result, err := a.defaultTransformerExecutor.Boolean.Mutate(v, a.defaultTransformerExecutor.Boolean.Opts)
			if err != nil {
				return v
			}
			return derefPointer(result)
		}
		return v
	default:
		return v
	}
	// }
}

// AnonymizeJSONObjects takes a JSON string representing an array of objects
// applies the configured anonymization transformations to each object, and returns the modified JSON string.
func (a *JsonAnonymizer) AnonymizeJSONObjects(jsonStrs []string) ([]string, []*AnonymizeJsonError) {
	anonymizeErrors := []*AnonymizeJsonError{}
	anonymizedJsonStrs := []string{}
	for idx, jStr := range jsonStrs {
		processedJSON, err := a.AnonymizeJSONObject(jStr)
		if err != nil {
			anonymizeErrors = append(anonymizeErrors, &AnonymizeJsonError{
				InputIndex: int64(idx),
				Message:    err.Error(),
			})
			if a.haltOnFailure {
				return anonymizedJsonStrs, anonymizeErrors
			}
		}
		anonymizedJsonStrs = append(anonymizedJsonStrs, processedJSON)
	}

	return anonymizedJsonStrs, anonymizeErrors
}

func (a *JsonAnonymizer) AnonymizeJSONObject(jsonStr string) (string, error) {
	var data any
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON string: %v", err)
	}
	iter := a.compiledQuery.Run(data)
	result, ok := iter.Next()
	if !ok {
		return "", fmt.Errorf("failed to anonymize JSON: unknown error")
	}
	if err, ok := result.(error); ok {
		return "", fmt.Errorf("failed to anonymize JSON: %v", err)
	}

	processedJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal anonymized data: %v", err)
	}

	return string(processedJSON), nil
}

func initTransformerExecutors(transformerMappings map[string]*mgmtv1alpha1.TransformerConfig) (map[string]*transformer.TransformerExecutor, error) {
	executorMap := map[string]*transformer.TransformerExecutor{}

	for fieldPath, transformerConfig := range transformerMappings {
		executor, err := transformer.InitializeTransformerByConfigType(transformerConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize transformer for field '%s': %v", fieldPath, err)
		}
		executorMap[fieldPath] = executor
	}

	return executorMap, nil
}

type DefaultExecutors struct {
	S       *transformer.TransformerExecutor
	N       *transformer.TransformerExecutor
	Boolean *transformer.TransformerExecutor
}

func initDefaultTransformerExecutors(defaultTransformer *mgmtv1alpha1.DefaultTransformersConfig) (*DefaultExecutors, error) {
	var stringExecutor, numberExecutor, booleanExecutor *transformer.TransformerExecutor
	var err error
	if defaultTransformer.S != nil {
		stringExecutor, err = transformer.InitializeTransformerByConfigType(defaultTransformer.S)
		if err != nil {
			return nil, err
		}
	}
	if defaultTransformer.N != nil {
		numberExecutor, err = transformer.InitializeTransformerByConfigType(defaultTransformer.N)
		if err != nil {
			return nil, err
		}
	}
	if defaultTransformer.Boolean != nil {
		booleanExecutor, err = transformer.InitializeTransformerByConfigType(defaultTransformer.Boolean)
		if err != nil {
			return nil, err
		}
	}
	return &DefaultExecutors{
		S:       stringExecutor,
		N:       numberExecutor,
		Boolean: booleanExecutor,
	}, nil
}
