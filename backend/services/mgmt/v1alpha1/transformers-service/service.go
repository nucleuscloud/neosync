package v1alpha1_transformersservice

import (
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
)

type Service struct {
	cfg            *Config
	db             *neosyncdb.NeosyncDb
	entityclient   presidioapi.EntityInterface
	userdataclient userdata.Interface
}

type Config struct {
	IsPresidioEnabled bool
	IsNeosyncCloud    bool
}

func New(
	cfg *Config,
	db *neosyncdb.NeosyncDb,
	recognizerclient presidioapi.EntityInterface,
	userdataclient userdata.Interface,
) *Service {
	return &Service{
		cfg:            cfg,
		db:             db,
		entityclient:   recognizerclient,
		userdataclient: userdataclient,
	}
}
