package v1alpha1_useraccountservice

import (
	auth_client "github.com/nucleuscloud/neosync/backend/internal/auth/client"
	"github.com/nucleuscloud/neosync/backend/internal/authmgmt"
	"github.com/nucleuscloud/neosync/backend/internal/ee/rbac"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"github.com/nucleuscloud/neosync/backend/internal/temporal/clientmanager"
	"github.com/nucleuscloud/neosync/internal/billing"
)

type Service struct {
	cfg                    *Config
	db                     *neosyncdb.NeosyncDb
	temporalConfigProvider clientmanager.ConfigProvider
	authclient             auth_client.Interface
	authadminclient        authmgmt.Interface
	billingclient          billing.Interface
	rbacClient             *rbac.Rbac
}

type Config struct {
	IsAuthEnabled            bool
	IsNeosyncCloud           bool
	DefaultMaxAllowedRecords *int64
}

func New(
	cfg *Config,
	db *neosyncdb.NeosyncDb,
	temporalConfigProvider clientmanager.ConfigProvider,
	authclient auth_client.Interface,
	authadminclient authmgmt.Interface,
	billingclient billing.Interface,
	rbacClient *rbac.Rbac,
) *Service {
	return &Service{
		cfg:                    cfg,
		db:                     db,
		temporalConfigProvider: temporalConfigProvider,
		authclient:             authclient,
		authadminclient:        authadminclient,
		billingclient:          billingclient,
		rbacClient:             rbacClient,
	}
}
