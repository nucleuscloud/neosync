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

type PrimitiveType int

const (
	Boolean PrimitiveType = iota
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
		"neosync_default_mapping",
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
	defaultTransformerMap      map[PrimitiveType]*mgmtv1alpha1.JobMappingTransformer
	defaultTransformersInitMap map[PrimitiveType]*DefaultTransformer
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

	dtmStr, err := conf.FieldString("job_source_options_string")
	if err != nil {
		return nil, err
	}
	var jobSourceOptions mgmtv1alpha1.JobSourceOptions
	err = protojson.Unmarshal([]byte(dtmStr), &jobSourceOptions)
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
		defaultTransformerMap:      defaultTransformerMap,
		defaultTransformersInitMap: defaultTransformersInitMap,
		logger:                     mgr.Logger(),
	}, nil

}

func getDefaultTransformerMap(jobSourceOptions *mgmtv1alpha1.JobSourceOptions) map[PrimitiveType]*mgmtv1alpha1.JobMappingTransformer {
	switch cfg := jobSourceOptions.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Dynamodb:
		unmappedTransformers := cfg.Dynamodb.UnmappedTransforms
		return map[PrimitiveType]*mgmtv1alpha1.JobMappingTransformer{
			Boolean: unmappedTransformers.Boolean,
			Byte:    unmappedTransformers.B,
			Number:  unmappedTransformers.N,
			String:  unmappedTransformers.S,
		}

	default:
		return map[PrimitiveType]*mgmtv1alpha1.JobMappingTransformer{}
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

func (m *defaultTransformerProcessor) getValue(transformerKey PrimitiveType, value any, shouldMutate bool) (any, error) {
	t := m.defaultTransformersInitMap[transformerKey]
	if t != nil && shouldMutate {
		return t.mutate(value, t.opts)
	}
	return value, nil
}

type DefaultTransformer struct {
	opts   any
	mutate func(value any, opts any) (any, error)
}

// initializes transformers and options
func initTransformerOpts(transformerMapping *mgmtv1alpha1.JobMappingTransformer) (*DefaultTransformer, error) {
	switch transformerMapping.Source {
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CATEGORICAL:
		categories := transformerMapping.Config.GetGenerateCategoricalConfig().Categories
		opts, err := transformer.NewGenerateCategoricalOpts(categories)
		if err != nil {
			return nil, err
		}
		generate := transformer.NewGenerateCategorical().Generate
		return &DefaultTransformer{
			opts: opts,
			mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL:
		opts, err := transformer.NewGenerateBoolOpts(nil)
		if err != nil {
			return nil, err
		}
		generate := transformer.NewGenerateBool().Generate
		return &DefaultTransformer{
			opts: opts,
			mutate: func(value any, opts any) (any, error) {
				return generate(opts)
			},
		}, nil

	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_STRING:
		pl := transformerMapping.Config.GetTransformStringConfig().PreserveLength
		minLength := int64(3) // todo: we need to pull in this value from the database schema
		opts, err := transformer.NewTransformStringOpts(&pl, &minLength, nil)
		if err != nil {
			return nil, err
		}
		transform := transformer.NewTransformString().Transform
		return &DefaultTransformer{
			opts: opts,
			mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FULL_NAME:
		pl := transformerMapping.Config.GetTransformFullNameConfig().PreserveLength
		opts, err := transformer.NewTransformFullNameOpts(nil, &pl, nil)
		if err != nil {
			return nil, err
		}
		transform := transformer.NewTransformFullName().Transform
		return &DefaultTransformer{
			opts: opts,
			mutate: func(value any, opts any) (any, error) {
				return transform(value, opts)
			},
		}, nil
	default:
		return nil, nil
	}

}

func initDefaultTransformers(defaultTransformerMap map[PrimitiveType]*mgmtv1alpha1.JobMappingTransformer) (map[PrimitiveType]*DefaultTransformer, error) {
	transformersInit := map[PrimitiveType]*DefaultTransformer{}
	for k, t := range defaultTransformerMap {
		init, err := initTransformerOpts(t)
		if err != nil {
			return nil, err
		}
		transformersInit[k] = init
	}
	return transformersInit, nil
}

// func computeMutationFunction(value any, transformerMapping *mgmtv1alpha1.JobMappingTransformer, colInfo *sqlmanager_shared.ColumnInfo) (any, error) {
// 	var maxLen int64 = 10000
// 	if colInfo != nil && colInfo.CharacterMaximumLength != nil && *colInfo.CharacterMaximumLength > 0 {
// 		maxLen = int64(*colInfo.CharacterMaximumLength)
// 	}

// 	switch transformerMapping.Source {
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CATEGORICAL:
// 		categories := transformerMapping.Config.GetGenerateCategoricalConfig().Categories
// 		v, err := transformer.NewGenerateCategorical().Generate(&transformer.GenerateCategoricalOpts{
// 			Categories: categories,
// 		})
// 		if err != nil {
// 			return nil, err
// 		}
// 		return v, nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_EMAIL:
// 		emailType := transformerMapping.GetConfig().GetGenerateEmailConfig().GetEmailType()
// 		if emailType == mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UNSPECIFIED {
// 			emailType = mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UUID_V4
// 		}
// 		return fmt.Sprintf(`generate_email(max_length:%d,email_type:%q)`, maxLen, dtoEmailTypeToBenthosEmailType(emailType)), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_EMAIL:
// 		pd := transformerMapping.Config.GetTransformEmailConfig().PreserveDomain
// 		pl := transformerMapping.Config.GetTransformEmailConfig().PreserveLength
// 		excludedDomains := transformerMapping.Config.GetTransformEmailConfig().ExcludedDomains

// 		excludedDomainsStr, err := convertStringSliceToString(excludedDomains)
// 		if err != nil {
// 			return "", err
// 		}
// 		emailType := transformerMapping.GetConfig().GetTransformEmailConfig().GetEmailType()
// 		if emailType == mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UNSPECIFIED {
// 			emailType = mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UUID_V4
// 		}

// 		invalidEmailAction := transformerMapping.GetConfig().GetTransformEmailConfig().GetInvalidEmailAction()
// 		if invalidEmailAction == mgmtv1alpha1.InvalidEmailAction_INVALID_EMAIL_ACTION_UNSPECIFIED {
// 			invalidEmailAction = mgmtv1alpha1.InvalidEmailAction_INVALID_EMAIL_ACTION_REJECT
// 		}

// 		return fmt.Sprintf(
// 			"transform_email(value:this.%q,preserve_domain:%t,preserve_length:%t,excluded_domains:%v,max_length:%d,email_type:%q,invalid_email_action:%q)",
// 			col.Column, pd, pl, excludedDomainsStr, maxLen, dtoEmailTypeToBenthosEmailType(emailType), dtoInvalidEmailActionToBenthosInvalidEmailAction(invalidEmailAction),
// 		), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL:
// 		return "generate_bool()", nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CARD_NUMBER:
// 		luhn := transformerMapping.Config.GetGenerateCardNumberConfig().ValidLuhn
// 		return fmt.Sprintf(`generate_card_number(valid_luhn:%t)`, luhn), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CITY:
// 		return fmt.Sprintf(`generate_city(max_length:%d)`, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_E164_PHONE_NUMBER:
// 		minValue := transformerMapping.Config.GetGenerateE164PhoneNumberConfig().Min
// 		maxValue := transformerMapping.Config.GetGenerateE164PhoneNumberConfig().Max
// 		return fmt.Sprintf(`generate_e164_phone_number(min:%d,max:%d)`, minValue, maxValue), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FIRST_NAME:
// 		return fmt.Sprintf(`generate_first_name(max_length:%d)`, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FLOAT64:
// 		randomSign := transformerMapping.Config.GetGenerateFloat64Config().RandomizeSign
// 		minValue := transformerMapping.Config.GetGenerateFloat64Config().Min
// 		maxValue := transformerMapping.Config.GetGenerateFloat64Config().Max

// 		var precision *int64
// 		if transformerMapping.GetConfig().GetGenerateFloat64Config().GetPrecision() > 0 {
// 			userDefinedPrecision := transformerMapping.GetConfig().GetGenerateFloat64Config().GetPrecision()
// 			precision = &userDefinedPrecision
// 		}
// 		if colInfo != nil && colInfo.NumericPrecision != nil && *colInfo.NumericPrecision > 0 {
// 			newPrecision := transformer_utils.Ceil(*precision, int64(*colInfo.NumericPrecision))
// 			precision = &newPrecision
// 		}

// 		var scale *int64
// 		if colInfo != nil && colInfo.NumericScale != nil && *colInfo.NumericScale >= 0 {
// 			newScale := int64(*colInfo.NumericScale)
// 			scale = &newScale
// 		}

// 		fnStr := []string{"randomize_sign:%t", "min:%f", "max:%f"}
// 		params := []any{randomSign, minValue, maxValue}

// 		if precision != nil {
// 			fnStr = append(fnStr, "precision: %d")
// 			params = append(params, *precision)
// 		}
// 		if scale != nil {
// 			fnStr = append(fnStr, "scale: %d")
// 			params = append(params, *scale)
// 		}
// 		template := fmt.Sprintf("generate_float64(%s)", strings.Join(fnStr, ", "))
// 		return fmt.Sprintf(template, params...), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_ADDRESS:
// 		return fmt.Sprintf(`generate_full_address(max_length:%d)`, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME:
// 		return fmt.Sprintf(`generate_full_name(max_length:%d)`, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_GENDER:
// 		ab := transformerMapping.Config.GetGenerateGenderConfig().Abbreviate
// 		return fmt.Sprintf(`generate_gender(abbreviate:%t,max_length:%d)`, ab, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64_PHONE_NUMBER:
// 		return "generate_int64_phone_number()", nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64:
// 		sign := transformerMapping.Config.GetGenerateInt64Config().RandomizeSign
// 		minValue := transformerMapping.Config.GetGenerateInt64Config().Min
// 		maxValue := transformerMapping.Config.GetGenerateInt64Config().Max
// 		return fmt.Sprintf(`generate_int64(randomize_sign:%t,min:%d, max:%d)`, sign, minValue, maxValue), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_LAST_NAME:
// 		return fmt.Sprintf(`generate_last_name(max_length:%d)`, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SHA256HASH:
// 		return `generate_sha256hash()`, nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SSN:
// 		return "generate_ssn()", nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STATE:
// 		generateFullName := transformerMapping.Config.GetGenerateStateConfig().GenerateFullName
// 		return fmt.Sprintf(`generate_state(generate_full_name:%t)`, generateFullName), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STREET_ADDRESS:
// 		return fmt.Sprintf(`generate_street_address(max_length:%d)`, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STRING_PHONE_NUMBER:
// 		minValue := transformerMapping.Config.GetGenerateStringPhoneNumberConfig().Min
// 		maxValue := transformerMapping.Config.GetGenerateStringPhoneNumberConfig().Max
// 		minValue = transformer_utils.MinInt(minValue, maxLen)
// 		maxValue = transformer_utils.Ceil(maxValue, maxLen)
// 		return fmt.Sprintf("generate_string_phone_number(min:%d,max:%d)", minValue, maxValue), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_RANDOM_STRING:
// 		minValue := transformerMapping.Config.GetGenerateStringConfig().Min
// 		maxValue := transformerMapping.Config.GetGenerateStringConfig().Max
// 		minValue = transformer_utils.MinInt(minValue, maxLen) // ensure the min is not larger than the max allowed length
// 		maxValue = transformer_utils.Ceil(maxValue, maxLen)
// 		// todo: we need to pull in the min from the database schema
// 		return fmt.Sprintf(`generate_string(min:%d,max:%d)`, minValue, maxValue), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UNIXTIMESTAMP:
// 		return "generate_unixtimestamp()", nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_USERNAME:
// 		return fmt.Sprintf(`generate_username(max_length:%d)`, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UTCTIMESTAMP:
// 		return "generate_utctimestamp()", nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UUID:
// 		ih := transformerMapping.Config.GetGenerateUuidConfig().IncludeHyphens
// 		return fmt.Sprintf("generate_uuid(include_hyphens:%t)", ih), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_ZIPCODE:
// 		return "generate_zipcode()", nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_E164_PHONE_NUMBER:
// 		pl := transformerMapping.Config.GetTransformE164PhoneNumberConfig().PreserveLength
// 		return fmt.Sprintf("transform_e164_phone_number(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FIRST_NAME:
// 		pl := transformerMapping.Config.GetTransformFirstNameConfig().PreserveLength
// 		return fmt.Sprintf("transform_first_name(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FLOAT64:
// 		rMin := transformerMapping.Config.GetTransformFloat64Config().RandomizationRangeMin
// 		rMax := transformerMapping.Config.GetTransformFloat64Config().RandomizationRangeMax

// 		var precision *int64
// 		if colInfo != nil && colInfo.NumericPrecision != nil && *colInfo.NumericPrecision > 0 {
// 			newPrecision := int64(*colInfo.NumericPrecision)
// 			precision = &newPrecision
// 		}

// 		var scale *int64
// 		if colInfo != nil && colInfo.NumericScale != nil && *colInfo.NumericScale >= 0 {
// 			newScale := int64(*colInfo.NumericScale)
// 			scale = &newScale
// 		}

// 		fnStr := []string{"value:this.%q", "randomization_range_min:%f", "randomization_range_max:%f"}
// 		params := []any{col.Column, rMin, rMax}

// 		if precision != nil {
// 			fnStr = append(fnStr, "precision:%d")
// 			params = append(params, *precision)
// 		}
// 		if scale != nil {
// 			fnStr = append(fnStr, "scale:%d")
// 			params = append(params, *scale)
// 		}
// 		template := fmt.Sprintf(`transform_float64(%s)`, strings.Join(fnStr, ", "))
// 		return fmt.Sprintf(template, params...), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FULL_NAME:
// 		pl := transformerMapping.Config.GetTransformFullNameConfig().PreserveLength
// 		return fmt.Sprintf("transform_full_name(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64_PHONE_NUMBER:
// 		pl := transformerMapping.Config.GetTransformInt64PhoneNumberConfig().PreserveLength
// 		return fmt.Sprintf("transform_int64_phone_number(value:this.%q,preserve_length:%t)", col.Column, pl), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64:
// 		rMin := transformerMapping.Config.GetTransformInt64Config().RandomizationRangeMin
// 		rMax := transformerMapping.Config.GetTransformInt64Config().RandomizationRangeMax
// 		return fmt.Sprintf(`transform_int64(value:this.%q,randomization_range_min:%d,randomization_range_max:%d)`, col.Column, rMin, rMax), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_LAST_NAME:
// 		pl := transformerMapping.Config.GetTransformLastNameConfig().PreserveLength
// 		return fmt.Sprintf("transform_last_name(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_PHONE_NUMBER:
// 		pl := transformerMapping.Config.GetTransformPhoneNumberConfig().PreserveLength
// 		return fmt.Sprintf("transform_phone_number(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_STRING:
// 		pl := transformerMapping.Config.GetTransformStringConfig().PreserveLength
// 		minLength := int64(3) // todo: we need to pull in this value from the database schema
// 		return fmt.Sprintf(`transform_string(value:this.%q,preserve_length:%t,min_length:%d,max_length:%d)`, col.Column, pl, minLength, maxLen), nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL:
// 		return shared.NullString, nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT:
// 		return `"DEFAULT"`, nil
// 	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_CHARACTER_SCRAMBLE:
// 		regex := transformerMapping.Config.GetTransformCharacterScrambleConfig().UserProvidedRegex

// 		if regex != nil {
// 			regexValue := *regex
// 			return fmt.Sprintf(`transform_character_scramble(value:this.%q,user_provided_regex:%q)`, col.Column, regexValue), nil
// 		} else {
// 			return fmt.Sprintf(`transform_character_scramble(value:this.%q)`, col.Column), nil
// 		}

// 	default:
// 		return "", fmt.Errorf("unsupported transformer")
// 	}
// }

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
