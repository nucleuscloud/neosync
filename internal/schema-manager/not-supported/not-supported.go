package schemamanager_notsupported

import (
	"context"

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

func (d *NotSupportedSchemaManager) CloseConnections() {
}
