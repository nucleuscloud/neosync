package sqlmanager_postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/doug-martin/goqu/v9"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
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

func (p *PostgresManager) GetDatabaseSchema(
	ctx context.Context,
) ([]*sqlmanager_shared.DatabaseSchemaRow, error) {
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
			UpdateAllowed: isColumnUpdateAllowed(
				row.IdentityGeneration,
				row.GeneratedType,
			),
		})
	}
	return result, nil
}

func isColumnUpdateAllowed(identityGeneration, generatedType string) bool {
	// generated always columns cannot be updated, generated always as identity columns cannot be updated
	if identityGeneration == "a" || generatedType == "s" {
		return false
	}
	return true
}

func (p *PostgresManager) GetDatabaseTableSchemasBySchemasAndTables(
	ctx context.Context,
	tables []*sqlmanager_shared.SchemaTable,
) ([]*sqlmanager_shared.DatabaseSchemaRow, error) {
	schemaTables := make([]string, 0, len(tables))
	for _, t := range tables {
		schemaTables = append(schemaTables, t.String())
	}
	rows, err := p.querier.GetDatabaseTableSchemasBySchemasAndTables(ctx, p.db, schemaTables)
	if err != nil {
		return nil, err
	}
	result := []*sqlmanager_shared.DatabaseSchemaRow{}
	for _, row := range rows {
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
			GeneratedType:          sqlmanager_shared.Ptr(row.GeneratedType),
			IdentityGeneration:     sqlmanager_shared.Ptr(row.IdentityGeneration),
			UpdateAllowed: isColumnUpdateAllowed(
				row.IdentityGeneration,
				row.GeneratedType,
			),
		})
	}
	return result, nil
}

func (p *PostgresManager) GetColumnsByTables(
	ctx context.Context,
	tables []*sqlmanager_shared.SchemaTable,
) ([]*sqlmanager_shared.TableColumn, error) {
	schemaTables := make([]string, 0, len(tables))
	for _, t := range tables {
		schemaTables = append(schemaTables, t.String())
	}
	rows, err := p.querier.GetDatabaseTableSchemasBySchemasAndTables(ctx, p.db, schemaTables)
	if err != nil {
		return nil, err
	}
	result := []*sqlmanager_shared.TableColumn{}
	for _, row := range rows {
		var sequenceDefinition *string
		if row.IdentityGeneration != "" && row.SeqStartValue.Valid && row.SeqMinValue.Valid &&
			row.SeqMaxValue.Valid && row.SeqIncrementBy.Valid && row.SeqCycleOption.Valid && row.SeqCacheValue.Valid {
			seqConfig := &SequenceConfiguration{
				StartValue:  row.SeqStartValue.Int64,
				MinValue:    row.SeqMinValue.Int64,
				MaxValue:    row.SeqMaxValue.Int64,
				IncrementBy: row.SeqIncrementBy.Int64,
				CycleOption: row.SeqCycleOption.Bool,
				CacheValue:  row.SeqCacheValue.Int64,
			}
			seqStr := buildSequenceDefinition(row.IdentityGeneration, seqConfig)
			sequenceDefinition = &seqStr
		}
		col := &sqlmanager_shared.TableColumn{
			Schema:             row.SchemaName,
			Table:              row.TableName,
			Name:               row.ColumnName,
			OrdinalPosition:    int(row.OrdinalPosition),
			DataType:           row.DataType,
			IsNullable:         row.IsNullable != "NO",
			ColumnDefault:      row.ColumnDefault,
			GeneratedType:      sqlmanager_shared.Ptr(row.GeneratedType),
			IdentityGeneration: sqlmanager_shared.Ptr(row.IdentityGeneration),
			SequenceDefinition: sequenceDefinition,
			Comment:            sqlmanager_shared.Ptr(row.ColumnComment),
		}
		shouldIncludeOrdinalPosition := true
		col.Fingerprint = sqlmanager_shared.BuildTableColumnFingerprint(col, shouldIncludeOrdinalPosition)
		result = append(result, col)
	}
	return result, nil
}

