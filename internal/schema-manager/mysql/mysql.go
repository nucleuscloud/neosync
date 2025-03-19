package schemamanager_mysql

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_mysql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mysql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	shared "github.com/nucleuscloud/neosync/internal/schema-manager/shared"
	"golang.org/x/sync/errgroup"
)

type MysqlSchemaManager struct {
	logger                *slog.Logger
	sqlmanagerclient      sqlmanager.SqlManagerClient
	sourceConnection      *mgmtv1alpha1.Connection
	destinationConnection *mgmtv1alpha1.Connection
	destOpts              *mgmtv1alpha1.MysqlDestinationConnectionOptions
	destdb                *sqlmanager.SqlConnection
	sourcedb              *sqlmanager.SqlConnection
}

func NewMysqlSchemaManager(
	ctx context.Context,
	logger *slog.Logger,
	session connectionmanager.SessionInterface,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	sourceConnection *mgmtv1alpha1.Connection,
	destinationConnection *mgmtv1alpha1.Connection,
	destOpts *mgmtv1alpha1.MysqlDestinationConnectionOptions,
) (*MysqlSchemaManager, error) {
	logger.Debug("creating mysql schema manager")
	logger.Debug("connecting to source database")
	sourcedb, err := sqlmanagerclient.NewSqlConnection(ctx, session, sourceConnection, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}
	logger.Debug("connecting to destination database")
	destdb, err := sqlmanagerclient.NewSqlConnection(ctx, session, destinationConnection, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}
	return &MysqlSchemaManager{
		logger:                logger,
		sqlmanagerclient:      sqlmanagerclient,
		sourceConnection:      sourceConnection,
		destinationConnection: destinationConnection,
		destOpts:              destOpts,
		destdb:                destdb,
		sourcedb:              sourcedb,
	}, nil
}

func (d *MysqlSchemaManager) CalculateSchemaDiff(ctx context.Context, uniqueTables map[string]*sqlmanager_shared.SchemaTable) (*shared.SchemaDifferences, error) {
	d.logger.Debug("calculating schema diff")
	tables := []*sqlmanager_shared.SchemaTable{}
	schemaMap := map[string][]*sqlmanager_shared.SchemaTable{}
	for _, schematable := range uniqueTables {
		tables = append(tables, schematable)
		schemaMap[schematable.Schema] = append(schemaMap[schematable.Schema], schematable)
	}

	sourceData := &shared.DatabaseData{}
	destData := &shared.DatabaseData{}
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		dbData, err := getDatabaseDataForSchemaDiff(errctx, d.sourcedb, tables, schemaMap)
		if err != nil {
			return fmt.Errorf("failed to get database data for schema diff: %w", err)
		}
		sourceData = dbData
		return nil
	})
	errgrp.Go(func() error {
		dbData, err := getDatabaseDataForSchemaDiff(errctx, d.destdb, tables, schemaMap)
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
	schemaMap map[string][]*sqlmanager_shared.SchemaTable,
) (*shared.DatabaseData, error) {
	columns := []*sqlmanager_shared.DatabaseSchemaRow{}
	nonFkConstraints := map[string]*sqlmanager_shared.NonForeignKeyConstraint{}
	fkConstraints := map[string]*sqlmanager_shared.ForeignKeyConstraint{}
	triggers := map[string]*sqlmanager_shared.TableTrigger{}
	functions := map[string]*sqlmanager_shared.DataType{}

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(5)

	errgrp.Go(func() error {
		cols, err := db.Db().GetDatabaseTableSchemasBySchemasAndTables(ctx, tables)
		if err != nil {
			return fmt.Errorf("failed to retrieve database table schemas: %w", err)
		}
		columns = cols
		return nil
	})

	errgrp.Go(func() error {
		datatypes, err := db.Db().GetSchemaTableDataTypes(ctx, tables)
		if err != nil {
			return fmt.Errorf("failed to retrieve database table functions: %w", err)
		}
		for _, tablefunction := range datatypes.Functions {
			functions[tablefunction.Fingerprint] = tablefunction
		}
		return nil
	})

	errgrp.Go(func() error {
		tabletriggers, err := db.Db().GetSchemaTableTriggers(ctx, tables)
		if err != nil {
			return fmt.Errorf("failed to retrieve database table triggers: %w", err)
		}
		for _, tabletrigger := range tabletriggers {
			triggers[tabletrigger.Fingerprint] = tabletrigger
		}
		return nil
	})

	mu := sync.Mutex{}
	for schema, tables := range schemaMap {
		tableNames := make([]string, len(tables))
		for i, table := range tables {
			tableNames[i] = table.Table
		}

		schema, tableNames := schema, tableNames
		errgrp.Go(func() error {
			tableconstraints, err := db.Db().GetTableConstraintsByTables(errctx, schema, tableNames)
			if err != nil {
				return fmt.Errorf("failed to retrieve  database table constraints for schema %s: %w", schema, err)
			}
			mu.Lock()
			defer mu.Unlock()
			for _, tableconstraint := range tableconstraints {
				for _, nonFkConstraint := range tableconstraint.NonForeignKeyConstraints {
					nonFkConstraints[nonFkConstraint.Fingerprint] = nonFkConstraint
				}
				for _, fkConstraint := range tableconstraint.ForeignKeyConstraints {
					fkConstraints[fkConstraint.Fingerprint] = fkConstraint
				}
			}
			return nil
		})
	}

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}
	columnsMap := sqlmanager_shared.GetUniqueSchemaColMappings(columns)

	return &shared.DatabaseData{
		Columns:                  columnsMap,
		NonForeignKeyConstraints: nonFkConstraints,
		ForeignKeyConstraints:    fkConstraints,
		Triggers:                 triggers,
		Functions:                functions,
	}, nil
}

