package ee_transformer_fns

import (
	"context"
	"fmt"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
)

var (
	supportedLanguage = "en"
)

func TransformPiiText(
	ctx context.Context,
	analyzeClient presidioapi.AnalyzeInterface,
	anonymizeClient presidioapi.AnonymizeInterface,
	config *mgmtv1alpha1.TransformPiiText,
	value string,
) (string, error) {
	if value == "" {
		return value, nil
	}
	threshold := float64(config.GetScoreThreshold())
	adhocRecognizers := buildAdhocRecognizers(config.GetDenyRecognizers())
	allowedEntities := config.GetAllowedEntities()
	analyzeResp, err := analyzeClient.PostAnalyzeWithResponse(ctx, presidioapi.AnalyzeRequest{
		Text:             value,
		Language:         config.GetLanguage(),
		ScoreThreshold:   &threshold,
		AdHocRecognizers: &adhocRecognizers,
		Entities:         &allowedEntities,
	})
	if err != nil {
		return "", fmt.Errorf("unable to analyze input: %w", err)
	}

	if analyzeResp.JSON200 == nil {
		return "", fmt.Errorf("received non-200 response from analyzer: %s %d %s", analyzeResp.Status(), analyzeResp.StatusCode(), string(analyzeResp.Body))
	}

	analysisResults := removeAllowedPhrases(*analyzeResp.JSON200, value, config.GetAllowedPhrases())

	analysisResults, hasNeosyncEntities := processAnalysisResultsForNeosyncTransformers(analysisResults, getNeosyncConfiguredEntities(config))

	// if neosync analyzer is configured for a specific entity, do the replacement and pop the entity from the analysis results

	anonymizers, err := buildAnonymizers(config)
	if err != nil {
		return "", fmt.Errorf("unable to build anonymizers: %w", err)
	}

	anonResp, err := anonymizeClient.PostAnonymizeWithResponse(ctx, presidioapi.AnonymizeRequest{
		AnalyzerResults: presidioapi.ToAnonymizeRecognizerResults(analysisResults),
		Text:            value,
		Anonymizers:     &anonymizers,
	})
	if err != nil {
		return "", fmt.Errorf("unable to anonymize input: %w", err)
	}
	err = handleAnonRespErr(anonResp)
	if err != nil {
		return "", err
	}
	output := *anonResp.JSON200.Text

	if hasNeosyncEntities {
		// do another pass to anonymize the neosync entities
		for _, item := range *anonResp.JSON200.Items {
			if strings.HasPrefix(item.EntityType, neosyncEntityPrefix) {
				// do the replacement
				presidioEntity := strings.TrimPrefix(item.EntityType, neosyncEntityPrefix)
				_ = presidioEntity
				// find transformer config from map
				// call function to get transformed data
				transformedSnippet := ""
				output = strings.ReplaceAll(output, item.EntityType, transformedSnippet)
			}
		}
	}

	return output, nil
}

func getNeosyncConfiguredEntities(config *mgmtv1alpha1.TransformPiiText) []string {
	entities := []string{}
	for entity := range config.GetEntityAnonymizers() {
		entities = append(entities, entity)
	}
	return entities
}

func buildAnonymizers(config *mgmtv1alpha1.TransformPiiText) (map[string]presidioapi.AnonymizeRequest_Anonymizers_AdditionalProperties, error) {
	output := map[string]presidioapi.AnonymizeRequest_Anonymizers_AdditionalProperties{}
	defaultAnon, ok, err := toPresidioAnonymizerConfig("", config.GetDefaultAnonymizer())
	if err != nil {
		return nil, fmt.Errorf("unable to build default anonymizer: %w", err)
	}
	if ok {
		output["DEFAULT"] = *defaultAnon
	}
	for entity, anonymizer := range config.GetEntityAnonymizers() {
		ap, ok, err := toPresidioAnonymizerConfig(entity, anonymizer)
		if err != nil {
			return nil, fmt.Errorf("unable to build entity %s anonymizer: %w", entity, err)
		}
		if ok {
			output[entity] = *ap
		}
	}

	return output, nil
}

func removeAllowedPhrases(
	results []presidioapi.RecognizerResultWithAnaysisExplanation,
	text string,
	allowedPhrases []string,
) []presidioapi.RecognizerResultWithAnaysisExplanation {
	output := []presidioapi.RecognizerResultWithAnaysisExplanation{}
	uniquePhrases := transformer_utils.ToSet(allowedPhrases)
	textLen := len(text)
	for _, result := range results {
		if result.Start < 0 || result.End > textLen {
			continue // Skip invalid ranges
		}

		phrase := text[result.Start:result.End]
		if _, ok := uniquePhrases[phrase]; !ok {
			output = append(output, result)
		}
	}

	return output
}

// type neosyncAnalysisResult struct {
// 	NeosyncAnalyzerResults  []presidioapi.RecognizerResultWithAnaysisExplanation
// 	PresidioAnalyzerResults []presidioapi.RecognizerResultWithAnaysisExplanation
// }

const (
	neosyncEntityPrefix = "NEOSYNC_"
)

