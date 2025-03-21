package schemamanager_postgres

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	shared "github.com/nucleuscloud/neosync/internal/schema-manager/shared"
	"golang.org/x/sync/errgroup"
)

type PostgresSchemaManager struct {
	logger                *slog.Logger
	sqlmanagerclient      sqlmanager.SqlManagerClient
	sourceConnection      *mgmtv1alpha1.Connection
	destinationConnection *mgmtv1alpha1.Connection
	destOpts              *mgmtv1alpha1.PostgresDestinationConnectionOptions
	destdb                *sqlmanager.SqlConnection
	sourcedb              *sqlmanager.SqlConnection
}

func NewPostgresSchemaManager(
	ctx context.Context,
	logger *slog.Logger,
	session connectionmanager.SessionInterface,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	sourceConnection *mgmtv1alpha1.Connection,
	destinationConnection *mgmtv1alpha1.Connection,
	destOpts *mgmtv1alpha1.PostgresDestinationConnectionOptions,
) (*PostgresSchemaManager, error) {
	sourcedb, err := sqlmanagerclient.NewSqlConnection(ctx, session, sourceConnection, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}

	destdb, err := sqlmanagerclient.NewSqlConnection(ctx, session, destinationConnection, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}

	return &PostgresSchemaManager{
		logger:                logger,
		sqlmanagerclient:      sqlmanagerclient,
		sourceConnection:      sourceConnection,
		destinationConnection: destinationConnection,
		destOpts:              destOpts,
		destdb:                destdb,
		sourcedb:              sourcedb,
	}, nil
}

func (d *PostgresSchemaManager) CalculateSchemaDiff(ctx context.Context, uniqueTables map[string]*sqlmanager_shared.SchemaTable) (*shared.SchemaDifferences, error) {
	d.logger.Debug("calculating schema diff")
	tables := []*sqlmanager_shared.SchemaTable{}
	for _, schematable := range uniqueTables {
		tables = append(tables, schematable)
	}

	sourceData := &shared.DatabaseData{}
	destData := &shared.DatabaseData{}
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		dbData, err := getDatabaseDataForSchemaDiff(errctx, d.sourcedb, tables)
		if err != nil {
			return fmt.Errorf("failed to get database data for schema diff: %w", err)
		}
		sourceData = dbData
		return nil
	})
	errgrp.Go(func() error {
		dbData, err := getDatabaseDataForSchemaDiff(errctx, d.destdb, tables)
		if err != nil {
			return fmt.Errorf("failed to get database data for schema diff: %w", err)
		}
		destData = dbData
		return nil
	})
	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}

	builder := shared.NewSchemaDifferencesBuilder(tables, sourceData, destData)
	return builder.Build(), nil
}

func getDatabaseDataForSchemaDiff(
	ctx context.Context,
	db *sqlmanager.SqlConnection,
	tables []*sqlmanager_shared.SchemaTable,
) (*shared.DatabaseData, error) {
	columns, err := db.Db().GetColumnsByTables(ctx, tables)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve database table schemas: %w", err)
	}
	columnsMap := shared.GetUniqueSchemaColMappings(columns)
	return &shared.DatabaseData{
		Columns: columnsMap,
	}, nil
}

func (d *PostgresSchemaManager) BuildSchemaDiffStatements(ctx context.Context, diff *shared.SchemaDifferences) ([]*sqlmanager_shared.InitSchemaStatements, error) {
	d.logger.Debug("building schema diff statements")
	if !d.destOpts.GetInitTableSchema() {
		d.logger.Info("skipping schema init as it is not enabled")
		return nil, nil
	}
	addColumnStatements := []string{}
	for _, column := range diff.ExistsInSource.Columns {
		stmt, err := sqlmanager_postgres.BuildAddColumnStatement(column)
		if err != nil {
			return nil, fmt.Errorf("failed to build add column statement: %w", err)
		}
		addColumnStatements = append(addColumnStatements, stmt)
	}

	return []*sqlmanager_shared.InitSchemaStatements{
		{
			Label:      sqlmanager_postgres.AddColumnsLabel,
			Statements: addColumnStatements,
		},
	}, nil
}

