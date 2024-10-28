package benthosbuilder_connections

import (
	"context"
	"errors"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal/shared"
)

type neosyncConnectionDataBuilder struct {
	connectionDataClient mgmtv1alpha1connect.ConnectionDataServiceClient
}

func NewNeosyncConnectionDataSyncBuilder(
	connectionDataClient mgmtv1alpha1connect.ConnectionDataServiceClient,
) bb_shared.ConnectionBenthosBuilder {
	return &neosyncConnectionDataBuilder{
		connectionDataClient: connectionDataClient,
	}
}

func (b *neosyncConnectionDataBuilder) BuildSourceConfigs(ctx context.Context, params *bb_shared.SourceParams) ([]*bb_shared.BenthosSourceConfig, error) {
	configs := []*bb_shared.BenthosSourceConfig{}
	return configs, nil
}

func (b *neosyncConnectionDataBuilder) BuildDestinationConfig(ctx context.Context, params *bb_shared.DestinationParams) (*bb_shared.BenthosDestinationConfig, error) {
	return nil, errors.ErrUnsupported
}