func (p *PostgresManager) GetTableConstraintsByTables(
	ctx context.Context,
	schema string,
	tables []string,
) (map[string]*sqlmanager_shared.AllTableConstraints, error) {
	if len(tables) == 0 {
		return map[string]*sqlmanager_shared.AllTableConstraints{}, nil
	}
	errgrp, errctx := errgroup.WithContext(ctx)
	var nonFkConstraints []*pg_queries.GetNonForeignKeyTableConstraintsBySchemaAndTablesRow
	var fkConstraints []*pg_queries.GetForeignKeyConstraintsBySchemasAndTablesRow
	errgrp.Go(func() error {
		var err error
		constraints, err := p.querier.GetNonForeignKeyTableConstraintsBySchemaAndTables(
			errctx,
			p.db,
			&pg_queries.GetNonForeignKeyTableConstraintsBySchemaAndTablesParams{
				Schema: schema,
				Tables: tables,
			},
		)
		if err != nil {
			return err
		}
		nonFkConstraints = constraints
		return nil
	})

	errgrp.Go(func() error {
		var err error
		constraints, err := p.querier.GetForeignKeyConstraintsBySchemasAndTables(
			errctx,
			p.db,
			&pg_queries.GetForeignKeyConstraintsBySchemasAndTablesParams{
				Schema: schema,
				Tables: tables,
			},
		)
		if err != nil {
			return err
		}
		fkConstraints = constraints
		return nil
	})

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	result := map[string]*sqlmanager_shared.AllTableConstraints{}

	for _, row := range nonFkConstraints {
		key := sqlmanager_shared.SchemaTable{
			Schema: row.SchemaName,
			Table:  row.TableName,
		}.String()
		if result[key] == nil {
			result[key] = &sqlmanager_shared.AllTableConstraints{
				NonForeignKeyConstraints: []*sqlmanager_shared.NonForeignKeyConstraint{},
				ForeignKeyConstraints:    []*sqlmanager_shared.ForeignKeyConstraint{},
			}
		}
		constraint := &sqlmanager_shared.NonForeignKeyConstraint{
			SchemaName:     row.SchemaName,
			TableName:      row.TableName,
			ConstraintName: row.ConstraintName,
			ConstraintType: row.ConstraintType,
			Columns:        row.ConstraintColumns,
			Definition:     row.ConstraintDefinition,
			Deferrable:     row.Deferrable,
		}
		constraint.Fingerprint = sqlmanager_shared.BuildNonForeignKeyConstraintFingerprint(
			constraint,
		)
		result[key].NonForeignKeyConstraints = append(
			result[key].NonForeignKeyConstraints,
			constraint,
		)
	}

	for _, row := range fkConstraints {
		key := sqlmanager_shared.SchemaTable{
			Schema: row.ReferencingSchema,
			Table:  row.ReferencingTable,
		}.String()
		if result[key] == nil {
			result[key] = &sqlmanager_shared.AllTableConstraints{
				NonForeignKeyConstraints: []*sqlmanager_shared.NonForeignKeyConstraint{},
				ForeignKeyConstraints:    []*sqlmanager_shared.ForeignKeyConstraint{},
			}
		}
		constraint := &sqlmanager_shared.ForeignKeyConstraint{
			ConstraintName:     row.ConstraintName,
			ConstraintType:     "FOREIGN KEY",
			ReferencingSchema:  row.ReferencingSchema,
			ReferencingTable:   row.ReferencingTable,
			ReferencingColumns: row.ReferencingColumns,
			ReferencedSchema:   row.ReferencedSchema,
			ReferencedTable:    row.ReferencedTable,
			ReferencedColumns:  row.ReferencedColumns,
			NotNullable:        row.NotNullable,
			Deferrable:         row.Deferrable,
		}
		constraint.Fingerprint = sqlmanager_shared.BuildForeignKeyConstraintFingerprint(constraint)
		result[key].ForeignKeyConstraints = append(result[key].ForeignKeyConstraints, constraint)
	}
	return result, nil
}

