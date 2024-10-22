package presidioapi

import (
	"context"
)

// Slimmed down Presidio Analyze Interface for use in Neosync systems
type AnalyzeInterface interface {
	PostAnalyzeWithResponse(ctx context.Context, body PostAnalyzeJSONRequestBody, reqEditors ...RequestEditorFn) (*PostAnalyzeResponse, error)
}

// Slimmed down Presidio Anonymize Interface for use in Neosync systems
type AnonymizeInterface interface {
	PostAnonymizeWithResponse(ctx context.Context, body PostAnonymizeJSONRequestBody, reqEditors ...RequestEditorFn) (*PostAnonymizeResponse, error)
}

// Slimmed down Presidio Recognizer Interface for use in Neosync systems
type RecognizerInterface interface {
	GetRecognizers(ctx context.Context, params *GetRecognizersParams, reqEditors ...RequestEditorFn) (*GetRecognizersResponse, error)
}
