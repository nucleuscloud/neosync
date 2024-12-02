package sqlmanager_postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/doug-martin/goqu/v9"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	"golang.org/x/sync/errgroup"
)

type PostgresManager struct {
	querier pg_queries.Querier
	db      pg_queries.DBTX
	close   func()
}

func NewManager(querier pg_queries.Querier, db pg_queries.DBTX, closer func()) *PostgresManager {
	return &PostgresManager{querier: querier, db: db, close: closer}
}

func (p *PostgresManager) GetDatabaseSchema(ctx context.Context) ([]*sqlmanager_shared.DatabaseSchemaRow, error) {
	dbSchemas, err := p.querier.GetDatabaseSchema(ctx, p.db)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return []*sqlmanager_shared.DatabaseSchemaRow{}, nil
	}
	result := []*sqlmanager_shared.DatabaseSchemaRow{}

	for _, row := range dbSchemas {
		var generatedType *string
		if row.GeneratedType != "" {
			generatedTypeCopy := row.GeneratedType
			generatedType = &generatedTypeCopy
		}
		var identityGeneration *string
		if row.IdentityGeneration != "" {
			val := row.IdentityGeneration
			identityGeneration = &val
		}
		result = append(result, &sqlmanager_shared.DatabaseSchemaRow{
			TableSchema:            row.SchemaName,
			TableName:              row.TableName,
			ColumnName:             row.ColumnName,
			DataType:               row.DataType,
			ColumnDefault:          row.ColumnDefault,
			IsNullable:             row.IsNullable != "NO",
			CharacterMaximumLength: int(row.CharacterMaximumLength),
			NumericPrecision:       int(row.NumericPrecision),
			NumericScale:           int(row.NumericScale),
			OrdinalPosition:        int(row.OrdinalPosition),
			GeneratedType:          generatedType,
			IdentityGeneration:     identityGeneration,
		})
	}
	return result, nil
}

// returns: {public.users: { id: struct{}{}, created_at: struct{}{}}}
func (p *PostgresManager) GetSchemaColumnMap(ctx context.Context) (map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow, error) {
	dbSchemas, err := p.GetDatabaseSchema(ctx)
	if err != nil {
		return nil, err
	}
	result := sqlmanager_shared.GetUniqueSchemaColMappings(dbSchemas)
	return result, nil
}

func (p *PostgresManager) GetTableConstraintsBySchema(ctx context.Context, schemas []string) (*sqlmanager_shared.TableConstraints, error) {
	if len(schemas) == 0 {
		return &sqlmanager_shared.TableConstraints{}, nil
	}
	rows, err := p.querier.GetTableConstraintsBySchema(ctx, p.db, schemas)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return &sqlmanager_shared.TableConstraints{}, nil
	}

	foreignKeyMap := map[string][]*sqlmanager_shared.ForeignConstraint{}
	primaryKeyMap := map[string][]string{}
	uniqueConstraintsMap := map[string][][]string{}
	for _, row := range rows {
		tableName := sqlmanager_shared.BuildTable(row.SchemaName, row.TableName)
		switch row.ConstraintType {
		case "f":
			if len(row.ConstraintColumns) != len(row.ForeignColumnNames) {
				return nil, fmt.Errorf("length of columns was not equal to length of foreign key cols: %d %d", len(row.ConstraintColumns), len(row.ForeignColumnNames))
			}
			if len(row.ConstraintColumns) != len(row.Notnullable) {
				return nil, fmt.Errorf("length of columns was not equal to length of not nullable cols: %d %d", len(row.ConstraintColumns), len(row.Notnullable))
			}

			foreignKeyMap[tableName] = append(foreignKeyMap[tableName], &sqlmanager_shared.ForeignConstraint{
				Columns:     row.ConstraintColumns,
				NotNullable: row.Notnullable,
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   sqlmanager_shared.BuildTable(row.ForeignSchemaName, row.ForeignTableName),
					Columns: row.ForeignColumnNames,
				},
			})
		case "p":
			if _, exists := primaryKeyMap[tableName]; !exists {
				primaryKeyMap[tableName] = []string{}
			}
			primaryKeyMap[tableName] = append(primaryKeyMap[tableName], sqlmanager_shared.DedupeSlice(row.ConstraintColumns)...)
		case "u":
			columns := sqlmanager_shared.DedupeSlice(row.ConstraintColumns)
			uniqueConstraintsMap[tableName] = append(uniqueConstraintsMap[tableName], columns)
		}
	}
	return &sqlmanager_shared.TableConstraints{
		ForeignKeyConstraints: foreignKeyMap,
		PrimaryKeyConstraints: primaryKeyMap,
		UniqueConstraints:     uniqueConstraintsMap,
	}, nil
}