func (p *PostgresManager) GetDataTypesByTables(
	ctx context.Context,
	tables []*sqlmanager_shared.SchemaTable,
) (*sqlmanager_shared.AllTableDataTypes, error) {
	if len(tables) == 0 {
		return &sqlmanager_shared.AllTableDataTypes{}, nil
	}

	tableNames := make([]string, 0, len(tables))
	for _, t := range tables {
		tableNames = append(tableNames, t.String())
	}

	errgrp, errctx := errgroup.WithContext(ctx)

	enumTypes := []*sqlmanager_shared.EnumDataType{}
	errgrp.Go(func() error {
		var err error
		enums, err := p.querier.GetEnumTypesByTables(errctx, p.db, tableNames)
		if err != nil {
			return err
		}
		for _, row := range enums {
			enum := &sqlmanager_shared.EnumDataType{
				Schema: row.Schema,
				Name:   row.Name,
				Values: row.Values,
			}
			enum.Fingerprint = sqlmanager_shared.BuildEnumDataTypeFingerprint(enum)
			enumTypes = append(enumTypes, enum)
		}
		return nil
	})

	compositeTypes := []*sqlmanager_shared.CompositeDataType{}
	errgrp.Go(func() error {
		var err error
		composites, err := p.querier.GetCompositeTypesByTables(errctx, p.db, tableNames)
		if err != nil {
			return err
		}
		for _, row := range composites {
			var attributes []*sqlmanager_shared.CompositeAttribute
			if err := json.Unmarshal(row.Attributes, &attributes); err != nil {
				return err
			}
			composite := &sqlmanager_shared.CompositeDataType{
				Schema:     row.Schema,
				Name:       row.Name,
				Attributes: attributes,
			}
			composite.Fingerprint = sqlmanager_shared.BuildCompositeDataTypeFingerprint(composite)
			compositeTypes = append(compositeTypes, composite)
		}
		return nil
	})

	domainTypes := []*sqlmanager_shared.DomainDataType{}
	errgrp.Go(func() error {
		var err error
		domains, err := p.querier.GetDomainsByTables(errctx, p.db, tableNames)
		if err != nil {
			return err
		}
		for _, row := range domains {
			var constraints []*sqlmanager_shared.DomainConstraint
			if err := json.Unmarshal(row.Constraints, &constraints); err != nil {
				return err
			}
			domain := &sqlmanager_shared.DomainDataType{
				Schema:      row.Schema,
				Name:        row.Name,
				Constraints: constraints,
			}
			domain.Fingerprint = sqlmanager_shared.BuildDomainDataTypeFingerprint(domain)
			domainTypes = append(domainTypes, domain)
		}
		return nil
	})

	schemaTablesMap := map[string][]string{}
	for _, t := range tables {
		schemaTablesMap[t.Schema] = append(schemaTablesMap[t.Schema], t.Table)
	}

	functions := []*sqlmanager_shared.DataType{}
	errgrp.Go(func() error {
		for schema, tables := range schemaTablesMap {
			rows, err := p.querier.GetCustomFunctionsBySchemaAndTables(ctx, p.db, &pg_queries.GetCustomFunctionsBySchemaAndTablesParams{
				Schema: schema,
				Tables: tables,
			})
			if err != nil && !neosyncdb.IsNoRows(err) {
				return err
			} else if err != nil && neosyncdb.IsNoRows(err) {
				return nil
			}
			for _, row := range rows {
				function := &sqlmanager_shared.DataType{
					Schema:     row.SchemaName,
					Name:       row.FunctionName,
					Definition: row.Definition,
				}
				function.Fingerprint = sqlmanager_shared.BuildFingerprint(function.Schema, function.Name, function.Definition)
				functions = append(functions, function)
			}
		}
		return nil
	})

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	return &sqlmanager_shared.AllTableDataTypes{
		Functions:  functions,
		Enums:      enumTypes,
		Composites: compositeTypes,
		Domains:    domainTypes,
	}, nil
}

func (p *PostgresManager) GetAllSchemas(ctx context.Context) ([]*sqlmanager_shared.DatabaseSchemaNameRow, error) {
	rows, err := p.querier.GetAllSchemas(ctx, p.db)
	if err != nil {
		return nil, err
	}
	result := make([]*sqlmanager_shared.DatabaseSchemaNameRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, &sqlmanager_shared.DatabaseSchemaNameRow{
			SchemaName: row,
		})
	}
	return result, nil
}