func (d *MysqlSchemaManager) BuildSchemaDiffStatements(ctx context.Context, diff *shared.SchemaDifferences) ([]*sqlmanager_shared.InitSchemaStatements, error) {
	d.logger.Debug("building schema diff statements")
	if !d.destOpts.GetInitTableSchema() {
		d.logger.Info("skipping schema init as it is not enabled")
		return nil, nil
	}
	addColumnStatements := []string{}
	for _, column := range diff.ExistsInSource.Columns {
		stmt, err := sqlmanager_mysql.BuildAddColumnStatement(column)
		if err != nil {
			return nil, fmt.Errorf("failed to build add column statement: %w", err)
		}
		addColumnStatements = append(addColumnStatements, stmt)
	}

	dropNonFkConstraintStatements := []string{}
	for _, constraint := range diff.ExistsInDestination.NonForeignKeyConstraints {
		dropNonFkConstraintStatements = append(dropNonFkConstraintStatements, sqlmanager_mysql.BuildDropConstraintStatement(constraint.SchemaName, constraint.TableName, constraint.ConstraintType, constraint.ConstraintName))
	}

	orderedForeignKeysToDrop := shared.BuildOrderedForeignKeyConstraintsToDrop(d.logger, diff)
	orderedForeignKeyDropStatements := []string{}
	for _, fk := range orderedForeignKeysToDrop {
		orderedForeignKeyDropStatements = append(orderedForeignKeyDropStatements, sqlmanager_mysql.BuildDropConstraintStatement(fk.ReferencingSchema, fk.ReferencingTable, fk.ConstraintType, fk.ConstraintName))
	}

	dropColumnStatements := []string{}
	for _, column := range diff.ExistsInDestination.Columns {
		dropColumnStatements = append(dropColumnStatements, sqlmanager_mysql.BuildDropColumnStatement(column))
	}

	dropTriggerStatements := []string{}
	for _, trigger := range diff.ExistsInDestination.Triggers {
		dropTriggerStatements = append(dropTriggerStatements, sqlmanager_mysql.BuildDropTriggerStatement(trigger.TriggerSchema, trigger.TriggerName))
	}

	dropFunctionStatements := []string{}
	for _, function := range diff.ExistsInDestination.Functions {
		dropFunctionStatements = append(dropFunctionStatements, sqlmanager_mysql.BuildDropFunctionStatement(function.Schema, function.Name))
	}

	return []*sqlmanager_shared.InitSchemaStatements{
		{
			Label:      sqlmanager_mysql.DropFunctionsLabel,
			Statements: dropFunctionStatements,
		},
		{
			Label:      sqlmanager_mysql.DropTriggersLabel,
			Statements: dropTriggerStatements,
		},
		{
			Label:      sqlmanager_mysql.AddColumnsLabel,
			Statements: addColumnStatements,
		},
		{
			Label:      sqlmanager_mysql.DropForeignKeyConstraintsLabel,
			Statements: orderedForeignKeyDropStatements,
		},
		{
			Label:      sqlmanager_mysql.DropNonForeignKeyConstraintsLabel,
			Statements: dropNonFkConstraintStatements,
		},
		{
			Label:      sqlmanager_mysql.DropColumnsLabel,
			Statements: dropColumnStatements,
		},
	}, nil
}