func (p *PostgresManager) GetRolePermissionsMap(ctx context.Context) (map[string][]string, error) {
	rows, err := p.querier.GetPostgresRolePermissions(ctx, p.db)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return map[string][]string{}, nil
	}

	schemaTablePrivsMap := map[string][]string{}
	for _, permission := range rows {
		key := sqlmanager_shared.BuildTable(permission.TableSchema, permission.TableName)
		schemaTablePrivsMap[key] = append(schemaTablePrivsMap[key], permission.PrivilegeType)
	}
	return schemaTablePrivsMap, err
}

func (p *PostgresManager) GetCreateTableStatement(ctx context.Context, schema, table string) (string, error) {
	errgrp, errctx := errgroup.WithContext(ctx)

	schematable := sqlmanager_shared.SchemaTable{Schema: schema, Table: table}

	var tableSchemas []*pg_queries.GetDatabaseTableSchemasBySchemasAndTablesRow
	errgrp.Go(func() error {
		result, err := p.querier.GetDatabaseTableSchemasBySchemasAndTables(errctx, p.db, []string{schematable.String()})
		if err != nil {
			return fmt.Errorf("unable to generate database table schema: %w", err)
		}
		tableSchemas = result
		return nil
	})
	var tableConstraints []*pg_queries.GetTableConstraintsRow
	errgrp.Go(func() error {
		result, err := p.querier.GetTableConstraints(errctx, p.db, &pg_queries.GetTableConstraintsParams{
			Schema: schema,
			Table:  table,
		})
		if err != nil {
			return fmt.Errorf("unable to generate table constraints: %w", err)
		}
		tableConstraints = result
		return nil
	})
	if err := errgrp.Wait(); err != nil {
		return "", err
	}

	return generateCreateTableStatement(
		schema,
		table,
		tableSchemas,
		tableConstraints,
	), nil
}

func (p *PostgresManager) GetSchemaTableTriggers(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableTrigger, error) {
	if len(tables) == 0 {
		return []*sqlmanager_shared.TableTrigger{}, nil
	}

	combined := make([]string, 0, len(tables))
	for _, t := range tables {
		combined = append(combined, t.String())
	}

	rows, err := p.querier.GetCustomTriggersBySchemaAndTables(ctx, p.db, combined)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return []*sqlmanager_shared.TableTrigger{}, nil
	}

	output := make([]*sqlmanager_shared.TableTrigger, 0, len(rows))
	for _, row := range rows {
		output = append(output, &sqlmanager_shared.TableTrigger{
			Schema:      row.SchemaName,
			Table:       row.TableName,
			TriggerName: row.TriggerName,
			Definition:  wrapPgIdempotentTrigger(row.SchemaName, row.TableName, row.TriggerName, row.Definition),
		})
	}
	return output, nil
}

