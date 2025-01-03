package v1alpha1_connectiondataservice

import (
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/connectiondata"
)

type Service struct {
	cfg                   *Config
	connectionService     mgmtv1alpha1connect.ConnectionServiceClient
	connectiondatabuilder connectiondata.ConnectionDataBuilder
}

type Config struct {
}

func New(
	cfg *Config,
	connectionService mgmtv1alpha1connect.ConnectionServiceClient,
	connectiondatabuilder connectiondata.ConnectionDataBuilder,
) *Service {
	return &Service{
		cfg:                   cfg,
		connectionService:     connectionService,
		connectiondatabuilder: connectiondatabuilder,
	}
}