func (p *PostgresManager) GetAllTables(
	ctx context.Context,
) ([]*sqlmanager_shared.DatabaseTableRow, error) {
	rows, err := p.querier.GetAllTables(ctx, p.db)
	if err != nil {
		return nil, err
	}
	result := make([]*sqlmanager_shared.DatabaseTableRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, &sqlmanager_shared.DatabaseTableRow{
			SchemaName: row.TableSchema,
			TableName:  row.TableName,
		})
	}
	return result, nil
}

// returns: {public.users: { id: struct{}{}, created_at: struct{}{}}}
func (p *PostgresManager) GetSchemaColumnMap(
	ctx context.Context,
) (map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow, error) {
	dbSchemas, err := p.GetDatabaseSchema(ctx)
	if err != nil {
		return nil, err
	}
	result := sqlmanager_shared.GetUniqueSchemaColMappings(dbSchemas)
	return result, nil
}

func (p *PostgresManager) GetTableConstraintsBySchema(
	ctx context.Context,
	schemas []string,
) (*sqlmanager_shared.TableConstraints, error) {
	if len(schemas) == 0 {
		return &sqlmanager_shared.TableConstraints{}, nil
	}
	errgrp, errctx := errgroup.WithContext(ctx)
	nonFkConstraints := []*pg_queries.GetNonForeignKeyTableConstraintsBySchemaRow{}
	errgrp.Go(func() error {
		rows, err := p.querier.GetNonForeignKeyTableConstraintsBySchema(ctx, p.db, schemas)
		if err != nil && !neosyncdb.IsNoRows(err) {
			return err
		} else if err != nil && neosyncdb.IsNoRows(err) {
			return nil
		}
		nonFkConstraints = rows
		return nil
	})

	fkConstraints := []*pg_queries.GetForeignKeyConstraintsBySchemasRow{}
	errgrp.Go(func() error {
		fks, err := p.querier.GetForeignKeyConstraintsBySchemas(errctx, p.db, schemas)
		if err != nil {
			return err
		}
		fkConstraints = fks
		return nil
	})

	uniqueIndexes := []*pg_queries.GetUniqueIndexesBySchemaRow{}
	errgrp.Go(func() error {
		indexes, err := p.querier.GetUniqueIndexesBySchema(errctx, p.db, schemas)
		if err != nil {
			return err
		}
		uniqueIndexes = indexes
		return nil
	})

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}
	primaryKeyMap := map[string][]string{}
	uniqueConstraintsMap := map[string][][]string{}
	for _, row := range nonFkConstraints {
		tableName := sqlmanager_shared.BuildTable(row.SchemaName, row.TableName)
		switch row.ConstraintType {
		case "p":
			if _, exists := primaryKeyMap[tableName]; !exists {
				primaryKeyMap[tableName] = []string{}
			}
			primaryKeyMap[tableName] = append(
				primaryKeyMap[tableName],
				sqlmanager_shared.DedupeSlice(row.ConstraintColumns)...)
		case "u":
			columns := sqlmanager_shared.DedupeSlice(row.ConstraintColumns)
			uniqueConstraintsMap[tableName] = append(uniqueConstraintsMap[tableName], columns)
		}
	}

	foreignKeyMap := map[string][]*sqlmanager_shared.ForeignConstraint{}
	for _, row := range fkConstraints {
		tableName := sqlmanager_shared.BuildTable(row.ReferencingSchema, row.ReferencingTable)
		if len(row.ReferencingColumns) != len(row.ReferencedColumns) {
			return nil, fmt.Errorf(
				"length of columns was not equal to length of foreign key cols: %d %d",
				len(row.ReferencingColumns),
				len(row.ReferencedColumns),
			)
		}
		if len(row.ReferencingColumns) != len(row.NotNullable) {
			return nil, fmt.Errorf(
				"length of columns was not equal to length of not nullable cols: %d %d",
				len(row.ReferencingColumns),
				len(row.NotNullable),
			)
		}

		foreignKeyMap[tableName] = append(
			foreignKeyMap[tableName],
			&sqlmanager_shared.ForeignConstraint{
				Columns:     row.ReferencingColumns,
				NotNullable: row.NotNullable,
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table: sqlmanager_shared.BuildTable(
						row.ReferencedSchema,
						row.ReferencedTable,
					),
					Columns: row.ReferencedColumns,
				},
			},
		)
	}

	uniqueIndexesMap := map[string][][]string{}
	for _, row := range uniqueIndexes {
		tableName := sqlmanager_shared.BuildTable(row.TableSchema, row.TableName)
		uniqueIndexesMap[tableName] = append(uniqueIndexesMap[tableName], row.IndexColumns)
	}

	return &sqlmanager_shared.TableConstraints{
		ForeignKeyConstraints: foreignKeyMap,
		PrimaryKeyConstraints: primaryKeyMap,
		UniqueConstraints:     uniqueConstraintsMap,
		UniqueIndexes:         uniqueIndexesMap,
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

func (p *PostgresManager) GetSchemaTableTriggers(
	ctx context.Context,
	tables []*sqlmanager_shared.SchemaTable,
) ([]*sqlmanager_shared.TableTrigger, error) {
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
		trigger := &sqlmanager_shared.TableTrigger{
			Schema:      row.SchemaName,
			Table:       row.TableName,
			TriggerName: row.TriggerName,
			Definition: wrapPgIdempotentTrigger(
				row.SchemaName,
				row.TableName,
				row.TriggerName,
				row.Definition,
			),
		}
		trigger.Fingerprint = sqlmanager_shared.BuildTriggerFingerprint(trigger)
		output = append(output, trigger)
	}
	return output, nil
}

