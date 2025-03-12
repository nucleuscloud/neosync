package schemamanager_notsupported

import (
	"context"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	shared "github.com/nucleuscloud/neosync/internal/schema-manager/shared"
)

type NotSupportedSchemaManager struct {
}

func NewNotSupportedSchemaManager() (*NotSupportedSchemaManager, error) {
	return &NotSupportedSchemaManager{}, nil
}

func (d *NotSupportedSchemaManager) InitializeSchema(ctx context.Context, uniqueTables map[string]struct{}) ([]*shared.InitSchemaError, error) {
	return []*shared.InitSchemaError{}, nil
}

func (d *NotSupportedSchemaManager) TruncateData(ctx context.Context, uniqueTables map[string]struct{}, uniqueSchemas []string) error {
	return nil
}

func (d *NotSupportedSchemaManager) CalculateSchemaDiff(ctx context.Context, uniqueTables map[string]*sqlmanager_shared.SchemaTable) (*shared.SchemaDifferences, error) {
	return nil, nil
}

func (d *NotSupportedSchemaManager) BuildSchemaDiffStatements(ctx context.Context, diff *shared.SchemaDifferences) ([]*sqlmanager_shared.InitSchemaStatements, error) {
	return nil, nil
}

func (d *NotSupportedSchemaManager) ReconcileDestinationSchema(ctx context.Context, uniqueTables map[string]*sqlmanager_shared.SchemaTable, schemaStatements []*sqlmanager_shared.InitSchemaStatements) ([]*shared.InitSchemaError, error) {
	return []*shared.InitSchemaError{}, nil
}

func (d *NotSupportedSchemaManager) CloseConnections() {
}
