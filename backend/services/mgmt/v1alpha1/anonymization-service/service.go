package v1alpha_anonymizationservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	"go.opentelemetry.io/otel/metric"
)

type Service struct {
	cfg                *Config
	meter              metric.Meter // optional
	userdataclient     userdata.Interface
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	transformerClient  mgmtv1alpha1connect.TransformersServiceClient
	analyze            presidioapi.AnalyzeInterface
	anonymize          presidioapi.AnonymizeInterface
	db                 *neosyncdb.NeosyncDb
}

type Config struct {
	IsAuthEnabled           bool
	IsPresidioEnabled       bool
	PresidioDefaultLanguage *string
	IsNeosyncCloud          bool
}

func New(
	cfg *Config,
	meter metric.Meter,
	userdataclient userdata.Interface,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	transformerClient mgmtv1alpha1connect.TransformersServiceClient,
	analyzeclient presidioapi.AnalyzeInterface,
	anonymizeclient presidioapi.AnonymizeInterface,
	db *neosyncdb.NeosyncDb,
) *Service {
	return &Service{
		cfg:                cfg,
		meter:              meter,
		userdataclient:     userdataclient,
		useraccountService: useraccountService,
		transformerClient:  transformerClient,
		analyze:            analyzeclient,
		anonymize:          anonymizeclient,
		db:                 db,
	}
}
