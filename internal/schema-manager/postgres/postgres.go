package schemamanager_postgres

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

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
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(5)

	columns := []*sqlmanager_shared.TableColumn{}
	errgrp.Go(func() error {
		cols, err := db.Db().GetColumnsByTables(errctx, tables)
		if err != nil {
			return fmt.Errorf("failed to retrieve database table schemas: %w", err)
		}
		columns = cols
		return nil
	})

	nonFkConstraints := map[string]*sqlmanager_shared.NonForeignKeyConstraint{}
	fkConstraints := map[string]*sqlmanager_shared.ForeignKeyConstraint{}
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
					key := fmt.Sprintf("%s.%s.%s", nonFkConstraint.SchemaName, nonFkConstraint.TableName, nonFkConstraint.ConstraintName)
					nonFkConstraints[key] = nonFkConstraint
				}
				for _, fkConstraint := range tableconstraint.ForeignKeyConstraints {
					key := fmt.Sprintf("%s.%s.%s", fkConstraint.ReferencingSchema, fkConstraint.ReferencingTable, fkConstraint.ConstraintName)
					fkConstraints[key] = fkConstraint
				}
			}
			return nil
		})
	}

	triggers := map[string]*sqlmanager_shared.TableTrigger{}
	errgrp.Go(func() error {
		tabletriggers, err := db.Db().GetSchemaTableTriggers(ctx, tables)
		if err != nil {
			return fmt.Errorf("failed to retrieve database table triggers: %w", err)
		}
		for _, tabletrigger := range tabletriggers {
			key := fmt.Sprintf("%s.%s.%s", tabletrigger.Schema, tabletrigger.Table, tabletrigger.TriggerName)
			triggers[key] = tabletrigger
		}
		return nil
	})

	functions := map[string]*sqlmanager_shared.DataType{}
	domains := map[string]*sqlmanager_shared.DataType{}
	enums := map[string]*sqlmanager_shared.DataType{}
	composites := map[string]*sqlmanager_shared.DataType{}
	errgrp.Go(func() error {
		rows, err := db.Db().GetSchemaTableDataTypes(ctx, tables)
		if err != nil {
			return fmt.Errorf("failed to retrieve database table functions: %w", err)
		}
		for _, tablefunction := range rows.Functions {
			key := fmt.Sprintf("%s.%s", tablefunction.Schema, tablefunction.Name)
			functions[key] = tablefunction
		}
		for _, dt := range rows.Domains {
			key := fmt.Sprintf("%s.%s", dt.Schema, dt.Name)
			domains[key] = dt
		}
		for _, dt := range rows.Enums {
			key := fmt.Sprintf("%s.%s", dt.Schema, dt.Name)
			enums[key] = dt
		}
		for _, dt := range rows.Composites {
			key := fmt.Sprintf("%s.%s", dt.Schema, dt.Name)
			composites[key] = dt
		}
		return nil
	})

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	columnsMap := shared.GetUniqueSchemaColMappings(columns)
	return &shared.DatabaseData{
		Columns:                  columnsMap,
		NonForeignKeyConstraints: nonFkConstraints,
		ForeignKeyConstraints:    fkConstraints,
		Triggers:                 triggers,
		Functions:                functions,
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
		stmt := sqlmanager_postgres.BuildAddColumnStatement(column)
		commentStmt := sqlmanager_postgres.BuildUpdateCommentStatement(column.Schema, column.Table, column.Name, column.Comment)
		addColumnStatements = append(addColumnStatements, stmt, commentStmt)
	}

	dropNonFkConstraintStatements := []string{}
	for _, constraint := range diff.ExistsInDestination.NonForeignKeyConstraints {
		dropNonFkConstraintStatements = append(dropNonFkConstraintStatements, sqlmanager_postgres.BuildDropConstraintStatement(constraint.SchemaName, constraint.TableName, constraint.ConstraintName))
	}
	// only way to update non fk constraint is to drop and recreate
	for _, constraint := range diff.ExistsInBoth.Different.NonForeignKeyConstraints {
		dropNonFkConstraintStatements = append(dropNonFkConstraintStatements, sqlmanager_postgres.BuildDropConstraintStatement(constraint.SchemaName, constraint.TableName, constraint.ConstraintName))
	}

	dropFkConstraintStatements := []string{}
	for _, constraint := range diff.ExistsInDestination.ForeignKeyConstraints {
		dropFkConstraintStatements = append(dropFkConstraintStatements, sqlmanager_postgres.BuildDropConstraintStatement(constraint.ReferencingSchema, constraint.ReferencingTable, constraint.ConstraintName))
	}
	// only way to update fk constraint is to drop and recreate
	for _, constraint := range diff.ExistsInBoth.Different.ForeignKeyConstraints {
		dropFkConstraintStatements = append(dropFkConstraintStatements, sqlmanager_postgres.BuildDropConstraintStatement(constraint.ReferencingSchema, constraint.ReferencingTable, constraint.ConstraintName))
	}

	dropColumnStatements := []string{}
	for _, column := range diff.ExistsInDestination.Columns {
		dropColumnStatements = append(dropColumnStatements, sqlmanager_postgres.BuildDropColumnStatement(column.Schema, column.Table, column.Name))
	}

	dropTriggerStatements := []string{}
	for _, trigger := range diff.ExistsInDestination.Triggers {
		dropTriggerStatements = append(dropTriggerStatements, sqlmanager_postgres.BuildDropTriggerStatement(trigger.Schema, trigger.Table, trigger.TriggerName))
	}
	// only way to update trigger is to drop and recreate
	for _, trigger := range diff.ExistsInBoth.Different.Triggers {
		dropTriggerStatements = append(dropTriggerStatements, sqlmanager_postgres.BuildDropTriggerStatement(trigger.Schema, trigger.Table, trigger.TriggerName))
	}

	dropFunctionStatements := []string{}
	for _, function := range diff.ExistsInDestination.Functions {
		dropFunctionStatements = append(dropFunctionStatements, sqlmanager_postgres.BuildDropFunctionStatement(function.Schema, function.Name))
	}

	updateFunctionStatements := []string{}
	for _, function := range diff.ExistsInBoth.Different.Functions {
		updateFunctionStatements = append(updateFunctionStatements, sqlmanager_postgres.BuildUpdateFunctionStatement(function.Schema, function.Name, function.Definition))
	}

	return []*sqlmanager_shared.InitSchemaStatements{
		{
			Label:      sqlmanager_shared.AddColumnsLabel,
			Statements: addColumnStatements,
		},
		{
			Label:      sqlmanager_shared.DropForeignKeyConstraintsLabel,
			Statements: dropFkConstraintStatements,
		},
		{
			Label:      sqlmanager_shared.DropNonForeignKeyConstraintsLabel,
			Statements: dropNonFkConstraintStatements,
		},
		{
			Label:      sqlmanager_shared.DropColumnsLabel,
			Statements: dropColumnStatements,
		},
		{
			Label:      sqlmanager_shared.DropTriggersLabel,
			Statements: dropTriggerStatements,
		},
		{
			Label:      sqlmanager_shared.UpdateFunctionsLabel,
			Statements: updateFunctionStatements,
		},
		{
			Label:      sqlmanager_shared.DropFunctionsLabel,
			Statements: dropFunctionStatements,
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
		if statement.Label == sqlmanager_shared.CreateTablesLabel {
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_shared.DropTriggersLabel]...)
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_shared.DropFunctionsLabel]...)
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_shared.DropForeignKeyConstraintsLabel]...)
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_shared.DropNonForeignKeyConstraintsLabel]...)
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_shared.DropColumnsLabel]...)
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_shared.AddColumnsLabel]...)
			statementBlocks = append(statementBlocks, schemaStatementsByLabel[sqlmanager_shared.UpdateFunctionsLabel]...)
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
	if len(uniqueTables) == 0 && len(uniqueSchemas) == 0 {
		d.logger.Info("no tables or schemas to truncate")
		return nil
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
			if block.Label != sqlmanager_shared.SchemasLabel && block.Label != sqlmanager_shared.ExtensionsLabel {
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