func (d *PostgresSchemaManager) ReconcileDestinationSchema(ctx context.Context, uniqueTables map[string]*sqlmanager_shared.SchemaTable, schemaStatements []*sqlmanager_shared.InitSchemaStatements) ([]*shared.InitSchemaError, error) {
	d.logger.Debug("reconciling destination schema")
	initErrors := []*shared.InitSchemaError{}
	if !d.destOpts.GetInitTableSchema() {
		d.logger.Info("skipping schema init as it is not enabled")
		return initErrors, nil
	}
	tables := []*sqlmanager_shared.SchemaTable{}
	for _, tableschema := range uniqueTables {
		tables = append(tables, tableschema)
	}

	initblocks, err := d.sourcedb.Db().GetSchemaInitStatements(ctx, tables)
	if err != nil {
		return nil, err
	}

	schemaStatementsByLabel := map[string][]*sqlmanager_shared.InitSchemaStatements{}
	for _, statement := range schemaStatements {
		schemaStatementsByLabel[statement.Label] = append(schemaStatementsByLabel[statement.Label], statement)
	}

	// insert add columns statements after create table statements
	// initblocks will eventually be replaced by schemastatements
	// this is a weird intermitten state for now
	statementBlocks := []*sqlmanager_shared.InitSchemaStatements{}
	for _, statement := range initblocks {
		statementBlocks = append(statementBlocks, statement)
		if statement.Label == sqlmanager_postgres.CreateTablesLabel {
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_postgres.AddColumnsLabel]...)
		}
	}

	for _, block := range statementBlocks {
		d.logger.Info(fmt.Sprintf("[%s] found %d statements to execute during schema initialization", block.Label, len(block.Statements)))
		if len(block.Statements) == 0 {
			continue
		}
		err = d.destdb.Db().BatchExec(ctx, shared.BatchSizeConst, block.Statements, &sqlmanager_shared.BatchExecOpts{})
		if err != nil {
			d.logger.Error(fmt.Sprintf("unable to exec postgres %s statements: %s", block.Label, err.Error()))
			for _, stmt := range block.Statements {
				err = d.destdb.Db().BatchExec(ctx, 1, []string{stmt}, &sqlmanager_shared.BatchExecOpts{})
				if err != nil {
					initErrors = append(initErrors, &shared.InitSchemaError{
						Statement: stmt,
						Error:     err.Error(),
					})
				}
			}
		}
	}
	return initErrors, nil
}

func (d *PostgresSchemaManager) TruncateTables(ctx context.Context, schemaDiff *shared.SchemaDifferences) error {
	if !d.destOpts.GetTruncateTable().GetTruncateBeforeInsert() && !d.destOpts.GetTruncateTable().GetCascade() {
		d.logger.Info("skipping truncate as it is not enabled")
		return nil
	}
	uniqueTables := map[string]struct{}{}
	schemaMap := map[string]struct{}{}
	for _, table := range schemaDiff.ExistsInBoth.Tables {
		uniqueTables[table.String()] = struct{}{}
		schemaMap[table.Schema] = struct{}{}
	}
	uniqueSchemas := []string{}
	for schema := range schemaMap {
		uniqueSchemas = append(uniqueSchemas, schema)
	}
	return d.TruncateData(ctx, uniqueTables, uniqueSchemas)
}

func (d *PostgresSchemaManager) InitializeSchema(ctx context.Context, uniqueTables map[string]struct{}) ([]*shared.InitSchemaError, error) {
	initErrors := []*shared.InitSchemaError{}
	if !d.destOpts.GetInitTableSchema() {
		d.logger.Info("skipping schema init as it is not enabled")
		return initErrors, nil
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
		err = d.destdb.Db().BatchExec(ctx, shared.BatchSizeConst, block.Statements, &sqlmanager_shared.BatchExecOpts{})
		if err != nil {
			d.logger.Error(fmt.Sprintf("unable to exec pg %s statements: %s", block.Label, err.Error()))
			if block.Label != sqlmanager_postgres.SchemasLabel && block.Label != sqlmanager_postgres.ExtensionsLabel {
				return nil, fmt.Errorf("unable to exec pg %s statements: %w", block.Label, err)
			}
			for _, stmt := range block.Statements {
				err := d.destdb.Db().Exec(ctx, stmt)
				if err != nil {
					initErrors = append(initErrors, &shared.InitSchemaError{
						Statement: stmt,
						Error:     err.Error(),
					})
				}
			}
		}
	}
	return initErrors, nil
}

