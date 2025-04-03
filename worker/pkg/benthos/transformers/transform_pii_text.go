package transformers

import (
	context "context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// Minimal interface that includes the config and value
// To be used deep in the transformers so we don't have to be aware of the account id at the benthos level
type TransformPiiTextApi interface {
	Transform(ctx context.Context, config *mgmtv1alpha1.TransformPiiText, value string) (string, error)
}

// Full interface that includes the account id
type AccountTransformPiiTextApi interface {
	Transform(ctx context.Context, accountId string, config *mgmtv1alpha1.TransformPiiText, value string) (string, error)
}

type AccountAwareAnonymizationPiiTextApi struct {
	anonApi   mgmtv1alpha1connect.AnonymizationServiceClient
	accountId string
}

func NewAccountAwareAnonymizationPiiTextApi(
	anonApi mgmtv1alpha1connect.AnonymizationServiceClient,
	accountId string,
) *AccountAwareAnonymizationPiiTextApi {
	return &AccountAwareAnonymizationPiiTextApi{
		anonApi:   anonApi,
		accountId: accountId,
	}
}

func (a *AccountAwareAnonymizationPiiTextApi) Transform(
	ctx context.Context,
	config *mgmtv1alpha1.TransformPiiText,
	value string,
) (string, error) {
	wrapper := valueWrapper{
		Input: value,
	}
	bits, err := json.Marshal(wrapper)
	if err != nil {
		return "", fmt.Errorf("unable to marshal value: %w", err)
	}

	resp, err := a.anonApi.AnonymizeSingle(ctx, connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
		InputData: string(bits),
		AccountId: a.accountId,
		TransformerMappings: []*mgmtv1alpha1.TransformerMapping{
			{
				Expression: ".input",
				Transformer: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{
						TransformPiiTextConfig: config,
					},
				},
			},
		},
	}))

	if err != nil {
		return "", fmt.Errorf("unable to anonymize text: %w", err)
	}

	outputData := resp.Msg.GetOutputData()
	var output valueWrapper
	err = json.Unmarshal([]byte(outputData), &output)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal output: %w", err)
	}

	return output.Input, nil
}

type valueWrapper struct {
	Input string `json:"input"`
}

func RegisterTransformPiiText(
	env *bloblang.Environment,
	api TransformPiiTextApi,
) error {
	spec := bloblang.NewPluginSpec().
		Description("Anonymizes and transforms freeform text.").
		Category("string").
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewFloat64Param("score_threshold").
			Default(0.5).
			Optional().
			Description("The minimum score for a text to be considered PII."),
		).
		Param(bloblang.NewStringParam("language").
			Optional().
			Default("en").
			Description("The language of the text to be anonymized."),
		).
		Param(bloblang.NewAnyParam("allowed_phrases").
			Optional().
			Default([]any{}).
			Description("A list of phrases that will not be considered PII."),
		).
		Param(bloblang.NewAnyParam("allowed_entities").
			Optional().
			Default([]any{}).
			Description("A list of entities to be used for PII analysis. If not provided or empty, all entities are considered. If specified, any ad-hoc, or deny_recognizers entity names must also be provided. To see available builtin entities, cal the GetPiiTextEntities() RPC method for your account."),
		).
		Param(bloblang.NewAnyParam("default_anonymizer").
			Optional().
			Description("The default anonymization configuration used for all instances of detected PII."),
		).
		Param(bloblang.NewAnyParam("deny_recognizers").
			Optional().
			Default([]any{}).
			Description("Configure deny lists where each word is treated as PII. Each entry should contain 'name' and 'deny_words' fields."),
		).
		Param(bloblang.NewAnyParam("entity_anonymizers").
			Optional().
			Default(map[string]any{}).
			Description("A map of entity names to anonymizer configurations. The key corresponds to a recognized entity (e.g. PERSON, PHONE_NUMBER) and the value is the anonymizer configuration."),
		)

	err := env.RegisterFunctionV2(
		"transform_pii_text",
		spec,
		func(args *bloblang.ParsedParams) (bloblang.Function, error) {
			valuePtr, err := args.GetOptionalString("value")
			if err != nil {
				return nil, err
			}

			scoreThresholdParam, err := args.GetOptionalFloat64("score_threshold")
			if err != nil {
				return nil, err
			}
			scoreThreshold := float32(0.5)
			if scoreThresholdParam != nil {
				scoreThreshold = float32(*scoreThresholdParam)
			}

			language, err := args.GetOptionalString("language")
			if err != nil {
				return nil, err
			}
			if language == nil {
				defaultLanguage := "en"
				language = &defaultLanguage
			}

			allowedPhrasesParam, err := args.Get("allowed_phrases")
			if err != nil {
				return nil, err
			}
			allowedPhrases, err := fromAnyToStringSlice(allowedPhrasesParam)
			if err != nil {
				return nil, err
			}

			allowedEntitiesParam, err := args.Get("allowed_entities")
			if err != nil {
				return nil, err
			}
			allowedEntities, err := fromAnyToStringSlice(allowedEntitiesParam)
			if err != nil {
				return nil, err
			}

			defaultAnonymizer, err := args.Get("default_anonymizer")
			if err != nil {
				return nil, err
			}
			// Convert to PiiAnonymizer struct
			var defaultAnonymizerConfig *mgmtv1alpha1.PiiAnonymizer
			if defaultAnonymizer != nil {
				defaultAnonymizerConfig, err = convertToPiiAnonymizer(defaultAnonymizer)
				if err != nil {
					return nil, fmt.Errorf("invalid default_anonymizer config: %w", err)
				}
			}

			denyRecognizersRaw, err := args.Get("deny_recognizers")
			if err != nil {
				return nil, err
			}
			// Convert to PiiDenyRecognizer array
			denyRecognizers, err := convertToPiiDenyRecognizerArray(denyRecognizersRaw)
			if err != nil {
				return nil, fmt.Errorf("invalid deny_recognizers config: %w", err)
			}

			entityAnonymizersRaw, err := args.Get("entity_anonymizers")
			if err != nil {
				return nil, err
			}
			// Convert to map[string]PiiAnonymizer
			entityAnonymizers, err := convertToPiiAnonymizerMap(entityAnonymizersRaw)
			if err != nil {
				return nil, fmt.Errorf("invalid entity_anonymizers config: %w", err)
			}

			config := &mgmtv1alpha1.TransformPiiText{
				ScoreThreshold:    float32(scoreThreshold),
				Language:          language,
				AllowedPhrases:    allowedPhrases,
				AllowedEntities:   allowedEntities,
				DefaultAnonymizer: defaultAnonymizerConfig,
				DenyRecognizers:   denyRecognizers,
				EntityAnonymizers: entityAnonymizers,
			}

			return func() (any, error) {
				res, err := transformPiiText(api, config, valuePtr)
				if err != nil {
					return nil, fmt.Errorf("unable to run transform_pii_text: %w", err)
				}
				return res, nil
			}, nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to register transform_pii_text: %w", err)
	}
	return nil
}