// Returns ansilary dependencies like sequences, datatypes, functions, etc that are used by tables, but live at the schema level
func (p *PostgresManager) GetSchemaTableDataTypes(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) (*sqlmanager_shared.SchemaTableDataTypeResponse, error) {
	if len(tables) == 0 {
		return &sqlmanager_shared.SchemaTableDataTypeResponse{}, nil
	}

	schemaTablesMap := map[string][]string{}
	for _, t := range tables {
		schemaTablesMap[t.Schema] = append(schemaTablesMap[t.Schema], t.Table)
	}

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(3) // Limit this to effectively one set per schema

	output := &sqlmanager_shared.SchemaTableDataTypeResponse{}
	// Could use a mutex per property, but this is fine for now
	mu := sync.Mutex{}
	for schema, tables := range schemaTablesMap {
		schema := schema
		tables := tables

		errgrp.Go(func() error {
			seqs, err := p.GetSequencesByTables(errctx, schema, tables)
			if err != nil {
				return fmt.Errorf("unable to get postgres custom sequences by tables: %w", err)
			}
			mu.Lock()
			output.Sequences = append(output.Sequences, seqs...)
			mu.Unlock()
			return nil
		})
		errgrp.Go(func() error {
			funcs, err := p.getFunctionsByTables(errctx, schema, tables)
			if err != nil {
				return fmt.Errorf("unable to get postgres custom functions by tables: %w", err)
			}
			mu.Lock()
			output.Functions = append(output.Functions, funcs...)
			mu.Unlock()
			return nil
		})
		errgrp.Go(func() error {
			datatypes, err := p.getDataTypesByTables(errctx, schema, tables)
			if err != nil {
				return fmt.Errorf("unable to get postgres custom data types by tables: %w", err)
			}
			mu.Lock()
			output.Composites = append(output.Composites, datatypes.Composites...)
			output.Enums = append(output.Enums, datatypes.Enums...)
			output.Domains = append(output.Domains, datatypes.Domains...)
			mu.Unlock()
			return nil
		})
	}
	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (p *PostgresManager) GetSequencesByTables(ctx context.Context, schema string, tables []string) ([]*sqlmanager_shared.DataType, error) {
	rows, err := p.querier.GetCustomSequencesBySchemaAndTables(ctx, p.db, &pg_queries.GetCustomSequencesBySchemaAndTablesParams{
		Schema: schema,
		Tables: tables,
	})
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return []*sqlmanager_shared.DataType{}, nil
	}

	output := make([]*sqlmanager_shared.DataType, 0, len(rows))
	for _, row := range rows {
		output = append(output, &sqlmanager_shared.DataType{
			Schema:     row.SchemaName,
			Name:       row.SequenceName,
			Definition: wrapPgIdempotentSequence(row.SchemaName, row.SequenceName, row.Definition),
		})
	}
	return output, nil
}

func (p *PostgresManager) getFunctionsByTables(ctx context.Context, schema string, tables []string) ([]*sqlmanager_shared.DataType, error) {
	rows, err := p.querier.GetCustomFunctionsBySchemaAndTables(ctx, p.db, &pg_queries.GetCustomFunctionsBySchemaAndTablesParams{
		Schema: schema,
		Tables: tables,
	})
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return []*sqlmanager_shared.DataType{}, nil
	}

	output := make([]*sqlmanager_shared.DataType, 0, len(rows))
	for _, row := range rows {
		output = append(output, &sqlmanager_shared.DataType{
			Schema:     row.SchemaName,
			Name:       row.FunctionName,
			Definition: wrapPgIdempotentFunction(row.SchemaName, row.FunctionName, row.FunctionSignature, row.Definition),
		})
	}
	return output, nil
}

type datatypes struct {
	Composites []*sqlmanager_shared.DataType
	Enums      []*sqlmanager_shared.DataType
	Domains    []*sqlmanager_shared.DataType
}

func (p *PostgresManager) getDataTypesByTables(ctx context.Context, schema string, tables []string) (*datatypes, error) {
	rows, err := p.querier.GetDataTypesBySchemaAndTables(ctx, p.db, &pg_queries.GetDataTypesBySchemaAndTablesParams{
		Schema: schema,
		Tables: tables,
	})
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return &datatypes{}, nil
	}

	output := &datatypes{}

	for _, row := range rows {
		dt := &sqlmanager_shared.DataType{
			Schema:     row.SchemaName,
			Name:       row.TypeName,
			Definition: wrapPgIdempotentDataType(row.SchemaName, row.TypeName, row.Definition),
		}
		switch row.Type {
		case "composite":
			output.Composites = append(output.Composites, dt)
		case "domain":
			output.Domains = append(output.Domains, dt)
		case "enum":
			output.Enums = append(output.Enums, dt)
		}
	}
	return output, nil
}