func (d *PostgresSchemaManager) TruncateData(ctx context.Context, uniqueTables map[string]struct{}, uniqueSchemas []string) error {
	if !d.destOpts.GetTruncateTable().GetTruncateBeforeInsert() && !d.destOpts.GetTruncateTable().GetCascade() {
		d.logger.Info("skipping truncate as it is not enabled")
		return nil
	}
	if d.destOpts.GetTruncateTable().GetCascade() {
		tableTruncateStmts := []string{}
		for table := range uniqueTables {
			schema, table := sqlmanager_shared.SplitTableKey(table)
			stmt, err := sqlmanager_postgres.BuildPgTruncateCascadeStatement(schema, table)
			if err != nil {
				return err
			}
			tableTruncateStmts = append(tableTruncateStmts, stmt)
		}
		d.logger.Info(fmt.Sprintf("executing %d sql statements that will truncate cascade tables", len(tableTruncateStmts)))
		err := d.destdb.Db().BatchExec(ctx, shared.BatchSizeConst, tableTruncateStmts, &sqlmanager_shared.BatchExecOpts{})
		if err != nil {
			return fmt.Errorf("unable to exec truncate cascade statements: %w", err)
		}
	} else if d.destOpts.GetTruncateTable().GetTruncateBeforeInsert() {
		tableDependencies, err := d.sourcedb.Db().GetTableConstraintsBySchema(ctx, uniqueSchemas)
		if err != nil {
			return fmt.Errorf("unable to retrieve database foreign key constraints: %w", err)
		}
		d.logger.Info(fmt.Sprintf("found %d foreign key constraints for database", len(tableDependencies.ForeignKeyConstraints)))
		tablePrimaryDependencyMap := shared.GetFilteredForeignToPrimaryTableMap(tableDependencies.ForeignKeyConstraints, uniqueTables)
		orderedTablesResp, err := tabledependency.GetTablesOrderedByDependency(tablePrimaryDependencyMap)
		if err != nil {
			return err
		}

		d.logger.Info(fmt.Sprintf("executing %d sql statements that will truncate tables", len(orderedTablesResp.OrderedTables)))
		truncateStmt, err := sqlmanager_postgres.BuildPgTruncateStatement(orderedTablesResp.OrderedTables)
		if err != nil {
			return fmt.Errorf("unable to build postgres truncate statement: %w", err)
		}
		err = d.destdb.Db().Exec(ctx, truncateStmt)
		if err != nil {
			return fmt.Errorf("unable to exec ordered truncate statements: %w", err)
		}
	}
	if d.destOpts.GetTruncateTable().GetTruncateBeforeInsert() || d.destOpts.GetTruncateTable().GetCascade() {
		// reset serial counts
		// identity counts are automatically reset with truncate identity restart clause
		schemaTableMap := map[string][]string{}
		for schemaTable := range uniqueTables {
			schema, table := sqlmanager_shared.SplitTableKey(schemaTable)
			schemaTableMap[schema] = append(schemaTableMap[schema], table)
		}

		for schema, tables := range schemaTableMap {
			sequences, err := d.destdb.Db().GetSequencesByTables(ctx, schema, tables)
			if err != nil {
				return err
			}
			resetSeqStmts := []string{}
			for _, seq := range sequences {
				resetSeqStmts = append(resetSeqStmts, sqlmanager_postgres.BuildPgResetSequenceSql(seq.Schema, seq.Name))
			}
			if len(resetSeqStmts) > 0 {
				err = d.destdb.Db().BatchExec(ctx, 10, resetSeqStmts, &sqlmanager_shared.BatchExecOpts{})
				if err != nil {
					// handle not found errors
					if !strings.Contains(err.Error(), `does not exist`) {
						return fmt.Errorf("unable to exec postgres sequence reset statements: %w", err)
					}
				}
			}
		}
	}
	return nil
}

func (d *PostgresSchemaManager) CloseConnections() {
	d.sourcedb.Db().Close()
	d.destdb.Db().Close()
}
