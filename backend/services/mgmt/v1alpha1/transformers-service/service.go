package v1alpha1_transformersservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
)

type Service struct {
	cfg                *Config
	db                 *neosyncdb.NeosyncDb
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
	recognizerclient   presidioapi.RecognizerInterface
}

type Config struct {
	IsPresidioEnabled bool
	IsNeosyncCloud    bool
}

func New(
	cfg *Config,
	db *neosyncdb.NeosyncDb,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
	recognizerclient presidioapi.RecognizerInterface,
) *Service {
	return &Service{
		cfg:                cfg,
		db:                 db,
		useraccountService: useraccountService,
		recognizerclient:   recognizerclient,
	}
}
