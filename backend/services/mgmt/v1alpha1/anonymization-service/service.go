package v1alpha_anonymizationservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"go.opentelemetry.io/otel/metric"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
)

type Service struct {
	cfg                *Config
	meter              metric.Meter // optional
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	analyze   presidioapi.AnalyzeInterface
	anonymize presidioapi.AnonymizeInterface
	
}

type Config struct {
	IsAuthEnabled bool
	IsPresidioEnabled bool
}

type Service struct {
	cfg       *Config
	analyze   presidioapi.AnalyzeInterface
	anonymize presidioapi.AnonymizeInterface
}

func New(
	cfg *Config,
	meter metric.Meter,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	analyzeclient   presidioapi.AnalyzeInterface
	anonymizeclient presidioapi.AnonymizeInterface
) *Service {
	return &Service{
		cfg:                cfg,
		meter:              meter,
		useraccountService: useraccountService,
		analyze:   analyzeclient,
		anonymize: anonymizeclient,
	}
}