// Returns ansilary dependencies like sequences, datatypes, functions, etc that are used by tables, but live at the schema level
func (p *PostgresManager) GetSchemaTableDataTypes(
	ctx context.Context,
	tables []*sqlmanager_shared.SchemaTable,
) (*sqlmanager_shared.SchemaTableDataTypeResponse, error) {
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

func (p *PostgresManager) GetSequencesByTables(
	ctx context.Context,
	schema string,
	tables []string,
) ([]*sqlmanager_shared.DataType, error) {
	rows, err := p.querier.GetCustomSequencesBySchemaAndTables(
		ctx,
		p.db,
		&pg_queries.GetCustomSequencesBySchemaAndTablesParams{
			Schema: schema,
			Tables: tables,
		},
	)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return []*sqlmanager_shared.DataType{}, nil
	}

	output := make([]*sqlmanager_shared.DataType, 0, len(rows))
	for _, row := range rows {
		seq := &sqlmanager_shared.DataType{
			Schema:     row.SchemaName,
			Name:       row.SequenceName,
			Definition: wrapPgIdempotentSequence(row.SchemaName, row.SequenceName, row.Definition),
		}
		seq.Fingerprint = sqlmanager_shared.BuildFingerprint(seq.Schema, seq.Name, seq.Definition)
		output = append(output, seq)
	}
	return output, nil
}

func (p *PostgresManager) getExtensionsBySchemas(
	ctx context.Context,
	schemas []string,
) ([]*sqlmanager_shared.ExtensionDataType, error) {
	rows, err := p.querier.GetExtensionsBySchemas(ctx, p.db, schemas)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return []*sqlmanager_shared.ExtensionDataType{}, nil
	}

	output := make([]*sqlmanager_shared.ExtensionDataType, 0, len(rows))
	for _, row := range rows {
		output = append(output, &sqlmanager_shared.ExtensionDataType{
			Name: row.ExtensionName,
			Definition: wrapPgIdempotentExtension(
				row.SchemaName,
				row.ExtensionName,
				row.InstalledVersion,
			),
		})
	}
	return output, nil
}