func (p *PostgresManager) GetTableInitStatements(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableInitStatement, error) {
	if len(tables) == 0 {
		return []*sqlmanager_shared.TableInitStatement{}, nil
	}

	combined := []string{}
	schemaset := map[string]struct{}{}
	for _, table := range tables {
		combined = append(combined, table.String())
		schemaset[table.Schema] = struct{}{}
	}
	schemas := []string{}
	for schema := range schemaset {
		schemas = append(schemas, schema)
	}

	errgrp, errctx := errgroup.WithContext(ctx)

	colDefMap := map[string][]*pg_queries.GetDatabaseTableSchemasBySchemasAndTablesRow{}
	errgrp.Go(func() error {
		columnDefs, err := p.querier.GetDatabaseTableSchemasBySchemasAndTables(errctx, p.db, combined)
		if err != nil {
			return err
		}
		for _, columnDefinition := range columnDefs {
			key := sqlmanager_shared.SchemaTable{Schema: columnDefinition.SchemaName, Table: columnDefinition.TableName}
			colDefMap[key.String()] = append(colDefMap[key.String()], columnDefinition)
		}
		return nil
	})

	constraintmap := map[string][]*pg_queries.GetTableConstraintsBySchemaRow{}
	errgrp.Go(func() error {
		constraints, err := p.querier.GetTableConstraintsBySchema(errctx, p.db, schemas) // todo: update this to only grab what is necessary instead of entire schema
		if err != nil {
			return err
		}
		for _, constraint := range constraints {
			key := sqlmanager_shared.SchemaTable{Schema: constraint.SchemaName, Table: constraint.TableName}
			constraintmap[key.String()] = append(constraintmap[key.String()], constraint)
		}
		return nil
	})

	indexmap := map[string][]string{}
	errgrp.Go(func() error {
		idxrecords, err := p.querier.GetIndicesBySchemasAndTables(errctx, p.db, combined)
		if err != nil {
			return err
		}
		for _, record := range idxrecords {
			key := sqlmanager_shared.SchemaTable{Schema: record.SchemaName, Table: record.TableName}
			indexmap[key.String()] = append(indexmap[key.String()], wrapPgIdempotentIndex(record.SchemaName, record.IndexName, record.IndexDefinition))
		}
		return nil
	})

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	output := []*sqlmanager_shared.TableInitStatement{}
	// using input here causes the output to always be consistent
	for _, schematable := range tables {
		key := schematable.String()
		tableData, ok := colDefMap[key]
		if !ok {
			continue
		}
		columns := make([]string, 0, len(tableData))
		for _, record := range tableData {
			record := record
			var seqConfig *SequenceConfiguration
			if record.IdentityGeneration != "" && record.SeqStartValue.Valid && record.SeqMinValue.Valid &&
				record.SeqMaxValue.Valid && record.SeqIncrementBy.Valid && record.SeqCycleOption.Valid && record.SeqCacheValue.Valid {
				seqConfig = &SequenceConfiguration{
					StartValue:  record.SeqStartValue.Int64,
					MinValue:    record.SeqMinValue.Int64,
					MaxValue:    record.SeqMaxValue.Int64,
					IncrementBy: record.SeqIncrementBy.Int64,
					CycleOption: record.SeqCycleOption.Bool,
					CacheValue:  record.SeqCacheValue.Int64,
				}
			}
			columns = append(columns, buildTableCol(&buildTableColRequest{
				ColumnName:    record.ColumnName,
				ColumnDefault: record.ColumnDefault,
				DataType:      record.DataType,
				IsNullable:    record.IsNullable == "YES",
				GeneratedType: record.GeneratedType,
				IsSerial:      record.SequenceType == "SERIAL",
				Sequence:      seqConfig,
				IdentityType:  &record.IdentityGeneration,
			}))
		}

		info := &sqlmanager_shared.TableInitStatement{
			CreateTableStatement: fmt.Sprintf("CREATE TABLE IF NOT EXISTS %q.%q (%s);", tableData[0].SchemaName, tableData[0].TableName, strings.Join(columns, ", ")),
			AlterTableStatements: []*sqlmanager_shared.AlterTableStatement{},
			IndexStatements:      indexmap[key],
		}
		for _, constraint := range constraintmap[key] {
			stmt, err := buildAlterStatementByConstraint(constraint)
			if err != nil {
				return nil, err
			}
			constraintType, err := sqlmanager_shared.ToConstraintType(constraint.ConstraintType)
			if err != nil {
				return nil, err
			}
			info.AlterTableStatements = append(info.AlterTableStatements, &sqlmanager_shared.AlterTableStatement{
				Statement:      wrapPgIdempotentConstraint(constraint.SchemaName, constraint.TableName, constraint.ConstraintName, stmt),
				ConstraintType: constraintType,
			})
		}
		output = append(output, info)
	}
	return output, nil
}

