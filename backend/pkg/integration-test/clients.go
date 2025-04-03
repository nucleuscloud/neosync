package integrationtests_test

import (
	"net/http"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	http_client "github.com/nucleuscloud/neosync/internal/http/client"
)

type NeosyncClients struct {
	httpUrl string
}

func newNeosyncClients(httpUrl string) *NeosyncClients {
	return &NeosyncClients{
		httpUrl: httpUrl,
	}
}

type clientConfig struct {
	userId string
}

type ClientConfigOption func(*clientConfig)

func WithUserId(userId string) ClientConfigOption {
	return func(c *clientConfig) {
		c.userId = userId
	}
}

func (s *NeosyncClients) Users(
	opts ...ClientConfigOption,
) mgmtv1alpha1connect.UserAccountServiceClient {
	config := getHydratedClientConfig(opts...)
	return mgmtv1alpha1connect.NewUserAccountServiceClient(getHttpClient(config), s.httpUrl)
}

func (s *NeosyncClients) Connections(
	opts ...ClientConfigOption,
) mgmtv1alpha1connect.ConnectionServiceClient {
	config := getHydratedClientConfig(opts...)
	return mgmtv1alpha1connect.NewConnectionServiceClient(getHttpClient(config), s.httpUrl)
}

func (s *NeosyncClients) Anonymize(
	opts ...ClientConfigOption,
) mgmtv1alpha1connect.AnonymizationServiceClient {
	config := getHydratedClientConfig(opts...)
	return mgmtv1alpha1connect.NewAnonymizationServiceClient(getHttpClient(config), s.httpUrl)
}

func (s *NeosyncClients) Jobs(opts ...ClientConfigOption) mgmtv1alpha1connect.JobServiceClient {
	config := getHydratedClientConfig(opts...)
	return mgmtv1alpha1connect.NewJobServiceClient(getHttpClient(config), s.httpUrl)
}

func (s *NeosyncClients) Transformers(
	opts ...ClientConfigOption,
) mgmtv1alpha1connect.TransformersServiceClient {
	config := getHydratedClientConfig(opts...)
	return mgmtv1alpha1connect.NewTransformersServiceClient(getHttpClient(config), s.httpUrl)
}

func (s *NeosyncClients) ConnectionData(
	opts ...ClientConfigOption,
) mgmtv1alpha1connect.ConnectionDataServiceClient {
	config := getHydratedClientConfig(opts...)
	return mgmtv1alpha1connect.NewConnectionDataServiceClient(getHttpClient(config), s.httpUrl)
}

func (s *NeosyncClients) AccountHooks(
	opts ...ClientConfigOption,
) mgmtv1alpha1connect.AccountHookServiceClient {
	config := getHydratedClientConfig(opts...)
	return mgmtv1alpha1connect.NewAccountHookServiceClient(getHttpClient(config), s.httpUrl)
}

func getHydratedClientConfig(opts ...ClientConfigOption) *clientConfig {
	config := &clientConfig{}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

func getHttpClient(config *clientConfig) *http.Client {
	if config.userId != "" {
		return http_client.WithBearerAuth(&http.Client{}, &config.userId)
	}
	return &http.Client{}
}
