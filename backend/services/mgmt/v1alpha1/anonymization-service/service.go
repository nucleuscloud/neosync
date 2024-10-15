package v1alpha_anonymizationservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	"go.opentelemetry.io/otel/metric"
)

type Service struct {
	cfg                *Config
	meter              metric.Meter // optional
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	analyze            presidioapi.AnalyzeInterface
	anonymize          presidioapi.AnonymizeInterface
}

type Config struct {
	IsAuthEnabled     bool
	IsPresidioEnabled bool
}

func New(
	cfg *Config,
	meter metric.Meter,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	analyzeclient presidioapi.AnalyzeInterface,
	anonymizeclient presidioapi.AnonymizeInterface,
) *Service {
	return &Service{
		cfg:                cfg,
		meter:              meter,
		useraccountService: useraccountService,
		analyze:            analyzeclient,
		anonymize:          anonymizeclient,
	}
}