func (p *PostgresManager) getFunctionsByTables(
	ctx context.Context,
	schema string,
	tables []string,
) ([]*sqlmanager_shared.DataType, error) {
	rows, err := p.querier.GetCustomFunctionsBySchemaAndTables(
		ctx,
		p.db,
		&pg_queries.GetCustomFunctionsBySchemaAndTablesParams{
			Schema: schema,
			Tables: tables,
		},
	)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return []*sqlmanager_shared.DataType{}, nil
	}

	output := make([]*sqlmanager_shared.DataType, 0, len(rows))
	for _, row := range rows {
		function := &sqlmanager_shared.DataType{
			Schema: row.SchemaName,
			Name:   row.FunctionName,
			Definition: wrapPgIdempotentFunction(
				row.SchemaName,
				row.FunctionName,
				row.FunctionSignature,
				row.Definition,
			),
		}
		function.Fingerprint = sqlmanager_shared.BuildFingerprint(
			function.Schema,
			function.Name,
			function.Definition,
		)
		output = append(output, function)
	}
	return output, nil
}

type datatypes struct {
	Composites []*sqlmanager_shared.DataType
	Enums      []*sqlmanager_shared.DataType
	Domains    []*sqlmanager_shared.DataType
}

func (p *PostgresManager) getDataTypesByTables(
	ctx context.Context,
	schema string,
	tables []string,
) (*datatypes, error) {
	rows, err := p.querier.GetDataTypesBySchemaAndTables(
		ctx,
		p.db,
		&pg_queries.GetDataTypesBySchemaAndTablesParams{
			Schema: schema,
			Tables: tables,
		},
	)
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
		dt.Fingerprint = sqlmanager_shared.BuildFingerprint(dt.Schema, dt.Name, dt.Definition)
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

func (p *PostgresManager) GetTableInitStatements(
	ctx context.Context,
	tables []*sqlmanager_shared.SchemaTable,
) ([]*sqlmanager_shared.TableInitStatement, error) {
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
		columnDefs, err := p.querier.GetDatabaseTableSchemasBySchemasAndTables(
			errctx,
			p.db,
			combined,
		)
		if err != nil {
			return err
		}
		for _, columnDefinition := range columnDefs {
			key := sqlmanager_shared.SchemaTable{
				Schema: columnDefinition.SchemaName,
				Table:  columnDefinition.TableName,
			}
			colDefMap[key.String()] = append(colDefMap[key.String()], columnDefinition)
		}
		return nil
	})

	constraintmap := map[string][]*pg_queries.GetNonForeignKeyTableConstraintsBySchemaRow{}
	errgrp.Go(func() error {
		constraints, err := p.querier.GetNonForeignKeyTableConstraintsBySchema(
			errctx,
			p.db,
			schemas,
		) // todo: update this to only grab what is necessary instead of entire schema
		if err != nil {
			return err
		}
		for _, constraint := range constraints {
			key := sqlmanager_shared.SchemaTable{
				Schema: constraint.SchemaName,
				Table:  constraint.TableName,
			}
			constraintmap[key.String()] = append(constraintmap[key.String()], constraint)
		}
		return nil
	})

	fkConstraintMap := map[string][]*pg_queries.GetForeignKeyConstraintsBySchemasRow{}
	errgrp.Go(func() error {
		fkConstraints, err := p.querier.GetForeignKeyConstraintsBySchemas(errctx, p.db, schemas)
		if err != nil {
			return err
		}
		for _, constraint := range fkConstraints {
			key := sqlmanager_shared.SchemaTable{
				Schema: constraint.ReferencingSchema,
				Table:  constraint.ReferencingTable,
			}
			fkConstraintMap[key.String()] = append(fkConstraintMap[key.String()], constraint)
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
			indexmap[key.String()] = append(
				indexmap[key.String()],
				wrapPgIdempotentIndex(record.SchemaName, record.IndexName, record.IndexDefinition),
			)
		}
		return nil
	})

	mu := sync.Mutex{}
	partitionTables := map[string]*pg_queries.GetPartitionedTablesBySchemaRow{}
	partitionHierarchy := map[string][]*pg_queries.GetPartitionHierarchyByTableRow{}
	errgrp.Go(func() error {
		partitiontables, err := p.querier.GetPartitionedTablesBySchema(errctx, p.db, schemas)
		if err != nil {
			return err
		}
		for _, record := range partitiontables {
			key := sqlmanager_shared.SchemaTable{Schema: record.SchemaName, Table: record.TableName}
			mu.Lock()
			partitionTables[key.String()] = record
			mu.Unlock()
			if !record.IsPartitioned {
				ks := key.String()
				errgrp.Go(func() error {
					partitionhierarchy, err := p.querier.GetPartitionHierarchyByTable(
						errctx,
						p.db,
						ks,
					)
					if err != nil {
						return err
					}
					mu.Lock()
					partitionHierarchy[ks] = partitionhierarchy
					mu.Unlock()
					return nil
				})
			}
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
			var seqDefinition *string
			if record.IdentityGeneration != "" &&
				record.SeqStartValue.Valid &&
				record.SeqMinValue.Valid &&
				record.SeqMaxValue.Valid &&
				record.SeqIncrementBy.Valid &&
				record.SeqCycleOption.Valid &&
				record.SeqCacheValue.Valid {
				seqConfig := &SequenceConfiguration{
					StartValue:  record.SeqStartValue.Int64,
					MinValue:    record.SeqMinValue.Int64,
					MaxValue:    record.SeqMaxValue.Int64,
					IncrementBy: record.SeqIncrementBy.Int64,
					CycleOption: record.SeqCycleOption.Bool,
					CacheValue:  record.SeqCacheValue.Int64,
				}
				seqStr := buildSequenceDefinition(record.IdentityGeneration, seqConfig)
				seqDefinition = &seqStr
			}
			columns = append(columns, buildTableCol(&buildTableColRequest{
				ColumnName:         record.ColumnName,
				ColumnDefault:      record.ColumnDefault,
				DataType:           record.DataType,
				IsNullable:         record.IsNullable == "YES",
				GeneratedType:      record.GeneratedType,
				SequenceDefinition: seqDefinition,
			}))
		}

		partition, ok := partitionTables[key]
		partitionKey := ""
		if ok && !partition.IsPartitioned && partition.PartitionKey != "" {
			partitionKey = fmt.Sprintf(" PARTITION BY %s", partition.PartitionKey)
		}
		info := &sqlmanager_shared.TableInitStatement{
			CreateTableStatement: fmt.Sprintf(
				"CREATE TABLE IF NOT EXISTS %q.%q (%s)%s;",
				tableData[0].SchemaName,
				tableData[0].TableName,
				strings.Join(columns, ", "),
				partitionKey,
			),
			AlterTableStatements: []*sqlmanager_shared.AlterTableStatement{},
			IndexStatements:      indexmap[key],
			PartitionStatements:  []string{},
		}
		for _, constraint := range constraintmap[key] {
			stmt, err := buildAlterStatementByConstraint(constraint)
			if err != nil {
				return nil, err
			}
			constraintType, err := sqlmanager_shared.ToConstraintType(constraint.ConstraintType)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to convert constraint type '%s': %w",
					constraint.ConstraintType,
					err,
				)
			}
			info.AlterTableStatements = append(
				info.AlterTableStatements,
				&sqlmanager_shared.AlterTableStatement{
					Statement: wrapPgIdempotentConstraint(
						constraint.SchemaName,
						constraint.TableName,
						constraint.ConstraintName,
						stmt,
					),
					ConstraintType: constraintType,
				},
			)
		}
		for _, constraint := range fkConstraintMap[key] {
			stmt, err := buildAlterStatementByForeignKeyConstraint(constraint)
			if err != nil {
				return nil, err
			}
			info.AlterTableStatements = append(
				info.AlterTableStatements,
				&sqlmanager_shared.AlterTableStatement{
					Statement: wrapPgIdempotentConstraint(
						constraint.ReferencingSchema,
						constraint.ReferencingTable,
						constraint.ConstraintName,
						stmt,
					),
					ConstraintType: sqlmanager_shared.ForeignConstraintType,
				},
			)
		}
		for _, partition := range partitionHierarchy[key] {
			if !partition.ParentSchemaName.Valid || !partition.ParentTableName.Valid {
				// skip root table
				continue
			}
			p, ok := partitionTables[partition.SchemaName+"."+partition.TableName]
			partitionKey := ""
			if ok && p.IsPartitioned && p.PartitionKey != "" {
				partitionKey = fmt.Sprintf(" PARTITION BY %s", p.PartitionKey)
			}
			info.PartitionStatements = append(
				info.PartitionStatements,
				fmt.Sprintf(
					"CREATE TABLE IF NOT EXISTS %q.%q PARTITION OF %q.%q %s %s;",
					partition.SchemaName,
					partition.TableName,
					partition.ParentSchemaName.String,
					partition.ParentTableName.String,
					partition.PartitionBound,
					partitionKey,
				),
			)
		}
		output = append(output, info)
	}
	return output, nil
}

