// source: transform_pii_text.go

package transformers

import (
	"encoding/json"
	"fmt"
	"strings"
)

type TransformPiiText struct {
	api TransformPiiTextApi
}

type TransformPiiTextOpts struct {
	scoreThreshold    *float64
	language          *string
	allowedPhrases    any
	allowedEntities   any
	defaultAnonymizer any
	denyRecognizers   any
	entityAnonymizers any
}

func NewTransformPiiText(
	api TransformPiiTextApi,
) *TransformPiiText {
	return &TransformPiiText{
		api: api,
	}
}

func NewTransformPiiTextOpts(
	scoreThresholdArg *float64,
	language *string,
	allowedPhrasesArg any,
	allowedEntitiesArg any,
	defaultAnonymizerArg any,
	denyRecognizersArg any,
	entityAnonymizersArg any,
) (*TransformPiiTextOpts, error) {
	scoreThreshold := float64(0.5)
	if scoreThresholdArg != nil {
		scoreThreshold = *scoreThresholdArg
	}

	return &TransformPiiTextOpts{
		scoreThreshold:    &scoreThreshold,
		language:          language,
		allowedPhrases:    allowedPhrasesArg,
		allowedEntities:   allowedEntitiesArg,
		defaultAnonymizer: defaultAnonymizerArg,
		denyRecognizers:   denyRecognizersArg,
		entityAnonymizers: entityAnonymizersArg,
	}, nil
}

func (o *TransformPiiTextOpts) BuildBloblangString(
	valuePath string,
) (string, error) {
	fnStr := []string{
		"value:this.%s",
	}

	params := []any{
		valuePath,
	}

	if o.scoreThreshold != nil {
		fnStr = append(fnStr, "score_threshold:%v")
		params = append(params, *o.scoreThreshold)
	}
	if o.language != nil {
		fnStr = append(fnStr, "language:%q")
		params = append(params, *o.language)
	}
	if o.allowedPhrases != nil {
		fnStr = append(fnStr, "allowed_phrases:%v")
		params = append(params, o.allowedPhrases)
	}
	if o.allowedEntities != nil {
		fnStr = append(fnStr, "allowed_entities:%v")
		params = append(params, o.allowedEntities)
	}
	if o.defaultAnonymizer != nil {
		fnStr = append(fnStr, "default_anonymizer:%s")
		json, err := json.Marshal(o.defaultAnonymizer)
		if err != nil {
			return "", fmt.Errorf("unable to marshal default_anonymizer: %w", err)
		}
		params = append(params, string(json))
	}
	if o.denyRecognizers != nil {
		fnStr = append(fnStr, "deny_recognizers:%s")
		json, err := json.Marshal(o.denyRecognizers)
		if err != nil {
			return "", fmt.Errorf("unable to marshal deny_recognizers: %w", err)
		}
		params = append(params, string(json))
	}
	if o.entityAnonymizers != nil {
		fnStr = append(fnStr, "entity_anonymizers:%s")
		json, err := json.Marshal(o.entityAnonymizers)
		if err != nil {
			return "", fmt.Errorf("unable to marshal entity_anonymizers: %w", err)
		}
		params = append(params, string(json))
	}

	template := fmt.Sprintf("transform_pii_text(%s)", strings.Join(fnStr, ","))
	return fmt.Sprintf(template, params...), nil
}

func (t *TransformPiiText) GetJsTemplateData() (*TemplateData, error) {
	return &TemplateData{
		Name:        "transformPiiText",
		Description: "Anonymizes and transforms freeform text.",
		Example:     "",
	}, nil
}

func (t *TransformPiiText) ParseOptions(opts map[string]any) (any, error) {
	transformerOpts := &TransformPiiTextOpts{}

	scoreThreshold, ok := opts["scoreThreshold"].(float64)
	if !ok {
		scoreThreshold = 0.5
	}
	transformerOpts.scoreThreshold = &scoreThreshold

	var language *string
	if arg, ok := opts["language"].(string); ok {
		language = &arg
	}
	transformerOpts.language = language

	allowedPhrases, ok := opts["allowedPhrases"].(any)
	if !ok {
		allowedPhrases = []any{}
	}
	transformerOpts.allowedPhrases = allowedPhrases

	allowedEntities, ok := opts["allowedEntities"].(any)
	if !ok {
		allowedEntities = []any{}
	}
	transformerOpts.allowedEntities = allowedEntities

	var defaultAnonymizer any
	if arg, ok := opts["defaultAnonymizer"].(any); ok {
		defaultAnonymizer = arg
	}
	transformerOpts.defaultAnonymizer = defaultAnonymizer

	denyRecognizers, ok := opts["denyRecognizers"].(any)
	if !ok {
		denyRecognizers = []any{}
	}
	transformerOpts.denyRecognizers = denyRecognizers

	entityAnonymizers, ok := opts["entityAnonymizers"].(any)
	if !ok {
		entityAnonymizers = map[string]any{}
	}
	transformerOpts.entityAnonymizers = entityAnonymizers

	return transformerOpts, nil
}
