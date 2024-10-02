package anonymizer

import (
	"encoding/json"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
)

type primitiveType int

const (
	Boolean primitiveType = iota
	Byte
	Number
	String
)

// Anonymizer struct
type Anonymizer struct {
	transformerMappings        map[string]*mgmtv1alpha1.TransformerConfig
	transformerExecutors       map[string]*transformer.TransformerExecutor
	defaultTransformers        *mgmtv1alpha1.DefaultTransformersConfig
	defaultTransformerExecutor *DefaultExecutors
	haltOnFailure              bool
}

// Option is a functional option for configuring the Anonymizer
type Option func(*Anonymizer)

// NewAnonymizer initializes a new Anonymizer with functional options
func NewAnonymizer(opts ...Option) (*Anonymizer, error) {
	a := &Anonymizer{
		transformerMappings: make(map[string]*mgmtv1alpha1.TransformerConfig),
	}
	for _, opt := range opts {
		opt(a)
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
	return a, nil
}

// WithTransformerMappings sets the transformer mappings
func WithTransformerMappings(mappings map[string]*mgmtv1alpha1.TransformerConfig) Option {
	return func(a *Anonymizer) {
		if mappings != nil {
			a.transformerMappings = mappings
		}
	}
}

// WithDefaultTransformers sets the default transformers
func WithDefaultTransformers(defaults *mgmtv1alpha1.DefaultTransformersConfig) Option {
	return func(a *Anonymizer) {
		a.defaultTransformers = defaults
	}
}

// WithHaltOnFailure sets the haltOnFailure flag
func WithHaltOnFailure(halt bool) Option {
	return func(a *Anonymizer) {
		a.haltOnFailure = halt
	}
}

// AnonymizeSingle processes a single JSON string
func (a *Anonymizer) AnonymizeSingle(jsonStr string) (string, error) {
	var data any
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	processedData, err := a.processData(data, "")
	if err != nil {
		return "", err
	}

	processedJSON, err := json.Marshal(processedData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal processed data: %v", err)
	}

	return string(processedJSON), nil
}

// AnonymizeMany processes multiple JSON strings
func (a *Anonymizer) AnonymizeMany(jsonStrs []string) ([]string, []*mgmtv1alpha1.AnonymizeManyErrors, error) {
	var outputData []string
	var errors []*mgmtv1alpha1.AnonymizeManyErrors

	for idx, jsonStr := range jsonStrs {
		output, err := a.AnonymizeSingle(jsonStr)
		if err != nil {
			errors = append(errors, &mgmtv1alpha1.AnonymizeManyErrors{
				InputIndex:   int64(idx),
				FieldPath:    "",
				ErrorMessage: err.Error(),
			})
			if a.haltOnFailure {
				return outputData, errors, nil
			}
			continue
		}
		outputData = append(outputData, output)
	}

	return outputData, errors, nil
}

// processData recursively processes the data structure
func (a *Anonymizer) processData(data any, parentPath string) (any, error) {
	switch v := data.(type) {
	case map[string]any:
		result := make(map[string]any)
		for key, value := range v {
			fieldPath := key
			if parentPath != "" {
				fieldPath = parentPath + "." + key
			}
			processedValue, err := a.processData(value, fieldPath)
			if err != nil {
				return nil, err
			}
			result[key] = processedValue
		}
		return result, nil

	case []any:
		result := make([]any, len(v))
		for i, elem := range v {
			var fieldPath string
			if parentPath != "" {
				fieldPath = fmt.Sprintf("%s[%d]", parentPath, i)
			} else {
				fieldPath = fmt.Sprintf("[%d]", i)
			}
			processedElem, err := a.processData(elem, fieldPath)
			if err != nil {
				return nil, err
			}
			result[i] = processedElem
		}
		return result, nil

	default:
		fieldPath := parentPath
		executor, ok := a.transformerExecutors[fieldPath]

		var transformedValue interface{}
		var err error

		if ok {
			transformedValue, err = a.applyTransformer(v, executor)
		} else if a.defaultTransformers != nil {
			transformedValue, err = a.applyDefaultTransformer(v)
		} else {
			transformedValue = v
		}

		if err != nil {
			return nil, err
		}

		return transformedValue, nil
	}
}

// applyTransformer applies the specified transformer to the value
func (a *Anonymizer) applyTransformer(value any, executor *transformer.TransformerExecutor) (any, error) {
	if executor == nil {
		return value, nil
	}
	return executor.Mutate(value, executor.Opts)
}

// applyDefaultTransformer applies the default transformer based on the value's type
func (a *Anonymizer) applyDefaultTransformer(value any) (any, error) {
	switch value.(type) {
	case float64, int, int64:
		return a.applyTransformer(value, a.defaultTransformerExecutor.N)
	case string:
		return a.applyTransformer(value, a.defaultTransformerExecutor.S)
	case bool:
		return a.applyTransformer(value, a.defaultTransformerExecutor.Boolean)
	case []byte:
		return a.applyTransformer(value, a.defaultTransformerExecutor.B)
	default:
		return value, nil
	}
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
	B       *transformer.TransformerExecutor
	N       *transformer.TransformerExecutor
	Boolean *transformer.TransformerExecutor
}

func initDefaultTransformerExecutors(defaultTransformer *mgmtv1alpha1.DefaultTransformersConfig) (*DefaultExecutors, error) {
	var stringExecutor, byteExecutor, numberExecutor, booleanExecutor *transformer.TransformerExecutor
	if defaultTransformer.S != nil {
		sExecutor, err := transformer.InitializeTransformerByConfigType(defaultTransformer.S)
		if err != nil {
			return nil, err
		}
		stringExecutor = sExecutor
	}
	if defaultTransformer.B != nil {
		bExecutor, err := transformer.InitializeTransformerByConfigType(defaultTransformer.B)
		if err != nil {
			return nil, err
		}
		byteExecutor = bExecutor
	}
	if defaultTransformer.N != nil {
		nExecutor, err := transformer.InitializeTransformerByConfigType(defaultTransformer.N)
		if err != nil {
			return nil, err
		}
		numberExecutor = nExecutor
	}
	if defaultTransformer.Boolean != nil {
		bExecutor, err := transformer.InitializeTransformerByConfigType(defaultTransformer.Boolean)
		if err != nil {
			return nil, err
		}
		booleanExecutor = bExecutor
	}
	return &DefaultExecutors{
		S:       stringExecutor,
		B:       byteExecutor,
		N:       numberExecutor,
		Boolean: booleanExecutor,
	}, nil
}

func shouldProcess(t *mgmtv1alpha1.TransformerConfig) bool {
	if t == nil {
		return false
	}

	switch t.Config.(type) {
	case *mgmtv1alpha1.TransformerConfig_PassthroughConfig:
		return false
	default:
		return true
	}
}