func (p *PostgresManager) GetSchemaInitStatements(
	ctx context.Context,
	tables []*sqlmanager_shared.SchemaTable,
) ([]*sqlmanager_shared.InitSchemaStatements, error) {
	errgrp, errctx := errgroup.WithContext(ctx)
	dataTypeStmts := []string{}
	errgrp.Go(func() error {
		datatypeCfg, err := p.GetSchemaTableDataTypes(errctx, tables)
		if err != nil {
			return fmt.Errorf("unable to retrieve postgres schema table data types: %w", err)
		}
		dataTypeStmts = datatypeCfg.GetStatements()
		return nil
	})

	tableTriggerStmts := []string{}
	errgrp.Go(func() error {
		tableTriggers, err := p.GetSchemaTableTriggers(ctx, tables)
		if err != nil {
			return fmt.Errorf("unable to retrieve postgres schema table triggers: %w", err)
		}
		for _, ttrig := range tableTriggers {
			tableTriggerStmts = append(tableTriggerStmts, ttrig.Definition)
		}
		return nil
	})

	createTables := []string{}
	nonFkAlterStmts := []string{}
	fkAlterStmts := []string{}
	idxStmts := []string{}
	errgrp.Go(func() error {
		initStatementCfgs, err := p.GetTableInitStatements(ctx, tables)
		if err != nil {
			return fmt.Errorf("unable to retrieve postgres schema table create statements: %w", err)
		}
		for _, stmtCfg := range initStatementCfgs {
			createTables = append(createTables, stmtCfg.CreateTableStatement)
			for _, alter := range stmtCfg.AlterTableStatements {
				if alter.ConstraintType == sqlmanager_shared.ForeignConstraintType {
					fkAlterStmts = append(fkAlterStmts, alter.Statement)
				} else {
					nonFkAlterStmts = append(nonFkAlterStmts, alter.Statement)
				}
			}
			idxStmts = append(idxStmts, stmtCfg.IndexStatements...)
		}
		return nil
	})
	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}

	return []*sqlmanager_shared.InitSchemaStatements{
		{Label: "data types", Statements: dataTypeStmts},
		{Label: "create table", Statements: createTables},
		{Label: "non-fk alter table", Statements: nonFkAlterStmts},
		{Label: "fk alter table", Statements: fkAlterStmts},
		{Label: "table index", Statements: idxStmts},
		{Label: "table triggers", Statements: tableTriggerStmts},
	}, nil
}

func wrapPgIdempotentIndex(
	schema,
	constraintname,
	alterStatement string,
) string {
	stmt := fmt.Sprintf(`
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relkind = 'i'
		AND c.relname = '%s'
		AND n.nspname = '%s'
	) THEN
		%s
	END IF;
END $$;
`, constraintname, schema, addSuffixIfNotExist(alterStatement, ";"))
	return strings.TrimSpace(stmt)
}