func (d *MysqlSchemaManager) ReconcileDestinationSchema(ctx context.Context, uniqueTables map[string]*sqlmanager_shared.SchemaTable, schemaStatements []*sqlmanager_shared.InitSchemaStatements) ([]*shared.InitSchemaError, error) {
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
		if statement.Label == sqlmanager_mysql.SchemasLabel {
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_mysql.DropFunctionsLabel]...)
		}
		if statement.Label == sqlmanager_mysql.CreateTablesLabel {
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_mysql.DropTriggersLabel]...)
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_mysql.DropForeignKeyConstraintsLabel]...)
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_mysql.DropNonForeignKeyConstraintsLabel]...)
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_mysql.DropColumnsLabel]...)
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_mysql.AddColumnsLabel]...)
		}
	}

	for _, block := range statementBlocks {
		d.logger.Info(fmt.Sprintf("[%s] found %d statements to execute during schema initialization", block.Label, len(block.Statements)))
		if len(block.Statements) == 0 {
			continue
		}
		err = d.destdb.Db().BatchExec(ctx, shared.BatchSizeConst, block.Statements, &sqlmanager_shared.BatchExecOpts{})
		if err != nil {
			d.logger.Error(fmt.Sprintf("unable to exec mysql %s statements: %s", block.Label, err.Error()))
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

func (d *MysqlSchemaManager) TruncateTables(ctx context.Context, schemaDiff *shared.SchemaDifferences) error {
	if !d.destOpts.GetTruncateTable().GetTruncateBeforeInsert() {
		d.logger.Info("skipping truncate as it is not enabled")
		return nil
	}
	tableTruncate := []string{}
	for _, schemaTable := range schemaDiff.ExistsInBoth.Tables {
		stmt, err := sqlmanager_mysql.BuildMysqlTruncateStatement(schemaTable.Schema, schemaTable.Table)
		if err != nil {
			return err
		}
		tableTruncate = append(tableTruncate, stmt)
	}
	d.logger.Info(fmt.Sprintf("executing %d sql statements that will truncate tables", len(tableTruncate)))
	disableFkChecks := sqlmanager_shared.DisableForeignKeyChecks
	err := d.destdb.Db().BatchExec(ctx, shared.BatchSizeConst, tableTruncate, &sqlmanager_shared.BatchExecOpts{Prefix: &disableFkChecks})
	if err != nil {
		return err
	}
	return nil
}

func (d *MysqlSchemaManager) InitializeSchema(ctx context.Context, uniqueTables map[string]struct{}) ([]*shared.InitSchemaError, error) {
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
			d.logger.Error(fmt.Sprintf("unable to exec mysql %s statements: %s", block.Label, err.Error()))
			if block.Label != sqlmanager_mysql.SchemasLabel {
				return nil, fmt.Errorf("unable to exec mysql %s statements: %w", block.Label, err)
			}
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

func (d *MysqlSchemaManager) TruncateData(ctx context.Context, uniqueTables map[string]struct{}, uniqueSchemas []string) error {
	if !d.destOpts.GetTruncateTable().GetTruncateBeforeInsert() {
		d.logger.Info("skipping truncate as it is not enabled")
		return nil
	}
	tableTruncate := []string{}
	for table := range uniqueTables {
		schema, table := sqlmanager_shared.SplitTableKey(table)
		stmt, err := sqlmanager_mysql.BuildMysqlTruncateStatement(schema, table)
		if err != nil {
			return err
		}
		tableTruncate = append(tableTruncate, stmt)
	}
	d.logger.Info(fmt.Sprintf("executing %d sql statements that will truncate tables", len(tableTruncate)))
	disableFkChecks := sqlmanager_shared.DisableForeignKeyChecks
	err := d.destdb.Db().BatchExec(ctx, shared.BatchSizeConst, tableTruncate, &sqlmanager_shared.BatchExecOpts{Prefix: &disableFkChecks})
	if err != nil {
		return err
	}
	return nil
}

func (d *MysqlSchemaManager) CloseConnections() {
	d.sourcedb.Db().Close()
	d.destdb.Db().Close()
}