func (p *PostgresManager) GetSchemaInitStatements(
	ctx context.Context,
	tables []*sqlmanager_shared.SchemaTable,
) ([]*sqlmanager_shared.InitSchemaStatements, error) {
	uniqueSchemas := map[string]struct{}{}
	for _, table := range tables {
		uniqueSchemas[table.Schema] = struct{}{}
	}
	schemas := []string{}
	for schema := range uniqueSchemas {
		schemas = append(schemas, schema)
	}
	errgrp, errctx := errgroup.WithContext(ctx)

	schemaStmts := []string{}
	errgrp.Go(func() error {
		for schema := range uniqueSchemas {
			schemaStmts = append(
				schemaStmts,
				fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %q;", schema),
			)
		}
		return nil
	})

	datatypes := &sqlmanager_shared.SchemaTableDataTypeResponse{}
	dataTypeStmts := []string{}
	errgrp.Go(func() error {
		datatypeCfg, err := p.GetSchemaTableDataTypes(errctx, tables)
		if err != nil {
			return fmt.Errorf("unable to retrieve postgres schema table data types: %w", err)
		}
		dataTypeStmts = datatypeCfg.GetStatements()
		datatypes = datatypeCfg
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

	extensionStmts := []string{}
	errgrp.Go(func() error {
		extensions, err := p.getExtensionsBySchemas(errctx, schemas)
		if err != nil {
			return fmt.Errorf("unable to get postgres extensions: %w", err)
		}
		for _, extension := range extensions {
			extensionStmts = append(extensionStmts, extension.Definition)
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
			createTables = append(createTables, stmtCfg.PartitionStatements...)
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

	schematables := []string{}
	for _, table := range tables {
		schematables = append(schematables, table.String())
	}
	sequenceOwnerStmts := []string{}
	errgrp.Go(func() error {
		sequences, err := p.querier.GetSequencesOwnedByTables(errctx, p.db, schematables)
		if err != nil {
			return fmt.Errorf("unable to retrieve postgres schema table sequences: %w", err)
		}
		for _, seq := range sequences {
			sequenceOwnerStmts = append(sequenceOwnerStmts, BuildSequencOwnerStatement(seq))
		}
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}

	additionalSchemaStmts := getSchemaCreationStatementsFromDataTypes(tables, datatypes)
	schemaStmts = append(schemaStmts, additionalSchemaStmts...)

	return []*sqlmanager_shared.InitSchemaStatements{
		{Label: sqlmanager_shared.SchemasLabel, Statements: schemaStmts},
		{Label: sqlmanager_shared.ExtensionsLabel, Statements: extensionStmts},
		{Label: "data types", Statements: dataTypeStmts},
		{Label: sqlmanager_shared.CreateTablesLabel, Statements: createTables},
		{Label: "non-fk alter table", Statements: nonFkAlterStmts},
		{Label: "table index", Statements: idxStmts},
		{Label: "fk alter table", Statements: fkAlterStmts},
		{Label: "table triggers", Statements: tableTriggerStmts},
		{Label: "sequence owner", Statements: sequenceOwnerStmts},
	}, nil
}

func (p *PostgresManager) BatchExec(
	ctx context.Context,
	batchSize int,
	statements []string,
	opts *sqlmanager_shared.BatchExecOpts,
) error {
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
	compiledSql, _, err := query.ToSQL()
	if err != nil {
		return 0, err
	}
	var count int64
	err = p.db.QueryRowContext(ctx, compiledSql).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, err
}
