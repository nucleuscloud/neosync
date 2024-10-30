package benthosbuilder_builders

import (
	"context"
	"errors"

	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
)

type generateAIBuilder struct {
}

func NewGenerateAIBuilder() bb_internal.ConnectionBenthosBuilder {
	return &generateBuilder{}
}

func (b *generateAIBuilder) BuildSourceConfigs(ctx context.Context, params *bb_internal.SourceParams) ([]*bb_internal.BenthosSourceConfig, error) {
	return nil, errors.ErrUnsupported
}

func (b *generateAIBuilder) BuildDestinationConfig(ctx context.Context, params *bb_internal.DestinationParams) (*bb_internal.BenthosDestinationConfig, error) {
	config := &bb_internal.BenthosDestinationConfig{}

	return config, nil
}
