package v1alpha_anonymizationservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"go.opentelemetry.io/otel/metric"
)

type Service struct {
	cfg                *Config
	meter              metric.Meter // optional
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
}

type Config struct {
	IsAuthEnabled bool
}

func New(
	cfg *Config,
	meter metric.Meter,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
) *Service {
	return &Service{
		cfg:                cfg,
		meter:              meter,
		useraccountService: useraccountService,
	}
}
