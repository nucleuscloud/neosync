package neosync_benthos_defaulttransform

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/warpstreamlabs/bento/public/service"
)

type primitiveType int

const (
	Boolean primitiveType = iota
	Byte
	Number
	String
)

func defaultTransformerProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringListField("mapped_keys")).
		Field(service.NewStringField("job_source_options_string"))
}

func ReisterDefaultTransformerProcessor(env *service.Environment) error {
	return env.RegisterBatchProcessor(
		"neosync_default_transformer",
		defaultTransformerProcessorConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			proc, err := newDefaultTransformerProcessor(conf, mgr)
			if err != nil {
				return nil, err
			}

			return proc, nil
		})
}

type defaultTransformerProcessor struct {
	mappedKeys                 map[string]struct{}
	defaultTransformersInitMap map[primitiveType]*transformer.TransformerExecutor
	logger                     *service.Logger
}

func newDefaultTransformerProcessor(conf *service.ParsedConfig, mgr *service.Resources) (*defaultTransformerProcessor, error) {
	mappedKeys, err := conf.FieldStringList("mapped_keys")
	if err != nil {
		return nil, err
	}
	mappedKeysMap := map[string]struct{}{}
	for _, k := range mappedKeys {
		mappedKeysMap[k] = struct{}{}
	}

	jobSourceOptsStr, err := conf.FieldString("job_source_options_string")
	if err != nil {
		return nil, err
	}
	var jobSourceOptions mgmtv1alpha1.JobSourceOptions
	err = protojson.Unmarshal([]byte(jobSourceOptsStr), &jobSourceOptions)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	defaultTransformerMap := getDefaultTransformerMap(&jobSourceOptions)
	defaultTransformersInitMap, err := initDefaultTransformers(defaultTransformerMap)
	if err != nil {
		return nil, err
	}

	return &defaultTransformerProcessor{
		mappedKeys:                 mappedKeysMap,
		defaultTransformersInitMap: defaultTransformersInitMap,
		logger:                     mgr.Logger(),
	}, nil
}

func getDefaultTransformerMap(jobSourceOptions *mgmtv1alpha1.JobSourceOptions) map[primitiveType]*mgmtv1alpha1.JobMappingTransformer {
	switch cfg := jobSourceOptions.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Dynamodb:
		unmappedTransformers := cfg.Dynamodb.UnmappedTransforms
		if unmappedTransformers == nil {
			return map[primitiveType]*mgmtv1alpha1.JobMappingTransformer{}
		}
		return map[primitiveType]*mgmtv1alpha1.JobMappingTransformer{
			Boolean: unmappedTransformers.Boolean,
			Byte:    unmappedTransformers.B,
			Number:  unmappedTransformers.N,
			String:  unmappedTransformers.S,
		}
	default:
		return map[primitiveType]*mgmtv1alpha1.JobMappingTransformer{}
	}
}

func (m *defaultTransformerProcessor) ProcessBatch(ctx context.Context, batch service.MessageBatch) ([]service.MessageBatch, error) {
	newBatch := make(service.MessageBatch, 0, len(batch))
	for _, msg := range batch {
		root, err := msg.AsStructuredMut()
		if err != nil {
			return nil, err
		}
		newRoot, err := m.transformRoot("", root)
		if err != nil {
			return nil, err
		}
		newMsg := msg.Copy()
		newMsg.SetStructured(newRoot)
		newBatch = append(newBatch, newMsg)
	}

	if len(newBatch) == 0 {
		return nil, nil
	}
	return []service.MessageBatch{newBatch}, nil
}

func (m *defaultTransformerProcessor) Close(context.Context) error {
	return nil
}

// returns new root
func (m *defaultTransformerProcessor) transformRoot(path string, root any) (any, error) {
	_, isMappedKey := m.mappedKeys[path] // don't mutate mapped keys
	switch v := root.(type) {
	case map[string]any:
		newMap := make(map[string]any)
		for k, v2 := range v {
			p := k
			if path != "" {
				p = fmt.Sprintf("%s.%s", path, k)
			}
			newValue, err := m.transformRoot(p, v2)
			if err != nil {
				return nil, err
			}
			newMap[k] = dereferenceValue(newValue)
		}
		return newMap, nil
	case [][]byte:
		newSlice := make([][]byte, len(v))
		for i, v2 := range v {
			p := fmt.Sprintf("[%d]", i)
			if path != "" {
				p = fmt.Sprintf("%s[%d]", path, i)
			}
			newValue, err := m.transformRoot(p, v2)
			if err != nil {
				return nil, err
			}
			bits, err := toByteSlice(newValue)
			if err != nil {
				return nil, err
			}
			newSlice[i] = bits
		}
		return newSlice, nil
	case []any:
		newSlice := make([]any, len(v))
		for i, v2 := range v {
			p := fmt.Sprintf("[%d]", i)
			if path != "" {
				p = fmt.Sprintf("%s[%d]", path, i)
			}
			newValue, err := m.transformRoot(p, v2)
			if err != nil {
				return nil, err
			}
			newSlice[i] = dereferenceValue(newValue)
		}
		return newSlice, nil
	case []byte:
		return m.getValue(Byte, v, !isMappedKey)
	case string:
		return m.getValue(String, v, !isMappedKey)
	case json.Number:
		return m.getValue(String, v, !isMappedKey)
	case float64:
		return m.getValue(Number, v, !isMappedKey)
	case int:
		return m.getValue(Number, v, !isMappedKey)
	case int64:
		return m.getValue(Number, v, !isMappedKey)
	case bool:
		return m.getValue(Boolean, v, !isMappedKey)
	default:
		return v, nil
	}
}

func (m *defaultTransformerProcessor) getValue(transformerKey primitiveType, value any, shouldMutate bool) (any, error) {
	t := m.defaultTransformersInitMap[transformerKey]
	if t != nil && shouldMutate {
		return t.Mutate(value, t.Opts)
	}
	return value, nil
}

func initDefaultTransformers(defaultTransformerMap map[primitiveType]*mgmtv1alpha1.JobMappingTransformer) (map[primitiveType]*transformer.TransformerExecutor, error) {
	transformersInit := map[primitiveType]*transformer.TransformerExecutor{}
	for k, t := range defaultTransformerMap {
		if !shouldProcess(t) {
			continue
		}
		init, err := transformer.InitializeTransformer(t)
		if err != nil {
			return nil, err
		}
		transformersInit[k] = init
	}
	return transformersInit, nil
}

func shouldProcess(t *mgmtv1alpha1.JobMappingTransformer) bool {
	switch t.GetConfig().GetConfig().(type) {
	case *mgmtv1alpha1.TransformerConfig_PassthroughConfig,
		nil:
		return false
	default:
		return true
	}
}

func dereferenceValue(value any) any {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		return rv.Elem().Interface()
	}
	return value
}

func toByteSlice(value any) ([]byte, error) {
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return []byte(fmt.Sprintf("%v", v)), nil
	default:
		if reflect.TypeOf(v).Kind() == reflect.Ptr {
			if reflect.ValueOf(v).IsNil() {
				return []byte("null"), nil
			}
			v = reflect.ValueOf(v).Elem().Interface()
		}
		return json.Marshal(v)
	}
}