func NewTransformPiiTextOptsFromConfig(
	config *mgmtv1alpha1.TransformPiiText,
) (*TransformPiiTextOpts, error) {
	if config == nil {
		defaultLanguage := "en"
		config = &mgmtv1alpha1.TransformPiiText{
			Language:       &defaultLanguage,
			ScoreThreshold: 0.5,
		}
	}
	scoreThreshold := float64(config.ScoreThreshold)
	return NewTransformPiiTextOpts(
		&scoreThreshold,
		config.Language,
		config.AllowedPhrases,
		config.AllowedEntities,
		config.DefaultAnonymizer,
		config.DenyRecognizers,
		config.EntityAnonymizers,
	)
}

func (t *TransformPiiText) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformPiiTextOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	allowedPhrases, err := fromAnyToStringSlice(parsedOpts.allowedPhrases)
	if err != nil {
		return nil, fmt.Errorf("invalid allowed_phrases: %w", err)
	}

	allowedEntities, err := fromAnyToStringSlice(parsedOpts.allowedEntities)
	if err != nil {
		return nil, fmt.Errorf("invalid allowed_entities: %w", err)
	}

	defaultAnonymizer, err := convertToPiiAnonymizer(parsedOpts.defaultAnonymizer)
	if err != nil {
		return nil, fmt.Errorf("invalid default_anonymizer: %w", err)
	}

	denyRecognizers, err := convertToPiiDenyRecognizerArray(parsedOpts.denyRecognizers)
	if err != nil {
		return nil, fmt.Errorf("invalid deny_recognizers: %w", err)
	}

	entityAnonymizers, err := convertToPiiAnonymizerMap(parsedOpts.entityAnonymizers)
	if err != nil {
		return nil, fmt.Errorf("invalid entity_anonymizers: %w", err)
	}

	config := &mgmtv1alpha1.TransformPiiText{
		ScoreThreshold:    float32(*parsedOpts.scoreThreshold),
		Language:          parsedOpts.language,
		AllowedPhrases:    allowedPhrases,
		AllowedEntities:   allowedEntities,
		DefaultAnonymizer: defaultAnonymizer,
		DenyRecognizers:   denyRecognizers,
		EntityAnonymizers: entityAnonymizers,
	}

	return transformPiiText(t.api, config, &valueStr)
}

func transformPiiText(api TransformPiiTextApi, config *mgmtv1alpha1.TransformPiiText, value any) (*string, error) {
	if value == nil {
		return nil, nil
	}

	v := reflect.ValueOf(value)
	var result string
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return nil, nil
		}
		result = v.Elem().String()
	case reflect.String:
		result = v.String()
	default:
		result = v.String()
	}

	if result == "" {
		return &result, nil
	}

	transformedResult, err := api.Transform(context.Background(), config, result)
	if err != nil {
		return nil, fmt.Errorf("unable to transform PII text: %w", err)
	}

	return &transformedResult, nil
}

