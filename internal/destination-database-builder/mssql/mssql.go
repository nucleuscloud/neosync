package ddbuilder_mssql

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_mssql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mssql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	destdb_shared "github.com/nucleuscloud/neosync/internal/destination-database-builder/shared"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	ee_sqlmanager_mssql "github.com/nucleuscloud/neosync/internal/ee/mssql-manager"
)

type MssqlDestinationDatabaseBuilderService struct {
	logger                *slog.Logger
	eelicense             license.EEInterface
	sqlmanagerclient      sqlmanager.SqlManagerClient
	sourceConnection      *mgmtv1alpha1.Connection
	destinationConnection *mgmtv1alpha1.Connection
	destOpts              *mgmtv1alpha1.MssqlDestinationConnectionOptions
	destdb                *sqlmanager.SqlConnection
	sourcedb              *sqlmanager.SqlConnection
}

func NewMssqlDestinationDatabaseBuilderService(
	ctx context.Context,
	logger *slog.Logger,
	eelicense license.EEInterface,
	session connectionmanager.SessionInterface,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	sourceConnection *mgmtv1alpha1.Connection,
	destinationConnection *mgmtv1alpha1.Connection,
	destOpts *mgmtv1alpha1.MssqlDestinationConnectionOptions,
) (*MssqlDestinationDatabaseBuilderService, error) {
	sourcedb, err := sqlmanagerclient.NewSqlConnection(ctx, session, sourceConnection, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}

	destdb, err := sqlmanagerclient.NewSqlConnection(ctx, session, destinationConnection, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}

	return &MssqlDestinationDatabaseBuilderService{
		logger:                logger,
		eelicense:             eelicense,
		sqlmanagerclient:      sqlmanagerclient,
		sourceConnection:      sourceConnection,
		destinationConnection: destinationConnection,
		destOpts:              destOpts,
		destdb:                destdb,
		sourcedb:              sourcedb,
	}, nil
}

func (d *MssqlDestinationDatabaseBuilderService) InitializeSchema(ctx context.Context, uniqueTables map[string]struct{}) ([]*destdb_shared.InitSchemaError, error) {
	initErrors := []*destdb_shared.InitSchemaError{}
	if !d.destOpts.GetInitTableSchema() {
		d.logger.Info("skipping schema init as it is not enabled")
		return initErrors, nil
	}
	if !d.eelicense.IsValid() {
		return nil, fmt.Errorf("invalid or non-existent Neosync License. SQL Server schema init requires valid Enterprise license.")
	}
	tables := []*sqlmanager_shared.SchemaTable{}
	for tableKey := range uniqueTables {
		schema, table := sqlmanager_shared.SplitTableKey(tableKey)
		tables = append(tables, &sqlmanager_shared.SchemaTable{Schema: schema, Table: table})
	}

	initblocks, err := d.sourcedb.Db().GetSchemaInitStatements(ctx, tables)
	if err != nil {
		return nil, err
	}

	for _, block := range initblocks {
		d.logger.Info(fmt.Sprintf("[%s] found %d statements to execute during schema initialization", block.Label, len(block.Statements)))
		if len(block.Statements) == 0 {
			continue
		}
		for _, stmt := range block.Statements {
			err = d.destdb.Db().Exec(ctx, stmt)
			if err != nil {
				d.logger.Error(fmt.Sprintf("unable to exec mssql %s statements: %s", block.Label, err.Error()))
				if block.Label != ee_sqlmanager_mssql.SchemasLabel && block.Label != ee_sqlmanager_mssql.ViewsFunctionsLabel && block.Label != ee_sqlmanager_mssql.TableIndexLabel {
					return nil, fmt.Errorf("unable to exec mssql %s statements: %w", block.Label, err)
				}
				initErrors = append(initErrors, &destdb_shared.InitSchemaError{
					Statement: stmt,
					Error:     err.Error(),
				})
			}
		}
	}
	return initErrors, nil
}

func (d *MssqlDestinationDatabaseBuilderService) TruncateData(ctx context.Context, uniqueTables map[string]struct{}, uniqueSchemas []string) error {
	if !d.destOpts.GetTruncateTable().GetTruncateBeforeInsert() {
		d.logger.Info("skipping truncate as it is not enabled")
		return nil
	}
	tableDependencies, err := d.sourcedb.Db().GetTableConstraintsBySchema(ctx, uniqueSchemas)
	if err != nil {
		return fmt.Errorf("unable to retrieve database foreign key constraints: %w", err)
	}
	d.logger.Info(fmt.Sprintf("found %d foreign key constraints for database", len(tableDependencies.ForeignKeyConstraints)))
	tablePrimaryDependencyMap := destdb_shared.GetFilteredForeignToPrimaryTableMap(tableDependencies.ForeignKeyConstraints, uniqueTables)
	orderedTablesResp, err := tabledependency.GetTablesOrderedByDependency(tablePrimaryDependencyMap)
	if err != nil {
		return err
	}

	orderedTableDelete := []string{}
	for i := len(orderedTablesResp.OrderedTables) - 1; i >= 0; i-- {
		st := orderedTablesResp.OrderedTables[i]
		stmt, err := sqlmanager_mssql.BuildMssqlDeleteStatement(st.Schema, st.Table)
		if err != nil {
			return err
		}
		orderedTableDelete = append(orderedTableDelete, stmt)
	}

	d.logger.Info(fmt.Sprintf("executing %d sql statements that will delete from tables", len(orderedTableDelete)))
	err = d.destdb.Db().BatchExec(ctx, 10, orderedTableDelete, &sqlmanager_shared.BatchExecOpts{})
	if err != nil {
		return fmt.Errorf("unable to exec ordered delete from statements: %w", err)
	}

	// reset identity column counts
	schemaColMap, err := d.sourcedb.Db().GetSchemaColumnMap(ctx)
	if err != nil {
		return err
	}

	identityStmts := []string{}
	for table, cols := range schemaColMap {
		if _, ok := uniqueTables[table]; !ok {
			continue
		}
		for _, c := range cols {
			if c.IdentityGeneration != nil && *c.IdentityGeneration != "" {
				schema, table := sqlmanager_shared.SplitTableKey(table)
				identityResetStatement := sqlmanager_mssql.BuildMssqlIdentityColumnResetStatement(schema, table, c.IdentitySeed, c.IdentityIncrement)
				identityStmts = append(identityStmts, identityResetStatement)
			}
		}
	}
	if len(identityStmts) > 0 {
		err = d.destdb.Db().BatchExec(ctx, 10, identityStmts, &sqlmanager_shared.BatchExecOpts{})
		if err != nil {
			return fmt.Errorf("unable to exec identity reset statements: %w", err)
		}
	}
	return nil
}

func (d *MssqlDestinationDatabaseBuilderService) CloseConnections() {
	d.destdb.Db().Close()
	d.sourcedb.Db().Close()
}
