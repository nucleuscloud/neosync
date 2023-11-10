package v1alpha1_apikeyservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
)

type Service struct {
	cfg                *Config
	db                 *nucleusdb.NucleusDb
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient
}

type Config struct {
	IsAuthEnabled bool
}

func New(
	cfg *Config,
	db *nucleusdb.NucleusDb,
	useraccountService mgmtv1alpha1connect.UserAccountServiceClient,
) *Service {
	return &Service{cfg: cfg, db: db, useraccountService: useraccountService}
}