func convertToPiiDenyRecognizerArray(raw any) ([]*mgmtv1alpha1.PiiDenyRecognizer, error) {
	denyRecognizers := make([]*mgmtv1alpha1.PiiDenyRecognizer, 0)
	if raw == nil {
		return denyRecognizers, nil
	}
	denyRecognizersRawArray, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("deny_recognizers must be an array")
	}
	for _, recognizer := range denyRecognizersRawArray {
		recognizerMap, ok := recognizer.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("deny_recognizer must be a map, was: %T", recognizer)
		}
		denyRecognizer, err := convertToPiiDenyRecognizer(recognizerMap)
		if err != nil {
			return nil, fmt.Errorf("invalid deny_recognizer config: %w", err)
		}
		denyRecognizers = append(denyRecognizers, denyRecognizer)
	}
	return denyRecognizers, nil
}

func convertToPiiAnonymizerMap(raw any) (map[string]*mgmtv1alpha1.PiiAnonymizer, error) {
	entityAnonymizers := make(map[string]*mgmtv1alpha1.PiiAnonymizer)
	if raw == nil {
		return entityAnonymizers, nil
	}
	entityAnonymizersRawMap, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("entity_anonymizers must be a map, was: %T", raw)
	}
	for entity, anonymizer := range entityAnonymizersRawMap {
		anonymizerConfig, err := convertToPiiAnonymizer(anonymizer)
		if err != nil {
			return nil, fmt.Errorf("invalid entity_anonymizer config for entity %s: %w", entity, err)
		}
		entityAnonymizers[entity] = anonymizerConfig
	}
	return entityAnonymizers, nil
}

func convertToPiiAnonymizer(raw any) (*mgmtv1alpha1.PiiAnonymizer, error) {
	if raw == nil {
		return nil, nil
	}

	configMap, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("anonymizer config must be a map")
	}

	anonymizer := &mgmtv1alpha1.PiiAnonymizer{}

	// Check for each possible config type and set accordingly
	if replace, ok := configMap["replace"].(map[string]any); ok {
		var value *string
		valueParam, ok := replace["value"].(string)
		if ok && valueParam != "" {
			value = &valueParam
		}
		anonymizer.Config = &mgmtv1alpha1.PiiAnonymizer_Replace_{
			Replace: &mgmtv1alpha1.PiiAnonymizer_Replace{
				Value: value,
			},
		}
	} else if _, ok := configMap["redact"].(map[string]any); ok {
		anonymizer.Config = &mgmtv1alpha1.PiiAnonymizer_Redact_{
			Redact: &mgmtv1alpha1.PiiAnonymizer_Redact{},
		}
	} else if mask, ok := configMap["mask"].(map[string]any); ok {
		maskConfig := &mgmtv1alpha1.PiiAnonymizer_Mask{}
		if char, ok := mask["masking_char"].(string); ok {
			maskConfig.MaskingChar = &char
		}
		if chars, ok := mask["chars_to_mask"].(float64); ok {
			intChars := int32(chars)
			maskConfig.CharsToMask = &intChars
		}
		if fromEnd, ok := mask["from_end"].(bool); ok {
			maskConfig.FromEnd = &fromEnd
		}
		anonymizer.Config = &mgmtv1alpha1.PiiAnonymizer_Mask_{
			Mask: &mgmtv1alpha1.PiiAnonymizer_Mask{
				MaskingChar: maskConfig.MaskingChar,
				CharsToMask: maskConfig.CharsToMask,
				FromEnd:     maskConfig.FromEnd,
			},
		}
	} else if hash, ok := configMap["hash"].(map[string]any); ok {
		if algo, ok := hash["algo"].(int64); ok {
			convertedAlgo := mgmtv1alpha1.PiiAnonymizer_Hash_HashType(algo) //nolint:gosec
			if _, ok := mgmtv1alpha1.PiiAnonymizer_Hash_HashType_name[int32(convertedAlgo)]; !ok {
				return nil, fmt.Errorf("invalid hash algorithm: %d", convertedAlgo)
			}
			anonymizer.Config = &mgmtv1alpha1.PiiAnonymizer_Hash_{
				Hash: &mgmtv1alpha1.PiiAnonymizer_Hash{
					Algo: &convertedAlgo,
				},
			}
		} else {
			return nil, fmt.Errorf("invalid hash algorithm: %T", hash["algo"])
		}
	} else if _, ok := configMap["transform"].(map[string]any); ok {
		return nil, fmt.Errorf("transform not currently supported")
	} else {
		return nil, fmt.Errorf("invalid anonymizer config: must contain one of replace, redact, mask, hash, or transform")
	}

	return anonymizer, nil
}

func convertToPiiDenyRecognizer(raw map[string]any) (*mgmtv1alpha1.PiiDenyRecognizer, error) {
	name, ok := raw["name"].(string)
	if !ok {
		return nil, fmt.Errorf("deny_recognizer must have a name")
	}

	denyWordsRaw, ok := raw["deny_words"].([]any)
	if !ok {
		return nil, fmt.Errorf("deny_recognizer must have deny_words array")
	}

	denyWords := make([]string, 0)
	for _, word := range denyWordsRaw {
		if str, ok := word.(string); ok {
			denyWords = append(denyWords, str)
		}
	}

	return &mgmtv1alpha1.PiiDenyRecognizer{
		Name:      name,
		DenyWords: denyWords,
	}, nil
}
