package presidioapi

import (
	"context"
	"net/http"
)

func NewPresidioClientFromClients(
	analyzeClient AnalyzeInterface,
	anonymizeClient AnonymizeInterface,
) PresidioClientInterface {
	return &PresidioClient{analyze: analyzeClient, anonymize: anonymizeClient}
}

type PresidioClientConfig struct {
	authToken *string

	httpclient *http.Client
}

type PresidioClientOption func(any)

func NewPresidioClientFromConfig(
	analyzeUrl string,
	anonymizeUrl string,
) {
}

var _ PresidioClientInterface = (*PresidioClient)(nil)

type PresidioClient struct {
	analyze   AnalyzeInterface
	anonymize AnonymizeInterface
}

func (p *PresidioClient) PostAnalyzeWithResponse(ctx context.Context, body PostAnalyzeJSONRequestBody, reqEditors ...RequestEditorFn) (*PostAnalyzeResponse, error) { //nolint
	return p.analyze.PostAnalyzeWithResponse(ctx, body, reqEditors...)
}

func (p *PresidioClient) PostAnonymizeWithResponse(ctx context.Context, body PostAnonymizeJSONRequestBody, reqEditors ...RequestEditorFn) (*PostAnonymizeResponse, error) {
	return p.anonymize.PostAnonymizeWithResponse(ctx, body, reqEditors...)
}
