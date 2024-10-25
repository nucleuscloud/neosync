package benthosbuilder_connections

import (
	"context"
	"fmt"

	bb_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
)

func NewPostgresBenthosBuilder(jobType bb_shared.JobType) (bb_shared.ConnectionBenthosBuilder, error) {
	switch jobType {
	case bb_shared.JobTypeSync:
		return NewPostgresSyncBuilder(), nil
	case bb_shared.JobTypeGenerate:
		return NewPostgresGenerateBuilder(), nil
	case bb_shared.JobTypeAIGenerate:
		return NewPostgresAIGenerateBuilder(), nil
	default:
		return nil, fmt.Errorf("unsupported postgres job type: %s", jobType)
	}
}

/*
	Sync
*/

type postgresSyncBuilder struct {
}

func NewPostgresSyncBuilder() bb_shared.ConnectionBenthosBuilder {
	return &postgresSyncBuilder{}
}

func (b *postgresSyncBuilder) BuildSourceConfig(ctx context.Context, params *bb_shared.SourceParams) (*bb_shared.BenthosSourceConfig, error) {
	config := &bb_shared.BenthosSourceConfig{
		ConnectionType: bb_shared.ConnectionTypePostgres,
		JobType:        bb_shared.JobTypeSync,
	}

	// Build SQL queries
	// Handle transformations
	// Configure processors
	// etc.

	return config, nil
}

func (b *postgresSyncBuilder) BuildDestinationConfig(ctx context.Context, params *bb_shared.DestinationParams) (*bb_shared.BenthosDestinationConfig, error) {
	config := &bb_shared.BenthosDestinationConfig{}

	return config, nil
}

/*
	Generate
*/

type postgresGenerateBuilder struct {
}

func NewPostgresGenerateBuilder() bb_shared.ConnectionBenthosBuilder {
	return &postgresGenerateBuilder{}
}

func (b *postgresGenerateBuilder) BuildSourceConfig(ctx context.Context, params *bb_shared.SourceParams) (*bb_shared.BenthosSourceConfig, error) {
	return &bb_shared.BenthosSourceConfig{
		ConnectionType: bb_shared.ConnectionTypePostgres,
		JobType:        bb_shared.JobTypeGenerate,
	}, nil
}

func (b *postgresGenerateBuilder) BuildDestinationConfig(ctx context.Context, params *bb_shared.DestinationParams) (*bb_shared.BenthosDestinationConfig, error) {
	config := &bb_shared.BenthosDestinationConfig{}

	return config, nil
}

/*
	AI Generate
*/

type postgresAIGenerateBuilder struct {
}

func NewPostgresAIGenerateBuilder() bb_shared.ConnectionBenthosBuilder {
	return &postgresGenerateBuilder{}
}

func (b *postgresAIGenerateBuilder) BuildAIGenerateSourceConfig(ctx context.Context, params *bb_shared.SourceParams) (*bb_shared.BenthosSourceConfig, error) {
	return &bb_shared.BenthosSourceConfig{
		ConnectionType: bb_shared.ConnectionTypePostgres,
		JobType:        bb_shared.JobTypeAIGenerate,
	}, nil
}

func (b *postgresAIGenerateBuilder) BuildAIGenerateDestinationConfig(ctx context.Context, params *bb_shared.DestinationParams) (*bb_shared.BenthosDestinationConfig, error) {
	config := &bb_shared.BenthosDestinationConfig{}

	return config, nil
}