func processAnalysisResultsForNeosyncTransformers(
	inputResults []presidioapi.RecognizerResultWithAnaysisExplanation,
	neosyncEnabledEntities []string,
) ([]presidioapi.RecognizerResultWithAnaysisExplanation, bool) {
	entitySet := map[string]struct{}{}
	for _, entity := range neosyncEnabledEntities {
		entitySet[entity] = struct{}{}
	}

	output := make([]presidioapi.RecognizerResultWithAnaysisExplanation, 0, len(inputResults))
	hasNeosyncEntities := false
	for _, result := range inputResults {
		if _, ok := entitySet[result.EntityType]; ok {
			result.EntityType = fmt.Sprintf("%s%s", neosyncEntityPrefix, result.EntityType)
			hasNeosyncEntities = true
		}
		output = append(output, result)
	}

	return output, hasNeosyncEntities
}

func buildAdhocRecognizers(dtos []*mgmtv1alpha1.PiiDenyRecognizer) []presidioapi.PatternRecognizer {
	output := []presidioapi.PatternRecognizer{}
	for _, dto := range dtos {
		name := dto.GetName()
		denywords := dto.GetDenyWords()
		output = append(output, presidioapi.PatternRecognizer{
			Name:              &name,
			SupportedEntity:   &name,
			DenyList:          &denywords,
			SupportedLanguage: &supportedLanguage,
		})
	}
	return output
}

func toPresidioAnonymizerConfig(entity string, dto *mgmtv1alpha1.PiiAnonymizer) (*presidioapi.AnonymizeRequest_Anonymizers_AdditionalProperties, bool, error) {
	switch cfg := dto.GetConfig().(type) {
	case *mgmtv1alpha1.PiiAnonymizer_Redact_:
		ap := &presidioapi.AnonymizeRequest_Anonymizers_AdditionalProperties{}
		err := ap.FromRedact(presidioapi.Redact{Type: "redact"})
		if err != nil {
			return nil, false, err
		}
		return ap, true, nil
	case *mgmtv1alpha1.PiiAnonymizer_Replace_:
		ap := &presidioapi.AnonymizeRequest_Anonymizers_AdditionalProperties{}
		err := ap.FromReplace(presidioapi.Replace{Type: "replace", NewValue: cfg.Replace.GetValue()})
		if err != nil {
			return nil, false, err
		}
		return ap, true, nil
	case *mgmtv1alpha1.PiiAnonymizer_Hash_:
		ap := &presidioapi.AnonymizeRequest_Anonymizers_AdditionalProperties{}
		hashtype := toPresidioHashType(cfg.Hash.GetAlgo())
		err := ap.FromHash(presidioapi.Hash{Type: "hash", HashType: &hashtype})
		if err != nil {
			return nil, false, err
		}
		return ap, true, nil
	case *mgmtv1alpha1.PiiAnonymizer_Mask_:
		ap := &presidioapi.AnonymizeRequest_Anonymizers_AdditionalProperties{}
		fromend := cfg.Mask.GetFromEnd()
		err := ap.FromMask(presidioapi.Mask{
			Type:        "mask",
			CharsToMask: int(cfg.Mask.GetCharsToMask()),
			FromEnd:     &fromend,
			MaskingChar: cfg.Mask.GetMaskingChar(),
		})
		if err != nil {
			return nil, false, err
		}
		return ap, true, nil
	case *mgmtv1alpha1.PiiAnonymizer_Transform_:
		ap := &presidioapi.AnonymizeRequest_Anonymizers_AdditionalProperties{}
		err := ap.FromReplace(presidioapi.Replace{Type: "replace", NewValue: fmt.Sprintf("{{%s%s}}", neosyncEntityPrefix, entity)})
		if err != nil {
			return nil, false, err
		}
		return ap, true, nil
	}
	return nil, false, nil
}

func toPresidioHashType(dto mgmtv1alpha1.PiiAnonymizer_Hash_HashType) presidioapi.HashHashType {
	switch dto {
	case mgmtv1alpha1.PiiAnonymizer_Hash_HASH_TYPE_MD5:
		return presidioapi.Md5
	case mgmtv1alpha1.PiiAnonymizer_Hash_HASH_TYPE_SHA256:
		return presidioapi.Sha256
	case mgmtv1alpha1.PiiAnonymizer_Hash_HASH_TYPE_SHA512:
		return presidioapi.Sha512
	default:
		return presidioapi.Md5
	}
}

func handleAnonRespErr(resp *presidioapi.PostAnonymizeResponse) error {
	if resp == nil {
		return fmt.Errorf("resp was nil")
	}
	if resp.JSON400 != nil {
		return fmt.Errorf("%s", *resp.JSON400.Error)
	}
	if resp.JSON422 != nil {
		return fmt.Errorf("%s", *resp.JSON422.Error)
	}
	if resp.JSON200 == nil {
		return fmt.Errorf("received non-200 response from anonymizer: %s %d %s", resp.Status(), resp.StatusCode(), string(resp.Body))
	}
	return nil
}
