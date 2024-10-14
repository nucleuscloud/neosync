package ee_transformer_fns

import (
	"context"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
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
	analyzeResp, err := analyzeClient.PostAnalyzeWithResponse(ctx, presidioapi.AnalyzeRequest{
		Text:           value,
		Language:       "en",
		ScoreThreshold: &threshold,
		// AdHocRecognizers: &[]presidioapi.PatternRecognizer{{}}, // todo: implement allow and deny lists
	})
	if err != nil {
		return "", fmt.Errorf("unable to analyze input: %w", err)
	}
	if analyzeResp.JSON200 == nil {
		return "", fmt.Errorf("received non-200 response from analyzer: %s %d %s", analyzeResp.Status(), analyzeResp.StatusCode(), string(analyzeResp.Body))
	}

	defaultAnon, err := getDefaultAnonymizer(config.GetDefaultAnonymizer())
	if err != nil {
		return "", fmt.Errorf("unable to build default anonymizer: %w", err)
	}
	anonResp, err := anonymizeClient.PostAnonymizeWithResponse(ctx, presidioapi.AnonymizeRequest{
		AnalyzerResults: presidioapi.ToAnonymizeRecognizerResults(*analyzeResp.JSON200),
		Text:            value,
		Anonymizers: &map[string]presidioapi.AnonymizeRequest_Anonymizers_AdditionalProperties{
			"DEFAULT": *defaultAnon,
		},
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

func getDefaultAnonymizer(dto *mgmtv1alpha1.PiiAnonymizer) (*presidioapi.AnonymizeRequest_Anonymizers_AdditionalProperties, error) {
	ap := &presidioapi.AnonymizeRequest_Anonymizers_AdditionalProperties{}
	switch cfg := dto.GetConfig().(type) {
	case *mgmtv1alpha1.PiiAnonymizer_Redact_:
		err := ap.FromRedact(presidioapi.Redact{Type: "redact"})
		if err != nil {
			return nil, err
		}
	case *mgmtv1alpha1.PiiAnonymizer_Replace_:
		err := ap.FromReplace(presidioapi.Replace{Type: "replace", NewValue: cfg.Replace.GetValue()})
		if err != nil {
			return nil, err
		}
	case *mgmtv1alpha1.PiiAnonymizer_Hash_:
		hashtype := toPresidioHashType(cfg.Hash.GetAlgo())
		err := ap.FromHash(presidioapi.Hash{Type: "hash", HashType: &hashtype})
		if err != nil {
			return nil, err
		}
	case *mgmtv1alpha1.PiiAnonymizer_Mask_:
		fromend := cfg.Mask.GetFromEnd()
		err := ap.FromMask(presidioapi.Mask{
			Type:        "mask",
			CharsToMask: int(cfg.Mask.GetCharsToMask()),
			FromEnd:     &fromend,
			MaskingChar: cfg.Mask.GetMaskingChar(),
		})
		if err != nil {
			return nil, err
		}
	default:
		err := ap.FromReplace(presidioapi.Replace{Type: "replace", NewValue: "<REDACTED>"})
		if err != nil {
			return nil, err
		}
	}
	return ap, nil
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
