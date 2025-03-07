package ddbuilder_notsupported

import (
	"context"

	destdb_shared "github.com/nucleuscloud/neosync/internal/destination-database-builder/shared"
)

type NotSupportedDestinationDatabaseBuilderService struct {
}

func NewNotSupportedDestinationDatabaseBuilderService() (*NotSupportedDestinationDatabaseBuilderService, error) {
	return &NotSupportedDestinationDatabaseBuilderService{}, nil
}

func (d *NotSupportedDestinationDatabaseBuilderService) InitializeSchema(ctx context.Context, uniqueTables map[string]struct{}) ([]*destdb_shared.InitSchemaError, error) {
	return []*destdb_shared.InitSchemaError{}, nil
}

func (d *NotSupportedDestinationDatabaseBuilderService) TruncateData(ctx context.Context, uniqueTables map[string]struct{}, uniqueSchemas []string) error {
	return nil
}

func (d *NotSupportedDestinationDatabaseBuilderService) CloseConnections() {
}
