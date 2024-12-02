package jsonanonymizer

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/itchyny/gojq"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	transformer "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
)

type AnonymizeJsonError struct {
	InputIndex int64
	Message    string
}

type JsonAnonymizer struct {
	transformerMappings        []*mgmtv1alpha1.TransformerMapping
	transformerExecutors       []*transformer.TransformerExecutor
	defaultTransformers        *mgmtv1alpha1.DefaultTransformersConfig
	defaultTransformerExecutor *DefaultExecutors
	compiledQuery              *gojq.Code
	haltOnFailure              bool
	skipPaths                  map[string]struct{}
	anonymizeConfig            *anonymizeConfig
}

type anonymizeConfig struct {
	analyze         presidioapi.AnalyzeInterface
	anonymize       presidioapi.AnonymizeInterface
	defaultLanguage *string
}

// Option is a functional option for configuring the Anonymizer
type Option func(*JsonAnonymizer)

// NewAnonymizer initializes a new Anonymizer with functional options
func NewAnonymizer(opts ...Option) (*JsonAnonymizer, error) {
	a := &JsonAnonymizer{
		transformerMappings: make([]*mgmtv1alpha1.TransformerMapping, 0),
	}
	for _, opt := range opts {
		opt(a)
	}

	if len(a.transformerMappings) == 0 && a.defaultTransformers == nil {
		return nil, fmt.Errorf("failed to initialize JSON anonymizer. must provide either default transformers or transformer mappings.")
	}

	// Initialize transformerExecutors
	var err error
	a.transformerExecutors, err = initTransformerExecutors(a.transformerMappings, a.anonymizeConfig)
	if err != nil {
		return nil, err
	}

	// Initialize defaultTransformerExecutor if needed
	if a.defaultTransformers != nil {
		a.defaultTransformerExecutor, err = initDefaultTransformerExecutors(a.defaultTransformers, a.anonymizeConfig)
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

// WithAnonymizeConfig sets the analyze and anonymize clients for use by the presidio transformers only if isEnabled is true
func WithConditionalAnonymizeConfig(isEnabled bool, analyze presidioapi.AnalyzeInterface, anonymize presidioapi.AnonymizeInterface, defaultLanguage *string) Option {
	return func(ja *JsonAnonymizer) {
		if isEnabled && analyze != nil && anonymize != nil {
			ja.anonymizeConfig = &anonymizeConfig{
				analyze:         analyze,
				anonymize:       anonymize,
				defaultLanguage: defaultLanguage,
			}
		}
	}
}

// WithTransformerMappings sets the transformer mappings
func WithTransformerMappings(mappings []*mgmtv1alpha1.TransformerMapping) Option {
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

// Compiles JQ query and initializes transformer functions
func (a *JsonAnonymizer) initializeJq() error {
	queryString, functionNames := a.buildJqQuery()
	query, err := gojq.Parse(queryString)
	if err != nil {
		return fmt.Errorf("failed to parse jq query: %v", err)
	}

	var compilerOpts []gojq.CompilerOption

	a.skipPaths = map[string]struct{}{}
	for idx, mapping := range a.transformerMappings {
		fnName := functionNames[idx]
		exec := a.transformerExecutors[idx]
		path := mapping.GetExpression()
		compilerOpts = append(compilerOpts, gojq.WithFunction(fnName, 1, 1, func(_ any, args []any) any {
			value := args[0]
			result, err := exec.Mutate(value, exec.Opts)
			if err != nil {
				return fmt.Errorf("unable to anonymize value. expression: %s  error: %w", path, err)
			}
			return derefPointer(result)
		}))

		sanitizedPath := strings.ReplaceAll(path, "?", "")
		a.skipPaths[sanitizedPath] = struct{}{}
	}

	applyDefaultTransformersFunc := func(value any, args []any) gojq.Iter {
		result := a.applyDefaultTransformers(value, "")
		return gojq.NewIter(result)
	}
	compilerOpts = append(compilerOpts, gojq.WithIterFunction("applyDefaultTransformers", 0, 0, applyDefaultTransformersFunc))

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

// Build JQ query. Sets fields to transformer functions and defines default transformer function
func (a *JsonAnonymizer) buildJqQuery() (query string, transformerFunctions []string) {
	queryParts := []string{}
	functionNames := []string{}

	if a.defaultTransformers != nil {
		if a.defaultTransformers.S != nil || a.defaultTransformers.N != nil || a.defaultTransformers.Boolean != nil {
			queryParts = append(queryParts, "applyDefaultTransformers")
		}
	}
	for idx, mapping := range a.transformerMappings {
		fieldPath := mapping.GetExpression()
		functionName := fmt.Sprintf("transformFunc_%d", idx)
		functionNames = append(functionNames, functionName)
		queryPart := fmt.Sprintf("%s? |= %s(.)", fieldPath, functionName)
		queryParts = append(queryParts, queryPart)
	}

	queryString := strings.Join(queryParts, " | ")
	return queryString, functionNames
}

// JQ function to apply all transformers to values that are unmapped in transformer mapping
func (a *JsonAnonymizer) applyDefaultTransformers(value any, path string) any {
	switch v := value.(type) {
	case map[string]any:
		newMap := make(map[string]any)
		for key, val := range v {
			newPath := fmt.Sprintf("%s.%s", path, key)
			if a.shouldSkipPath(newPath) {
				newMap[key] = val
			} else {
				newMap[key] = a.applyDefaultTransformers(val, newPath)
			}
		}
		return newMap
	case []any:
		newArray := make([]any, len(v))
		for i, elem := range v {
			newPath := fmt.Sprintf("%s[%d]", path, i)
			if a.shouldSkipPath(newPath) {
				newArray[i] = elem
			} else {
				newArray[i] = a.applyDefaultTransformers(elem, newPath)
			}
		}
		return newArray
	default:
		if a.shouldSkipPath(path) {
			return value
		} else {
			return a.executeTransformation(value)
		}
	}
}

// .departments[0].projects[1].name -> .departments[].projects[].name
func removeNumbersInBrackets(input string) string {
	// Regex pattern to match digits inside square brackets
	re := regexp.MustCompile(`\[\d+\]`)
	// Replace the digits with empty brackets
	result := re.ReplaceAllString(input, "[]")
	return result
}

func (a *JsonAnonymizer) shouldSkipPath(path string) bool {
	_, exists := a.skipPaths[path]
	if exists {
		return true
	}
	// checks for array syntax
	// ex: .departments[].projects[].name should match .departments[0].projects[1].name
	_, exists = a.skipPaths[removeNumbersInBrackets(path)]
	return exists
}

// Transforms value based on type
func (a *JsonAnonymizer) executeTransformation(value any) any {
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
}

// AnonymizeJSONObjects takes a slice of JSON strings
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

// AnonymizeJSONObject takes a JSON string
// applies the configured anonymization transformations to each object, and returns the modified JSON string.
func (a *JsonAnonymizer) AnonymizeJSONObject(jsonStr string) (string, error) {
	if jsonStr == "" {
		return jsonStr, nil
	}
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

func initTransformerExecutors(
	transformerMappings []*mgmtv1alpha1.TransformerMapping,
	anonymizeConfig *anonymizeConfig,
) ([]*transformer.TransformerExecutor, error) {
	executors := []*transformer.TransformerExecutor{}
	execOpts := []transformer.TransformerExecutorOption{}
	if anonymizeConfig != nil && anonymizeConfig.analyze != nil && anonymizeConfig.anonymize != nil {
		execOpts = append(execOpts, transformer.WithTransformPiiTextConfig(anonymizeConfig.analyze, anonymizeConfig.anonymize, anonymizeConfig.defaultLanguage))
	}

	for _, mapping := range transformerMappings {
		executor, err := transformer.InitializeTransformerByConfigType(mapping.GetTransformer(), execOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize transformer for expression '%s': %v", mapping.GetExpression(), err)
		}
		executors = append(executors, executor)
	}

	return executors, nil
}

type DefaultExecutors struct {
	S       *transformer.TransformerExecutor
	N       *transformer.TransformerExecutor
	Boolean *transformer.TransformerExecutor
}

func initDefaultTransformerExecutors(
	defaultTransformer *mgmtv1alpha1.DefaultTransformersConfig,
	anonymizeConfig *anonymizeConfig,
) (*DefaultExecutors, error) {
	execOpts := []transformer.TransformerExecutorOption{}
	if anonymizeConfig != nil && anonymizeConfig.analyze != nil && anonymizeConfig.anonymize != nil {
		execOpts = append(execOpts, transformer.WithTransformPiiTextConfig(anonymizeConfig.analyze, anonymizeConfig.anonymize, anonymizeConfig.defaultLanguage))
	}

	var stringExecutor, numberExecutor, booleanExecutor *transformer.TransformerExecutor
	var err error
	if defaultTransformer.S != nil {
		stringExecutor, err = transformer.InitializeTransformerByConfigType(defaultTransformer.S, execOpts...)
		if err != nil {
			return nil, err
		}
	}
	if defaultTransformer.N != nil {
		numberExecutor, err = transformer.InitializeTransformerByConfigType(defaultTransformer.N, execOpts...)
		if err != nil {
			return nil, err
		}
	}
	if defaultTransformer.Boolean != nil {
		booleanExecutor, err = transformer.InitializeTransformerByConfigType(defaultTransformer.Boolean, execOpts...)
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
