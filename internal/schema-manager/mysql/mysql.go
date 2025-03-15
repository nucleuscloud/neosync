package schemamanager_mysql

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_mysql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mysql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	shared "github.com/nucleuscloud/neosync/internal/schema-manager/shared"
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
	diff := &shared.SchemaDifferences{
		ExistsInSource: &shared.ExistsInSource{
			Tables:                   []*sqlmanager_shared.SchemaTable{},
			Columns:                  []*sqlmanager_shared.DatabaseSchemaRow{},
			NonForeignKeyConstraints: []*sqlmanager_shared.NonForeignKeyConstraint{},
			ForeignKeyConstraints:    []*sqlmanager_shared.ForeignKeyConstraint{},
		},
		ExistsInDestination: &shared.ExistsInDestination{
			Columns:                  []*sqlmanager_shared.DatabaseSchemaRow{},
			NonForeignKeyConstraints: []*sqlmanager_shared.NonForeignKeyConstraint{},
			ForeignKeyConstraints:    []*sqlmanager_shared.ForeignKeyConstraint{},
		},
		ExistsInBoth: &shared.ExistsInBoth{
			Tables: []*sqlmanager_shared.SchemaTable{},
		},
	}
	tables := []*sqlmanager_shared.SchemaTable{}
	schemaMap := map[string][]*sqlmanager_shared.SchemaTable{}
	for _, schematable := range uniqueTables {
		tables = append(tables, schematable)
		schemaMap[schematable.Schema] = append(schemaMap[schematable.Schema], schematable)
	}
	sourceColumns, err := d.sourcedb.Db().GetDatabaseTableSchemasBySchemasAndTables(ctx, tables)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve source database table schemas: %w", err)
	}
	destColumns, err := d.destdb.Db().GetDatabaseTableSchemasBySchemasAndTables(ctx, tables)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve destination database table schemas: %w", err)
	}
	sourceColMap := sqlmanager_shared.GetUniqueSchemaColMappings(sourceColumns)
	destColMap := sqlmanager_shared.GetUniqueSchemaColMappings(destColumns)

	// diff constraints
	sourceConstraints := map[string]*sqlmanager_shared.AllTableConstraints{}
	destConstraints := map[string]*sqlmanager_shared.AllTableConstraints{}
	for schema, tables := range schemaMap {
		tableNames := []string{}
		for _, table := range tables {
			tableNames = append(tableNames, table.Table)
		}
		sc, err := d.sourcedb.Db().GetTableConstraintsByTables(ctx, schema, tableNames)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve source database table constraints: %w", err)
		}
		dc, err := d.destdb.Db().GetTableConstraintsByTables(ctx, schema, tableNames)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve destination database table constraints: %w", err)
		}
		sourceConstraints = sc
		destConstraints = dc
	}

	jsonF, _ := json.MarshalIndent(sourceConstraints, "", " ")
	fmt.Printf("\n\n %s \n\n", string(jsonF))
	jsonF, _ = json.MarshalIndent(destConstraints, "", " ")
	fmt.Printf("\n\n %s \n\n", string(jsonF))

	for _, table := range tables {
		sourceTable := sourceColMap[table.String()]
		destTable := destColMap[table.String()]
		if len(sourceTable) > 0 && len(destTable) == 0 {
			diff.ExistsInSource.Tables = append(diff.ExistsInSource.Tables, table)
		} else if len(sourceTable) > 0 && len(destTable) > 0 {
			// table exists in both source and destination
			diff.ExistsInBoth.Tables = append(diff.ExistsInBoth.Tables, table)

			// column diff
			for _, column := range sourceTable {
				_, ok := destTable[column.ColumnName]
				if !ok {
					diff.ExistsInSource.Columns = append(diff.ExistsInSource.Columns, column)
				}
			}
			for _, column := range destTable {
				_, ok := sourceTable[column.ColumnName]
				if !ok {
					diff.ExistsInDestination.Columns = append(diff.ExistsInDestination.Columns, column)
				}
			}

			// constraint diff
			srcTableConstraints, hasSrcConstraints := sourceConstraints[table.String()]
			dstTableConstraints, hasDstConstraints := destConstraints[table.String()]

			// if there's nothing in source but something in dest => all in dest need to be dropped
			if !hasSrcConstraints && hasDstConstraints {
				diff.ExistsInDestination.NonForeignKeyConstraints = append(diff.ExistsInDestination.NonForeignKeyConstraints, dstTableConstraints.NonForeignKeyConstraints...)
				diff.ExistsInDestination.ForeignKeyConstraints = append(diff.ExistsInDestination.ForeignKeyConstraints, dstTableConstraints.ForeignKeyConstraints...)
			} else if hasSrcConstraints && !hasDstConstraints {
				// if there's constraints in source but none in dest => all in source need to be created
				diff.ExistsInSource.NonForeignKeyConstraints = append(diff.ExistsInSource.NonForeignKeyConstraints, srcTableConstraints.NonForeignKeyConstraints...)
				diff.ExistsInSource.ForeignKeyConstraints = append(diff.ExistsInSource.ForeignKeyConstraints, srcTableConstraints.ForeignKeyConstraints...)
			} else if hasSrcConstraints && hasDstConstraints {

				srcNonFkMap := make(map[string]*sqlmanager_shared.NonForeignKeyConstraint)
				for _, c := range srcTableConstraints.NonForeignKeyConstraints {
					srcNonFkMap[c.ConstraintName] = c
				}

				dstNonFkMap := make(map[string]*sqlmanager_shared.NonForeignKeyConstraint)
				for _, c := range dstTableConstraints.NonForeignKeyConstraints {
					dstNonFkMap[c.ConstraintName] = c
				}

				// in source but not in destination
				for cName, cObj := range srcNonFkMap {
					if _, ok := dstNonFkMap[cName]; !ok {
						diff.ExistsInSource.NonForeignKeyConstraints = append(diff.ExistsInSource.NonForeignKeyConstraints, cObj)
					}
				}
				// in destination but not in source
				for cName, cObj := range dstNonFkMap {
					if _, ok := srcNonFkMap[cName]; !ok {
						diff.ExistsInDestination.NonForeignKeyConstraints = append(diff.ExistsInDestination.NonForeignKeyConstraints, cObj)
					}
				}

				srcFkMap := make(map[string]*sqlmanager_shared.ForeignKeyConstraint)
				for _, c := range srcTableConstraints.ForeignKeyConstraints {
					srcFkMap[c.ConstraintName] = c
				}

				dstFkMap := make(map[string]*sqlmanager_shared.ForeignKeyConstraint)
				for _, c := range dstTableConstraints.ForeignKeyConstraints {
					dstFkMap[c.ConstraintName] = c
				}

				// in source but not in destination
				for cName, cObj := range srcFkMap {
					if _, ok := dstFkMap[cName]; !ok {
						diff.ExistsInSource.ForeignKeyConstraints = append(diff.ExistsInSource.ForeignKeyConstraints, cObj)
					}
				}
				// in destination but not in source
				for cName, cObj := range dstFkMap {
					if _, ok := srcFkMap[cName]; !ok {
						diff.ExistsInDestination.ForeignKeyConstraints = append(diff.ExistsInDestination.ForeignKeyConstraints, cObj)
					}
				}
			}

		}
	}

	fmt.Println()
	fmt.Println("DIFFERENCES")
	jsonF, _ = json.MarshalIndent(diff, "", " ")
	fmt.Printf("\n\n %s \n\n", string(jsonF))

	return diff, nil
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

	orderedForeignKeyDropStatements, err := d.buildOrderedForeignKeyDropStatements(diff)
	if err != nil {
		return nil, fmt.Errorf("failed to build ordered foreign key drop statements: %w", err)
	}

	dropColumnStatements := []string{}
	for _, column := range diff.ExistsInDestination.Columns {
		dropColumnStatements = append(dropColumnStatements, sqlmanager_mysql.BuildDropColumnStatement(column))
	}

	return []*sqlmanager_shared.InitSchemaStatements{
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

func (d *MysqlSchemaManager) buildOrderedForeignKeyDropStatements(diff *shared.SchemaDifferences) ([]string, error) {
	// 3) Build a map of childTableKey -> slice of (FK objects)
	//    so we know which foreign keys exist for each child table.
	fksToDrop := diff.ExistsInDestination.ForeignKeyConstraints
	childToFks := make(map[string][]*sqlmanager_shared.ForeignKeyConstraint)

	// Also track every parent table key that appears
	parentSet := make(map[string]bool)

	// Construct those sets
	for _, fk := range fksToDrop {
		childKey := fmt.Sprintf("%s.%s", fk.ReferencingSchema, fk.ReferencingTable)
		parentKey := fmt.Sprintf("%s.%s", fk.ReferencedSchema, fk.ReferencedTable)

		childToFks[childKey] = append(childToFks[childKey], fk)

		// Mark the parent key in a map
		parentSet[parentKey] = true
	}

	// We'll accumulate the final drop statements in order
	var orderedFkDrops []string

	// Keep track of how many FKs remain
	remainingFKCount := len(fksToDrop)

	// Keep track to avoid infinite loops in case of cycles
	maxIterations := remainingFKCount + 10

	// 4) Repeatedly find child tables that do NOT appear in parentSet
	//    => they can be dropped now (nothing depends on them).
	for iteration := 0; iteration < maxIterations; iteration++ {
		if remainingFKCount == 0 {
			// all done
			break
		}

		droppedAny := false
		var childKeysToRemove []string

		for childKey, fkList := range childToFks {
			// If childKey is not in parentSet, it is a "leaf" table
			// (no one depends on its constraints).
			if !parentSet[childKey] {
				// We can drop all FKs for this child
				for _, fk := range fkList {
					stmt := sqlmanager_mysql.BuildDropConstraintStatement(
						fk.ReferencingSchema,
						fk.ReferencingTable,
						fk.ConstraintType,
						fk.ConstraintName,
					)
					orderedFkDrops = append(orderedFkDrops, stmt)
				}

				// We'll remove this child from the map
				childKeysToRemove = append(childKeysToRemove, childKey)
				droppedAny = true

				// Decrement our total count
				remainingFKCount -= len(fkList)
			}
		}

		// Remove those child keys from the map
		for _, ckey := range childKeysToRemove {
			delete(childToFks, ckey)
		}

		// If no child was dropped in this pass, we might have a cycle
		// or everything left references each other.
		if !droppedAny {
			// If we haven't removed all FKs but can't drop anything more,
			// there's likely a cycle or a self-referencing FK.
			// We can either break or try a fallback approach:
			d.logger.Warn("Potential cycle detected while dropping FKs. A manual approach or specific ordering is needed.")
			break
		}

		// Rebuild parentSet for next iteration: we only want parent keys for
		// the *remaining* child → parent links
		newParentSet := make(map[string]bool)
		for _, fkList := range childToFks {
			for _, fk := range fkList {
				pk := sqlmanager_shared.SchemaTable{Schema: fk.ReferencedSchema, Table: fk.ReferencedTable}.String()
				newParentSet[pk] = true
			}
		}
		parentSet = newParentSet
	}
	return orderedFkDrops, nil
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
		if statement.Label == sqlmanager_mysql.CreateTablesLabel {
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
		fmt.Println()
		fmt.Println("executing statements", block.Label)
		for _, stmt := range block.Statements {
			fmt.Println()
			fmt.Println(stmt)
		}
		fmt.Println()
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