func wrapPgIdempotentConstraint(
	schema, table,
	constraintName,
	alterStatement string,
) string {
	stmt := fmt.Sprintf(`
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = '%s'
		AND connamespace = '%s'::regnamespace
		AND conrelid = (
			SELECT oid
			FROM pg_class
			WHERE relname = '%s'
			AND relnamespace = '%s'::regnamespace
		)
	) THEN
		%s
	END IF;
END $$;
	`, constraintName, schema, table, schema, addSuffixIfNotExist(alterStatement, ";"))
	return strings.TrimSpace(stmt)
}

func wrapPgIdempotentSequence(
	schema,
	sequenceName,
	createStatement string,
) string {
	stmt := fmt.Sprintf(`
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE c.relkind = 'S'
        AND c.relname = '%s'
        AND n.nspname = '%s'
    ) THEN
        %s
    END IF;
END $$;
`, sequenceName, schema, addSuffixIfNotExist(createStatement, ";"))
	return strings.TrimSpace(stmt)
}

func wrapPgIdempotentTrigger(
	schema,
	tableName,
	triggerName,
	createStatement string,
) string {
	stmt := fmt.Sprintf(`
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger t
        JOIN pg_class c ON c.oid = t.tgrelid
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE t.tgname = '%s'
        AND c.relname = '%s'
        AND n.nspname = '%s'
    ) THEN
        %s
    END IF;
END $$;
`, triggerName, tableName, schema, addSuffixIfNotExist(createStatement, ";"))
	return strings.TrimSpace(stmt)
}

func wrapPgIdempotentFunction(
	schema,
	functionName,
	functionSignature,
	createStatement string,
) string {
	stmt := fmt.Sprintf(`
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_proc p
        JOIN pg_namespace n ON n.oid = p.pronamespace
        WHERE p.proname = '%s'
        AND n.nspname = '%s'
        AND pg_catalog.pg_get_function_identity_arguments(p.oid) = '%s'
    ) THEN
        %s
    END IF;
END $$;
`, functionName, schema, functionSignature, addSuffixIfNotExist(createStatement, ";"))
	return strings.TrimSpace(stmt)
}

func wrapPgIdempotentDataType(
	schema,
	dataTypeName,
	createStatement string,
) string {
	stmt := fmt.Sprintf(`
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_type t
        JOIN pg_namespace n ON n.oid = t.typnamespace
        WHERE t.typname = '%s'
        AND n.nspname = '%s'
    ) THEN
        %s
    END IF;
END $$;
`, dataTypeName, schema, addSuffixIfNotExist(createStatement, ";"))
	return strings.TrimSpace(stmt)
}

//nolint:unparam
func addSuffixIfNotExist(input, suffix string) string {
	if !strings.HasSuffix(input, suffix) {
		return fmt.Sprintf("%s%s", input, suffix)
	}
	return input
}

func buildAlterStatementByConstraint(
	constraint *pg_queries.GetTableConstraintsBySchemaRow,
) (string, error) {
	if constraint == nil {
		return "", errors.New("unable to build alter statement as constraint is nil")
	}
	return fmt.Sprintf(
		"ALTER TABLE %q.%q ADD CONSTRAINT %s %s;",
		constraint.SchemaName, constraint.TableName, constraint.ConstraintName, constraint.ConstraintDefinition,
	), nil
}

