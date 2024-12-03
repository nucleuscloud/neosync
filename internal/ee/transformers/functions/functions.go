package ee_transformer_fns

import (
	"context"
	"fmt"

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

	defaultAnon, ok, err := getDefaultAnonymizer(config.GetDefaultAnonymizer())
	if err != nil {
		return "", fmt.Errorf("unable to build default anonymizer: %w", err)
	}
	anonymizers := map[string]presidioapi.AnonymizeRequest_Anonymizers_AdditionalProperties{}
	if ok {
		anonymizers["DEFAULT"] = *defaultAnon
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
	return *anonResp.JSON200.Text, nil
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

func getDefaultAnonymizer(dto *mgmtv1alpha1.PiiAnonymizer) (*presidioapi.AnonymizeRequest_Anonymizers_AdditionalProperties, bool, error) {
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
