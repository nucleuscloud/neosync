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
					srcNonFkMap[c.Fingerprint] = c
				}

				dstNonFkMap := make(map[string]*sqlmanager_shared.NonForeignKeyConstraint)
				for _, c := range dstTableConstraints.NonForeignKeyConstraints {
					dstNonFkMap[c.Fingerprint] = c
				}

				// in source but not in destination
				for fingerprint, cObj := range srcNonFkMap {
					if _, ok := dstNonFkMap[fingerprint]; !ok {
						diff.ExistsInSource.NonForeignKeyConstraints = append(diff.ExistsInSource.NonForeignKeyConstraints, cObj)
					}
				}
				// in destination but not in source
				for fingerprint, cObj := range dstNonFkMap {
					if _, ok := srcNonFkMap[fingerprint]; !ok {
						diff.ExistsInDestination.NonForeignKeyConstraints = append(diff.ExistsInDestination.NonForeignKeyConstraints, cObj)
					}
				}

				srcFkMap := make(map[string]*sqlmanager_shared.ForeignKeyConstraint)
				for _, c := range srcTableConstraints.ForeignKeyConstraints {
					srcFkMap[c.Fingerprint] = c
				}

				dstFkMap := make(map[string]*sqlmanager_shared.ForeignKeyConstraint)
				for _, c := range dstTableConstraints.ForeignKeyConstraints {
					dstFkMap[c.Fingerprint] = c
				}

				// in source but not in destination
				for fingerprint, cObj := range srcFkMap {
					if _, ok := dstFkMap[fingerprint]; !ok {
						diff.ExistsInSource.ForeignKeyConstraints = append(diff.ExistsInSource.ForeignKeyConstraints, cObj)
					}
				}
				// in destination but not in source
				for fingerprint, cObj := range dstFkMap {
					if _, ok := srcFkMap[fingerprint]; !ok {
						diff.ExistsInDestination.ForeignKeyConstraints = append(diff.ExistsInDestination.ForeignKeyConstraints, cObj)
					}
				}
			}

		}
	}

	fmt.Println()
	fmt.Println("DIFFERENCES")
	jsonF, _ = json.MarshalIndent(diff.ExistsInSource.ForeignKeyConstraints, "", " ")
	fmt.Printf("\n\n source fk constraints: %s \n\n", string(jsonF))
	jsonF, _ = json.MarshalIndent(diff.ExistsInDestination.ForeignKeyConstraints, "", " ")
	fmt.Printf("\n\n destination fk constraints: %s \n\n", string(jsonF))

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

	jsonF, _ := json.MarshalIndent(orderedForeignKeyDropStatements, "", " ")
	fmt.Printf("\n\n ordered fk drop statements: %s \n\n", string(jsonF))

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
	// 1) Collect all FKs that must be dropped
	fksToDrop := diff.ExistsInDestination.ForeignKeyConstraints

	// 2) Build a map of childTableKey -> slice of (FK objects)
	childToFks := make(map[string][]*sqlmanager_shared.ForeignKeyConstraint)

	// Also track every parent table key that appears
	parentSet := make(map[string]bool)

	for _, fk := range fksToDrop {
		childKey := fmt.Sprintf("%s.%s", fk.ReferencingSchema, fk.ReferencingTable)
		parentKey := fmt.Sprintf("%s.%s", fk.ReferencedSchema, fk.ReferencedTable)

		childToFks[childKey] = append(childToFks[childKey], fk)
		parentSet[parentKey] = true
	}

	// --------------------------------
	// STEP A: Drop self-referencing constraints first
	// --------------------------------
	var selfReferenceDrops []string
	for childKey, fkList := range childToFks {
		// We'll rebuild the slice of FKs after removing self-references
		var remainingFKs []*sqlmanager_shared.ForeignKeyConstraint
		for _, fk := range fkList {
			parentKey := fmt.Sprintf("%s.%s", fk.ReferencedSchema, fk.ReferencedTable)
			if childKey == parentKey {
				// This is self-referencing
				dropStmt := sqlmanager_mysql.BuildDropConstraintStatement(
					fk.ReferencingSchema,
					fk.ReferencingTable,
					fk.ConstraintType,
					fk.ConstraintName,
				)
				selfReferenceDrops = append(selfReferenceDrops, dropStmt)
			} else {
				remainingFKs = append(remainingFKs, fk)
			}
		}
		if len(remainingFKs) > 0 {
			childToFks[childKey] = remainingFKs
		} else {
			// If no FKs remain for this child, remove the entry
			delete(childToFks, childKey)
		}
	}

	// Remove self-referencing from parentSet as well,
	// by reconstructing the parentSet from whatever remains:
	newParentSet := make(map[string]bool)
	for _, fkList := range childToFks {
		for _, fk := range fkList {
			pk := fmt.Sprintf("%s.%s", fk.ReferencedSchema, fk.ReferencedTable)
			newParentSet[pk] = true
		}
	}
	parentSet = newParentSet

	// --------------------------------
	// STEP B: Topological-like approach for the rest
	// --------------------------------
	var orderedFkDrops []string

	// First, add all the self-referencing drops at the front
	// (or you can do so at the end, but typically you'd want them first).
	orderedFkDrops = append(orderedFkDrops, selfReferenceDrops...)

	// Keep track of how many FKs remain
	remainingFKCount := 0
	for _, fkList := range childToFks {
		remainingFKCount += len(fkList)
	}

	// Keep track to avoid infinite loops in case of multi-table cycles
	maxIterations := remainingFKCount + 10

	for iteration := 0; iteration < maxIterations; iteration++ {
		if remainingFKCount == 0 {
			// all done
			break
		}

		droppedAny := false
		var childKeysToRemove []string

		for childKey, fkList := range childToFks {
			// If childKey is not in parentSet, it is effectively a "leaf"
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
		if !droppedAny {
			// Log a warning or handle forcibly:
			d.logger.Warn("Potential cycle detected while dropping FKs. Some constraints cannot be topologically sorted.")
			// One fallback approach: just drop everything that remains
			// or return an error. The snippet below forcibly drops them:
			for _, fkList := range childToFks {
				for _, fk := range fkList {
					stmt := sqlmanager_mysql.BuildDropConstraintStatement(
						fk.ReferencingSchema,
						fk.ReferencingTable,
						fk.ConstraintType,
						fk.ConstraintName,
					)
					orderedFkDrops = append(orderedFkDrops, stmt)
				}
			}
			// We clear them out
			childToFks = make(map[string][]*sqlmanager_shared.ForeignKeyConstraint)
			break
		}

		// Rebuild parentSet for next iteration from the remaining FKs
		newParentSet := make(map[string]bool)
		for _, fkList := range childToFks {
			for _, fk := range fkList {
				pk := fmt.Sprintf("%s.%s", fk.ReferencedSchema, fk.ReferencedTable)
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