// This assumes that the schemas and constraints as for a single table, not an entire db schema
func generateCreateTableStatement(
	schema string,
	table string,
	tableSchemas []*pg_queries.GetDatabaseTableSchemasBySchemasAndTablesRow,
	tableConstraints []*pg_queries.GetTableConstraintsRow,
) string {
	columns := make([]string, len(tableSchemas))
	for idx := range tableSchemas {
		record := tableSchemas[idx]
		var seqConfig *SequenceConfiguration
		if record.IdentityGeneration != "" && record.SeqStartValue.Valid && record.SeqMinValue.Valid &&
			record.SeqMaxValue.Valid && record.SeqIncrementBy.Valid && record.SeqCycleOption.Valid && record.SeqCacheValue.Valid {
			seqConfig = &SequenceConfiguration{
				StartValue:  record.SeqStartValue.Int64,
				MinValue:    record.SeqMinValue.Int64,
				MaxValue:    record.SeqMaxValue.Int64,
				IncrementBy: record.SeqIncrementBy.Int64,
				CycleOption: record.SeqCycleOption.Bool,
				CacheValue:  record.SeqCacheValue.Int64,
			}
		}
		columns[idx] = buildTableCol(&buildTableColRequest{
			ColumnName:    record.ColumnName,
			ColumnDefault: record.ColumnDefault,
			DataType:      record.DataType,
			IsNullable:    record.IsNullable == "YES",
			GeneratedType: record.GeneratedType,
			IsSerial:      record.SequenceType == "SERIAL",
			Sequence:      seqConfig,
			IdentityType:  &record.IdentityGeneration,
		})
	}

	constraints := make([]string, len(tableConstraints))
	for idx := range tableConstraints {
		constraint := tableConstraints[idx]
		constraints[idx] = fmt.Sprintf("CONSTRAINT %s %s", constraint.ConstraintName, constraint.ConstraintDefinition)
	}
	tableDefs := append(columns, constraints...) //nolint:gocritic
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %q.%q (%s);`, schema, table, strings.Join(tableDefs, ", "))
}

type buildTableColRequest struct {
	ColumnName    string
	ColumnDefault string
	DataType      string
	IsNullable    bool
	GeneratedType string
	IsSerial      bool
	IdentityType  *string
	Sequence      *SequenceConfiguration
}

type SequenceConfiguration struct {
	IncrementBy int64
	MinValue    int64
	MaxValue    int64
	StartValue  int64
	CacheValue  int64
	CycleOption bool
}

func (s *SequenceConfiguration) ToGeneratedDefaultIdentity() string {
	return fmt.Sprintf("GENERATED BY DEFAULT AS IDENTITY ( %s )", s.identitySequenceConfiguration())
}
func (s *SequenceConfiguration) ToGeneratedAlwaysIdentity() string {
	return fmt.Sprintf("GENERATED ALWAYS AS IDENTITY ( %s )", s.identitySequenceConfiguration())
}

func (s *SequenceConfiguration) identitySequenceConfiguration() string {
	return fmt.Sprintf("INCREMENT BY %d MINVALUE %d MAXVALUE %d START %d CACHE %d %s",
		s.IncrementBy, s.MinValue, s.MaxValue, s.StartValue, s.CacheValue, s.toCycelText(),
	)
}

func (s *SequenceConfiguration) toCycelText() string {
	if s.CycleOption {
		return "CYCLE"
	}
	return "NO CYCLE"
}

func buildTableCol(record *buildTableColRequest) string {
	pieces := []string{EscapePgColumn(record.ColumnName), record.DataType, buildNullableText(record.IsNullable)}

	if record.IsSerial {
		if record.DataType == "smallint" {
			pieces[1] = "SMALLSERIAL"
		} else if record.DataType == "bigint" {
			pieces[1] = "BIGSERIAL"
		} else {
			pieces[1] = "SERIAL"
		}
	} else if record.IdentityType != nil && *record.IdentityType != "" && record.Sequence != nil {
		if *record.IdentityType == "d" {
			pieces = append(pieces, record.Sequence.ToGeneratedDefaultIdentity())
		} else if *record.IdentityType == "a" {
			pieces = append(pieces, record.Sequence.ToGeneratedAlwaysIdentity())
		}
	} else if record.ColumnDefault != "" {
		if record.GeneratedType == "s" {
			pieces = append(pieces, fmt.Sprintf("GENERATED ALWAYS AS (%s) STORED", record.ColumnDefault))
		} else if record.ColumnDefault != "NULL" {
			pieces = append(pieces, "DEFAULT", record.ColumnDefault)
		}
	}
	return strings.Join(pieces, " ")
}

func buildNullableText(isNullable bool) string {
	if isNullable {
		return "NULL"
	}
	return "NOT NULL"
}

func (p *PostgresManager) BatchExec(ctx context.Context, batchSize int, statements []string, opts *sqlmanager_shared.BatchExecOpts) error {
	for i := 0; i < len(statements); i += batchSize {
		end := i + batchSize
		if end > len(statements) {
			end = len(statements)
		}

		batchCmd := strings.Join(statements[i:end], "\n")
		if opts != nil && opts.Prefix != nil && *opts.Prefix != "" {
			batchCmd = fmt.Sprintf("%s %s", *opts.Prefix, batchCmd)
		}

		_, err := p.db.ExecContext(ctx, batchCmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresManager) Exec(ctx context.Context, statement string) error {
	_, err := p.db.ExecContext(ctx, statement)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresManager) Close() {
	if p.db != nil && p.close != nil {
		p.close()
	}
}

func (p *PostgresManager) GetTableRowCount(
	ctx context.Context,
	schema, table string,
	whereClause *string,
) (int64, error) {
	tableName := sqlmanager_shared.BuildTable(schema, table)
	builder := getGoquDialect()
	sqltable := goqu.I(tableName)

	query := builder.From(sqltable).Select(goqu.COUNT("*"))
	if whereClause != nil && *whereClause != "" {
		query = query.Where(goqu.L(*whereClause))
	}
	sql, _, err := query.ToSQL()
	if err != nil {
		return 0, err
	}
	var count int64
	err = p.db.QueryRowContext(ctx, sql).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, err
}

func getGoquDialect() goqu.DialectWrapper {
	return goqu.Dialect(sqlmanager_shared.DefaultPostgresDriver)
}

func BuildPgTruncateStatement(
	tables []*sqlmanager_shared.SchemaTable,
) (string, error) {
	builder := getGoquDialect()
	gTables := []any{}
	for _, t := range tables {
		gTables = append(gTables, goqu.S(t.Schema).Table(t.Table))
	}
	stmt, _, err := builder.From(gTables...).Truncate().Identity("RESTART").ToSQL()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s;", stmt), nil
}

func BuildPgTruncateCascadeStatement(
	schema string,
	table string,
) (string, error) {
	builder := getGoquDialect()
	sqltable := goqu.S(schema).Table(table)
	stmt, _, err := builder.From(sqltable).Truncate().Cascade().Identity("RESTART").ToSQL()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s;", stmt), nil
}

func EscapePgColumns(cols []string) []string {
	outcols := make([]string, len(cols))
	for idx := range cols {
		outcols[idx] = EscapePgColumn(cols[idx])
	}
	return outcols
}

func EscapePgColumn(col string) string {
	return fmt.Sprintf("%q", col)
}

func BuildPgIdentityColumnResetCurrentSql(
	schema, table, column string,
) string {
	return fmt.Sprintf("SELECT setval(pg_get_serial_sequence('%s.%s', '%s'), COALESCE((SELECT MAX(%q) FROM %q.%q), 1));", schema, table, column, column, schema, table)
}

func BuildPgInsertIdentityAlwaysSql(
	insertQuery string,
) string {
	sqlSplit := strings.Split(insertQuery, ") VALUES (")
	return sqlSplit[0] + ") OVERRIDING SYSTEM VALUE VALUES(" + sqlSplit[1]
}

func BuildPgResetSequenceSql(sequenceName string) string {
	return fmt.Sprintf("ALTER SEQUENCE %s RESTART;", sequenceName)
}

func GetPostgresColumnOverrideAndResetProperties(columnInfo *sqlmanager_shared.DatabaseSchemaRow) (needsOverride, needsReset bool) {
	needsOverride = false
	needsReset = false

	// check if the column is an idenitity type
	if columnInfo.IdentityGeneration != nil && *columnInfo.IdentityGeneration != "" {
		switch *columnInfo.IdentityGeneration {
		case "a": // ALWAYS
			needsOverride = true
			needsReset = true
		case "d": // DEFAULT
			needsReset = true
		}
		return
	}

	// check if column default is sequence
	if columnInfo.ColumnDefault != "" && gotypeutil.CaseInsensitiveContains(columnInfo.ColumnDefault, "nextVal") {
		needsReset = true
		return
	}

	return
}
