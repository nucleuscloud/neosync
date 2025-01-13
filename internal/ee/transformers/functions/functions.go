package ee_transformer_fns

import (
	"context"
	"fmt"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	"github.com/nucleuscloud/neosync/internal/queue"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
)

var (
	supportedLanguage = "en"
)

type NeosyncOperatorApi interface {
	Transform(ctx context.Context, config *mgmtv1alpha1.TransformerConfig, value string) (string, error)
}

func TransformPiiText(
	ctx context.Context,
	analyzeClient presidioapi.AnalyzeInterface,
	anonymizeClient presidioapi.AnonymizeInterface,
	neosyncOperatorApi NeosyncOperatorApi,
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

	analysisResults, neosyncEntityMap := processAnalysisResultsForNeosyncTransformers(analysisResults, getNeosyncConfiguredEntities(config), value)

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
	if len(neosyncEntityMap) == 0 {
		return *anonResp.JSON200.Text, nil
	}

	return handleNeosyncEntityAnonymization(
		ctx,
		anonResp.JSON200,
		config.GetEntityAnonymizers(),
		neosyncEntityMap,
		neosyncOperatorApi,
	)
}

func handleNeosyncEntityAnonymization(
	ctx context.Context,
	resp *presidioapi.AnonymizeResponse,
	entityAnonymizerMap map[string]*mgmtv1alpha1.PiiAnonymizer,
	neosyncEntityMap map[string]*queue.Queue[string],
	neosyncOperatorApi NeosyncOperatorApi,
) (string, error) {
	outputText := *resp.Text

	entityConfigMap := map[string]*mgmtv1alpha1.TransformerConfig{}
	for entity, config := range entityAnonymizerMap {
		transformConfig := config.GetTransform().GetConfig()
		if transformConfig == nil {
			transformConfig = getDefaultTransformerConfigByEntity(entity)
		}
		entityConfigMap[entity] = transformConfig
	}

	for _, item := range *resp.Items {
		if !strings.HasPrefix(item.EntityType, neosyncEntityPrefix) {
			// expected as not all items may contain neosync entities
			continue
		}

		presidioEntity := strings.TrimPrefix(item.EntityType, neosyncEntityPrefix)
		transformerConfig, ok := entityConfigMap[presidioEntity]
		if !ok {
			// todo: log mismatch between presidio entity and neosync entity
			continue
		}

		valueQueue, ok := neosyncEntityMap[item.EntityType]
		if !ok {
			// todo: log mismatch between presidio entity and neosync entity. Found neosync entity but no original values
			continue
		}
		originalValue, err := valueQueue.Dequeue()
		if err != nil {
			// todo: log mismatch between presidio entity and neosync entity. Each json200.item should correlate to a unique neosync entity and original value
			continue
		}
		transformedSnippet, err := neosyncOperatorApi.Transform(ctx, transformerConfig, originalValue)
		if err != nil {
			return "", fmt.Errorf("unable to transform neosync entity %s: %w", presidioEntity, err)
		}
		outputText = strings.Replace(outputText, *item.Text, transformedSnippet, 1)
	}

	return outputText, nil
}

func getDefaultTransformerConfigByEntity(entity string) *mgmtv1alpha1.TransformerConfig {
	switch entity {
	// case "IN_PASSPORT":
	// case "ES_NIF":
	// case "AU_TFN":
	// case "ES_NIE":
	// case "MEDICAL_LICENSE":
	// case "AU_MEDICARE":
	// case "IN_AADHAAR":
	// case "AU_ACN":
	// case "UK_NINO":
	// case "IN_VOTER":
	// case "IN_PAN":
	case "CREDIT_CARD":
		validLuhn := true
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
				GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{
					ValidLuhn: &validLuhn,
				},
			},
		}
	// case "NRP":
	// case "IT_FISCAL_CODE":
	case "PERSON":
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
				GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
			},
		}
	// case "US_DRIVER_LICENSE":
	// case "SG_NRIC_FIN":
	// case "IT_DRIVER_LICENSE":
	// case "URL":
	// case "LOCATION":
	// case "US_PASSPORT":
	// case "IN_VEHICLE_REGISTRATION":
	case "PHONE_NUMBER":
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig{
				GenerateStringPhoneNumberConfig: &mgmtv1alpha1.GenerateStringPhoneNumber{},
			},
		}
	// case "DATE_TIME":
	// case "CRYPTO":
	case "US_SSN":
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
				GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
			},
		}
	// case "US_BANK_NUMBER":
	case "IP_ADDRESS":
		ipType := mgmtv1alpha1.GenerateIpAddressType_GENERATE_IP_ADDRESS_TYPE_V4_PUBLIC
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateIpAddressConfig{
				GenerateIpAddressConfig: &mgmtv1alpha1.GenerateIpAddress{IpType: &ipType},
			},
		}
	// case "UK_NHS":
	// case "IBAN_CODE":
	// case "IT_VAT_CODE":
	// case "IT_PASSPORT":
	// case "IT_IDENTITY_CARD":
	// case "AU_ABN":
	// case "US_ITIN":
	case "EMAIL_ADDRESS":
		emailType := mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UUID_V4
		invalidEmailAction := mgmtv1alpha1.InvalidEmailAction_INVALID_EMAIL_ACTION_GENERATE
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
				TransformEmailConfig: &mgmtv1alpha1.TransformEmail{EmailType: &emailType, InvalidEmailAction: &invalidEmailAction},
			},
		}
	default:
		return &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{
				GenerateSha256HashConfig: &mgmtv1alpha1.GenerateSha256Hash{},
			},
		}
	}
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

const (
	neosyncEntityPrefix = "NEOSYNC_"
)

func processAnalysisResultsForNeosyncTransformers(
	inputResults []presidioapi.RecognizerResultWithAnaysisExplanation,
	neosyncEnabledEntities []string,
	inputText string,
) (analysisResults []presidioapi.RecognizerResultWithAnaysisExplanation, neosyncEntityMap map[string]*queue.Queue[string]) {
	entitySet := map[string]struct{}{}
	for _, entity := range neosyncEnabledEntities {
		entitySet[entity] = struct{}{}
	}

	output := make([]presidioapi.RecognizerResultWithAnaysisExplanation, 0, len(inputResults))
	neosyncEntityMap = map[string]*queue.Queue[string]{} // entity -> list of original values
	for _, result := range inputResults {
		if _, ok := entitySet[result.EntityType]; ok {
			result.EntityType = fmt.Sprintf("%s%s", neosyncEntityPrefix, result.EntityType)
			if _, ok := neosyncEntityMap[result.EntityType]; !ok {
				neosyncEntityMap[result.EntityType] = queue.NewQueue[string]()
			}
			neosyncEntityMap[result.EntityType].Enqueue(inputText[result.Start:result.End])
		}
		output = append(output, result)
	}

	return output, neosyncEntityMap
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
		err := ap.FromReplace(presidioapi.Replace{Type: "replace", NewValue: withNeosyncEntityBumpers(fmt.Sprintf("%s%s", neosyncEntityPrefix, entity))})
		if err != nil {
			return nil, false, err
		}
		return ap, true, nil
	}
	return nil, false, nil
}

func withNeosyncEntityBumpers(text string) string {
	return fmt.Sprintf("{{%s%s}}", neosyncEntityPrefix, text)
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
