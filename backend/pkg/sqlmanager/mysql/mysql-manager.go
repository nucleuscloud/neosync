package sqlmanager_mysql

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/doug-martin/goqu/v9"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
	"golang.org/x/sync/errgroup"
)

const (
	columnDefaultDefault = "Default"
	columnDefaultString  = "String"

	SchemasLabel = "schemas"
)

type MysqlManager struct {
	querier mysql_queries.Querier
	pool    mysql_queries.DBTX
	close   func()
}

func NewManager(querier mysql_queries.Querier, pool mysql_queries.DBTX, closer func()) *MysqlManager {
	return &MysqlManager{querier: querier, pool: pool, close: closer}
}

func (m *MysqlManager) GetDatabaseSchema(ctx context.Context) ([]*sqlmanager_shared.DatabaseSchemaRow, error) {
	dbSchemas, err := m.querier.GetDatabaseSchema(ctx, m.pool)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return []*sqlmanager_shared.DatabaseSchemaRow{}, nil
	}
	result := []*sqlmanager_shared.DatabaseSchemaRow{}
	for _, row := range dbSchemas {
		var generatedType *string
		if row.Extra.Valid && strings.Contains(row.Extra.String, "GENERATED") && !strings.Contains(row.Extra.String, "DEFAULT_GENERATED") {
			generatedTypeCopy := row.Extra.String
			generatedType = &generatedTypeCopy
		}

		columnDefaultStr, err := convertUInt8ToString(row.ColumnDefault)
		if err != nil {
			return nil, err
		}

		var columnDefaultType *string
		if row.Extra.Valid && columnDefaultStr != "" && row.Extra.String == "" {
			val := columnDefaultString // With this type columnDefaultStr will be surrounded by quotes when translated to SQL
			columnDefaultType = &val
		} else if row.Extra.Valid && columnDefaultStr != "" && row.Extra.String != "" {
			val := columnDefaultDefault // With this type columnDefaultStr will be surrounded by parentheses when translated to SQL
			columnDefaultType = &val
		}

		charMaxLength := -1
		if row.CharacterMaximumLength.Valid {
			charMaxLength = int(row.CharacterMaximumLength.Int64)
		}
		numericPrecision := -1
		if row.NumericPrecision.Valid {
			numericPrecision = int(row.NumericPrecision.Int64)
		}
		numericScale := -1
		if row.NumericScale.Valid {
			numericScale = int(row.NumericScale.Int64)
		}
		// Note: there is a slight mismatch here between how we bring this data in to be surfaced vs how we utilize it when building the init table statements.
		// They seem to be disconnected however
		var identityGeneration *string
		if row.Extra.Valid && row.Extra.String == "auto_increment" {
			val := row.Extra.String
			identityGeneration = &val
		}
		result = append(result, &sqlmanager_shared.DatabaseSchemaRow{
			TableSchema:            row.TableSchema,
			TableName:              row.TableName,
			ColumnName:             row.ColumnName,
			DataType:               row.DataType,
			ColumnDefault:          columnDefaultStr,
			ColumnDefaultType:      columnDefaultType,
			IsNullable:             row.IsNullable != "NO",
			GeneratedType:          generatedType,
			CharacterMaximumLength: charMaxLength,
			NumericPrecision:       numericPrecision,
			NumericScale:           numericScale,
			OrdinalPosition:        int(row.OrdinalPosition),
			IdentityGeneration:     identityGeneration,
		})
	}
	return result, nil
}

func (m *MysqlManager) GetDatabaseTableSchemasBySchemasAndTables(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.DatabaseSchemaRow, error) {
	if len(tables) == 0 {
		return []*sqlmanager_shared.DatabaseSchemaRow{}, nil
	}

	schemaset := map[string][]string{}
	for _, table := range tables {
		schemaset[table.Schema] = append(schemaset[table.Schema], table.Table)
	}

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(5)

	dbSchemas := map[string][]*mysql_queries.GetDatabaseTableSchemasBySchemasAndTablesRow{}
	var colDefMapMu sync.Mutex
	for schema, tables := range schemaset {
		errgrp.Go(func() error {
			columnDefs, err := m.querier.GetDatabaseTableSchemasBySchemasAndTables(errctx, m.pool, &mysql_queries.GetDatabaseTableSchemasBySchemasAndTablesParams{
				Schema: schema,
				Tables: tables,
			})
			if err != nil {
				return err
			}
			colDefMapMu.Lock()
			defer colDefMapMu.Unlock()
			for _, columnDefinition := range columnDefs {
				key := sqlmanager_shared.SchemaTable{Schema: columnDefinition.SchemaName, Table: columnDefinition.TableName}
				dbSchemas[key.String()] = append(dbSchemas[key.String()], columnDefinition)
			}
			return nil
		})
	}
	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	result := []*sqlmanager_shared.DatabaseSchemaRow{}
	for _, rows := range dbSchemas {
		for _, row := range rows {
			var generatedType *string
			if row.IdentityGeneration.Valid && strings.Contains(row.IdentityGeneration.String, "GENERATED") && !strings.Contains(row.IdentityGeneration.String, "DEFAULT_GENERATED") {
				generatedTypeCopy := row.IdentityGeneration.String
				generatedType = &generatedTypeCopy
			}

			columnDefaultStr, err := convertUInt8ToString(row.ColumnDefault)
			if err != nil {
				return nil, err
			}

			var columnDefaultType *string
			if row.IdentityGeneration.Valid && columnDefaultStr != "" && row.IdentityGeneration.String == "" {
				val := columnDefaultString // With this type columnDefaultStr will be surrounded by quotes when translated to SQL
				columnDefaultType = &val
			} else if row.IdentityGeneration.Valid && columnDefaultStr != "" && row.IdentityGeneration.String != "" {
				val := columnDefaultDefault // With this type columnDefaultStr will be surrounded by parentheses when translated to SQL
				columnDefaultType = &val
			}
			if (!row.ColumnDefaultRaw.Valid || (row.ColumnDefaultRaw.Valid && row.ColumnDefaultRaw.String == "" || row.ColumnDefaultRaw.String == "NULL")) && row.IsNullable == 1 {
				columnDefaultStr = ""
				columnDefaultType = nil
			}
			fmt.Println("FOO", columnDefaultStr, columnDefaultType)

			// Note: there is a slight mismatch here between how we bring this data in to be surfaced vs how we utilize it when building the init table statements.
			// They seem to be disconnected however
			var identityGeneration *string
			if row.IdentityGeneration.Valid && row.IdentityGeneration.String == "auto_increment" {
				val := row.IdentityGeneration.String
				identityGeneration = &val
			}
			result = append(result, &sqlmanager_shared.DatabaseSchemaRow{
				TableSchema:            row.SchemaName,
				TableName:              row.TableName,
				ColumnName:             row.ColumnName,
				DataType:               row.DataType,
				ColumnDefault:          columnDefaultStr,
				ColumnDefaultType:      columnDefaultType,
				IsNullable:             row.IsNullable == 1,
				GeneratedType:          generatedType,
				CharacterMaximumLength: int(row.CharacterMaximumLength),
				NumericPrecision:       int(row.NumericPrecision),
				NumericScale:           int(row.NumericScale),
				OrdinalPosition:        int(row.OrdinalPosition),
				IdentityGeneration:     identityGeneration,
			})
		}
	}
	return result, nil
}

// returns: {public.users: { id: struct{}{}, created_at: struct{}{}}}
func (m *MysqlManager) GetSchemaColumnMap(ctx context.Context) (map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow, error) {
	dbSchemas, err := m.GetDatabaseSchema(ctx)
	if err != nil {
		return nil, err
	}
	result := sqlmanager_shared.GetUniqueSchemaColMappings(dbSchemas)
	return result, nil
}

func (m *MysqlManager) GetTableConstraintsBySchema(ctx context.Context, schemas []string) (*sqlmanager_shared.TableConstraints, error) {
	if len(schemas) == 0 {
		return &sqlmanager_shared.TableConstraints{}, nil
	}

	rows, err := m.querier.GetTableConstraintsBySchemas(ctx, m.pool, schemas)
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
		constraintCols, err := jsonRawToSlice[string](row.ConstraintColumns)
		if err != nil {
			return nil, err
		}
		switch row.ConstraintType {
		case "FOREIGN KEY":
			fkCols, err := jsonRawToSlice[string](row.ReferencedColumnNames)
			if err != nil {
				return nil, err
			}
			notNullableInts, err := jsonRawToSlice[int](row.NotNullable)
			if err != nil {
				return nil, err
			}
			notNullable := []bool{}
			for _, notNullableInt := range notNullableInts {
				notNullable = append(notNullable, notNullableInt == 1)
			}
			if len(constraintCols) != len(fkCols) {
				return nil, fmt.Errorf("length of columns was not equal to length of foreign key cols: %d %d", len(constraintCols), len(fkCols))
			}
			if len(constraintCols) != len(notNullable) {
				return nil, fmt.Errorf("length of columns was not equal to length of not nullable cols: %d %d", len(constraintCols), len(notNullable))
			}

			foreignKeyMap[tableName] = append(foreignKeyMap[tableName], &sqlmanager_shared.ForeignConstraint{
				Columns:     constraintCols,
				NotNullable: notNullable,
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   sqlmanager_shared.BuildTable(row.ReferencedSchemaName, row.ReferencedTableName),
					Columns: fkCols,
				},
			})
		case "PRIMARY KEY":
			if _, exists := primaryKeyMap[tableName]; !exists {
				primaryKeyMap[tableName] = []string{}
			}
			primaryKeyMap[tableName] = append(primaryKeyMap[tableName], sqlmanager_shared.DedupeSlice(constraintCols)...)
		case "UNIQUE":
			columns := sqlmanager_shared.DedupeSlice(constraintCols)
			uniqueConstraintsMap[tableName] = append(uniqueConstraintsMap[tableName], columns)
		}
	}

	return &sqlmanager_shared.TableConstraints{
		ForeignKeyConstraints: foreignKeyMap,
		PrimaryKeyConstraints: primaryKeyMap,
		UniqueConstraints:     uniqueConstraintsMap,
		// there is no real distinction between unique indexes and unique constraints in mysql
		UniqueIndexes: map[string][][]string{},
	}, nil
}

func jsonRawToSlice[T any](j json.RawMessage) ([]T, error) {
	elements := []T{}
	if j == nil {
		return elements, nil
	}
	if err := json.Unmarshal(j, &elements); err != nil {
		return nil, err
	}
	return elements, nil
}

func (m *MysqlManager) GetRolePermissionsMap(ctx context.Context) (map[string][]string, error) {
	rows, err := m.querier.GetMysqlRolePermissions(ctx, m.pool)
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

func (m *MysqlManager) GetTableInitStatements(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableInitStatement, error) {
	if len(tables) == 0 {
		return []*sqlmanager_shared.TableInitStatement{}, nil
	}

	schemaset := map[string][]string{}
	for _, table := range tables {
		schemaset[table.Schema] = append(schemaset[table.Schema], table.Table)
	}

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(5)

	colDefMap := map[string][]*mysql_queries.GetDatabaseTableSchemasBySchemasAndTablesRow{}
	var colDefMapMu sync.Mutex
	for schema, tables := range schemaset {
		errgrp.Go(func() error {
			columnDefs, err := m.querier.GetDatabaseTableSchemasBySchemasAndTables(errctx, m.pool, &mysql_queries.GetDatabaseTableSchemasBySchemasAndTablesParams{
				Schema: schema,
				Tables: tables,
			})
			if err != nil {
				return fmt.Errorf("failed to build mysql database table schemas by schemas and tables: %w", err)
			}
			colDefMapMu.Lock()
			defer colDefMapMu.Unlock()
			for _, columnDefinition := range columnDefs {
				key := sqlmanager_shared.SchemaTable{Schema: columnDefinition.SchemaName, Table: columnDefinition.TableName}
				colDefMap[key.String()] = append(colDefMap[key.String()], columnDefinition)
			}
			return nil
		})
	}

	constraintmap := map[string][]*mysql_queries.GetTableConstraintsRow{}
	var constraintMapMu sync.Mutex
	for schema, tables := range schemaset {
		errgrp.Go(func() error {
			constraints, err := m.querier.GetTableConstraints(errctx, m.pool, &mysql_queries.GetTableConstraintsParams{
				Schema: schema,
				Tables: tables,
			})
			if err != nil {
				return fmt.Errorf("failed to build mysql table constraints: %w", err)
			}
			constraintMapMu.Lock()
			defer constraintMapMu.Unlock()
			for _, constraint := range constraints {
				key := sqlmanager_shared.SchemaTable{Schema: constraint.SchemaName, Table: constraint.TableName}
				constraintmap[key.String()] = append(constraintmap[key.String()], constraint)
			}
			return nil
		})
	}

	indexmap := map[string]map[string][]string{}
	errgrp.Go(func() error {
		resp, err := m.getIndicesBySchemasAndTables(ctx, schemaset)
		if err != nil {
			return err
		}
		indexmap = resp
		return nil
	})

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	output := []*sqlmanager_shared.TableInitStatement{}
	for _, schematable := range tables {
		key := schematable.String()
		tableData, ok := colDefMap[key]
		if !ok {
			continue
		}
		columns := make([]string, 0, len(tableData))
		for _, record := range tableData {
			record := record
			var identityType *string
			if record.IdentityGeneration.Valid {
				identityType = &record.IdentityGeneration.String
			}

			columnDefaultStr, err := convertUInt8ToString(record.ColumnDefault)
			if err != nil {
				return nil, fmt.Errorf("failed to convert column default to string: %w", err)
			}
			var columnDefaultType *string
			if identityType != nil && columnDefaultStr != "" && *identityType == "" {
				val := columnDefaultString // With this type columnDefaultStr will be surrounded by quotes when translated to SQL
				columnDefaultType = &val
			} else if identityType != nil && columnDefaultStr != "" && *identityType != "" {
				val := columnDefaultDefault // With this type columnDefaultStr will be surrounded by parentheses when translated to SQL
				columnDefaultType = &val
			}
			columnDefaultStr, err = EscapeMysqlDefaultColumn(columnDefaultStr, columnDefaultType)
			if err != nil {
				return nil, fmt.Errorf("failed to escape column default: %w", err)
			}

			if (!record.ColumnDefaultRaw.Valid || (record.ColumnDefaultRaw.Valid && record.ColumnDefaultRaw.String == "" || record.ColumnDefaultRaw.String == "NULL")) && record.IsNullable == 1 {
				columnDefaultStr = ""
				columnDefaultType = nil
			}
			fmt.Println("FOO", columnDefaultStr, columnDefaultType, identityType)
			if identityType != nil && *identityType == "" {
				fmt.Println("ID TYPE", *identityType)
			}

			genExp, err := convertUInt8ToString(record.GenerationExp)
			if err != nil {
				return nil, fmt.Errorf("failed to convert generation expression to string: %w", err)
			}
			columns = append(columns, buildTableCol(&buildTableColRequest{
				ColumnName:          record.ColumnName,
				ColumnDefault:       columnDefaultStr,
				DataType:            record.DataType,
				IsNullable:          record.IsNullable == 1,
				IdentityType:        identityType,
				GeneratedExpression: genExp,
			}))
		}

		info := &sqlmanager_shared.TableInitStatement{
			CreateTableStatement: fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s`.`%s` (%s);", tableData[0].SchemaName, tableData[0].TableName, strings.Join(columns, ", ")),
			AlterTableStatements: []*sqlmanager_shared.AlterTableStatement{},
			IndexStatements:      []string{},
		}
		for _, constraint := range constraintmap[key] {
			stmt, err := buildAlterStatementByConstraint(constraint)
			if err != nil {
				return nil, fmt.Errorf("failed to build alter table statement by constraint: %w", err)
			}
			info.AlterTableStatements = append(info.AlterTableStatements, stmt)
		}
		if tableIndices, ok := indexmap[key]; ok {
			for idxName, cols := range tableIndices {
				info.IndexStatements = append(
					info.IndexStatements,
					wrapIdempotentIndex(schematable.Schema, schematable.Table, idxName, cols),
				)
			}
		}
		output = append(output, info)
	}
	return output, nil
}

func (m *MysqlManager) getIndicesBySchemasAndTables(ctx context.Context, schemaset map[string][]string) (map[string]map[string][]string, error) {
	version, err := m.querier.GetVersion(ctx, m.pool)
	if err != nil {
		return nil, fmt.Errorf("failed to get mysql version: %w", err)
	}
	if isMariaDb(version) {
		return m.getMariaDbIndiciesBySchemasAndTables(ctx, schemaset)
	}
	return m.getMysqlIndiciesBySchemasAndTables(ctx, schemaset)
}

func (m *MysqlManager) getMariaDbIndiciesBySchemasAndTables(ctx context.Context, schemaset map[string][]string) (map[string]map[string][]string, error) {
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(5)

	var resultMu sync.Mutex
	results := make([][]*mysql_queries.GetMariaDbIndicesBySchemasAndTablesRow, 0, len(schemaset))
	for schema, tables := range schemaset {
		errgrp.Go(func() error {
			idxrecords, err := m.querier.GetMariaDbIndicesBySchemasAndTables(errctx, m.pool, &mysql_queries.GetMariaDbIndicesBySchemasAndTablesParams{
				Schema: schema,
				Tables: tables,
			})
			if err != nil {
				return fmt.Errorf("failed to build mysql indices by schemas and tables: %w", err)
			}
			resultMu.Lock()
			defer resultMu.Unlock()
			results = append(results, idxrecords)
			return nil
		})
	}
	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	indexmap := map[string]map[string][]string{}
	for _, records := range results {
		for _, record := range records {
			key := sqlmanager_shared.SchemaTable{Schema: record.SchemaName, Table: record.TableName}
			if _, exists := indexmap[key.String()]; !exists {
				indexmap[key.String()] = make(map[string][]string)
			}
			// Group composite columns by index name
			if record.ColumnName.Valid {
				indexmap[key.String()][record.IndexName] = append(
					indexmap[key.String()][record.IndexName],
					record.ColumnName.String,
				)
			}
		}
	}
	return indexmap, nil
}

func (m *MysqlManager) getMysqlIndiciesBySchemasAndTables(ctx context.Context, schemaset map[string][]string) (map[string]map[string][]string, error) {
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(5)

	var resultMu sync.Mutex
	results := make([][]*mysql_queries.GetMysqlIndicesBySchemasAndTablesRow, 0, len(schemaset))
	for schema, tables := range schemaset {
		errgrp.Go(func() error {
			idxrecords, err := m.querier.GetMysqlIndicesBySchemasAndTables(errctx, m.pool, &mysql_queries.GetMysqlIndicesBySchemasAndTablesParams{
				Schema: schema,
				Tables: tables,
			})
			if err != nil {
				return fmt.Errorf("failed to build mysql indices by schemas and tables: %w", err)
			}

			resultMu.Lock()
			defer resultMu.Unlock()
			results = append(results, idxrecords)
			return nil
		})
	}

	indexmap := map[string]map[string][]string{}
	for _, records := range results {
		for _, record := range records {
			key := sqlmanager_shared.SchemaTable{Schema: record.SchemaName, Table: record.TableName}
			if _, exists := indexmap[key.String()]; !exists {
				indexmap[key.String()] = make(map[string][]string)
			}
			// Group columns/expressions by index name
			if record.ColumnName.Valid {
				indexmap[key.String()][record.IndexName] = append(
					indexmap[key.String()][record.IndexName],
					record.ColumnName.String,
				)
			} else if record.Expression.Valid {
				indexmap[key.String()][record.IndexName] = append(
					indexmap[key.String()][record.IndexName],
					// expressions must be wrapped in parentheses on creation, but don't come out of the DB in that format /shrug
					fmt.Sprintf("(%s)", record.Expression.String),
				)
			}
		}
	}
	return indexmap, nil
}

func isMariaDb(version string) bool {
	return strings.Contains(version, "MariaDB")
}

func (m *MysqlManager) GetSequencesByTables(ctx context.Context, schema string, tables []string) ([]*sqlmanager_shared.DataType, error) {
	return nil, errors.ErrUnsupported
}

func convertUInt8ToString(value any) (string, error) {
	convertedType, ok := value.([]uint8)
	if !ok {
		return "", fmt.Errorf("failed to convert []uint8 to string")
	}
	return string(convertedType), nil
}

type buildTableColRequest struct {
	ColumnName          string
	ColumnDefault       string
	DataType            string
	IsNullable          bool
	GeneratedType       string
	GeneratedExpression string
	IdentityType        *string
}

func buildTableCol(record *buildTableColRequest) string {
	pieces := []string{EscapeMysqlColumn(record.ColumnName), record.DataType}

	if record.GeneratedExpression != "" {
		genType := ""
		if record.IdentityType != nil && *record.IdentityType == "STORED GENERATED" {
			genType = "STORED"
		} else if record.IdentityType != nil && *record.IdentityType == "VIRTUAL GENERATED" {
			genType = "VIRTUAL"
		}
		pieces = append(pieces, fmt.Sprintf("GENERATED ALWAYS AS (%s) %s", record.GeneratedExpression, genType))
	} else {
		pieces = append(pieces, buildNullableText(record.IsNullable))
	}

	if record.ColumnDefault != "" {
		pieces = append(pieces, fmt.Sprintf("DEFAULT %s", record.ColumnDefault))
	}

	if record.IdentityType != nil && *record.IdentityType == "auto_increment" {
		pieces = append(pieces, fmt.Sprintf("%s PRIMARY KEY", *record.IdentityType))
	}

	return strings.Join(pieces, " ")
}

func buildNullableText(isNullable bool) string {
	if isNullable {
		return "NULL"
	}
	return "NOT NULL"
}

func buildAlterStatementByConstraint(c *mysql_queries.GetTableConstraintsRow) (*sqlmanager_shared.AlterTableStatement, error) {
	constraintCols, err := jsonRawToSlice[string](c.ConstraintColumns)
	if err != nil {
		return nil, err
	}
	referencedCols, err := jsonRawToSlice[string](c.ReferencedColumnNames)
	if err != nil {
		return nil, err
	}
	switch c.ConstraintType {
	case "PRIMARY KEY":
		stmt := fmt.Sprintf("ALTER TABLE `%s`.`%s` ADD PRIMARY KEY (%s);", c.SchemaName, c.TableName, strings.Join(EscapeMysqlColumns(constraintCols), ","))
		return &sqlmanager_shared.AlterTableStatement{
			Statement:      wrapIdempotentConstraint(c.SchemaName, c.TableName, c.ConstraintName, stmt),
			ConstraintType: sqlmanager_shared.PrimaryConstraintType,
		}, nil
	case "UNIQUE":
		stmt := fmt.Sprintf("ALTER TABLE `%s`.`%s` ADD CONSTRAINT `%s` UNIQUE (%s);", c.SchemaName, c.TableName, c.ConstraintName, strings.Join(EscapeMysqlColumns(constraintCols), ","))
		return &sqlmanager_shared.AlterTableStatement{
			Statement:      wrapIdempotentConstraint(c.SchemaName, c.TableName, c.ConstraintName, stmt),
			ConstraintType: sqlmanager_shared.UniqueConstraintType,
		}, nil
	case "FOREIGN KEY":
		stmt := fmt.Sprintf("ALTER TABLE `%s`.`%s` ADD CONSTRAINT `%s` FOREIGN KEY (%s) REFERENCES `%s`.`%s`(%s) ON DELETE %s ON UPDATE %s;",
			c.SchemaName,
			c.TableName,
			c.ConstraintName,
			strings.Join(EscapeMysqlColumns(constraintCols), ","),
			c.ReferencedSchemaName,
			c.ReferencedTableName,
			strings.Join(EscapeMysqlColumns(referencedCols), ","),
			c.DeleteRule.String,
			c.UpdateRule.String,
		)
		return &sqlmanager_shared.AlterTableStatement{
			Statement:      wrapIdempotentConstraint(c.SchemaName, c.TableName, c.ConstraintName, stmt),
			ConstraintType: sqlmanager_shared.ForeignConstraintType,
		}, nil
	case "CHECK":
		checkStr, err := convertUInt8ToString(c.CheckClause)
		if err != nil {
			return nil, err
		}
		stmt := fmt.Sprintf("ALTER TABLE `%s`.`%s` ADD CONSTRAINT %s CHECK (%s);", c.SchemaName, c.TableName, c.ConstraintName, checkStr)
		return &sqlmanager_shared.AlterTableStatement{
			Statement:      wrapIdempotentConstraint(c.SchemaName, c.TableName, c.ConstraintName, stmt),
			ConstraintType: sqlmanager_shared.CheckConstraintType,
		}, nil
	}
	return nil, errors.ErrUnsupported
}

func (m *MysqlManager) GetSchemaTableDataTypes(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) (*sqlmanager_shared.SchemaTableDataTypeResponse, error) {
	if len(tables) == 0 {
		return &sqlmanager_shared.SchemaTableDataTypeResponse{}, nil
	}

	schemasMap := map[string]struct{}{}
	for _, t := range tables {
		schemasMap[t.Schema] = struct{}{}
	}
	schemas := []string{}
	for s := range schemasMap {
		schemas = append(schemas, s)
	}

	output := &sqlmanager_shared.SchemaTableDataTypeResponse{}
	funcs, err := m.getFunctionsByTables(ctx, schemas)
	if err != nil {
		return nil, fmt.Errorf("unable to get postgres custom functions by tables: %w", err)
	}
	output.Functions = append(output.Functions, funcs...)

	return output, nil
}

func (m *MysqlManager) GetSchemaTableTriggers(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableTrigger, error) {
	if len(tables) == 0 {
		return []*sqlmanager_shared.TableTrigger{}, nil
	}

	fullTableNames := make(map[string]struct{}, len(tables))
	schemaTableMap := map[string][]string{}
	for _, t := range tables {
		schemaTableMap[t.Schema] = append(schemaTableMap[t.Schema], t.Table)
		fullTableNames[t.String()] = struct{}{}
	}

	resMap := map[string][]*mysql_queries.GetCustomTriggersBySchemaAndTablesRow{}
	var resMapMu sync.Mutex

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(3)
	for schema, tables := range schemaTableMap {
		schema := schema
		tables := tables
		errgrp.Go(func() error {
			rows, err := m.querier.GetCustomTriggersBySchemaAndTables(errctx, m.pool, &mysql_queries.GetCustomTriggersBySchemaAndTablesParams{
				Schema: schema,
				Tables: tables,
			})
			if err != nil && !neosyncdb.IsNoRows(err) {
				return err
			} else if err != nil && neosyncdb.IsNoRows(err) {
				return nil
			}

			resMapMu.Lock()
			defer resMapMu.Unlock()
			resMap[schema] = append(resMap[schema], rows...)
			return nil
		})
	}
	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	output := []*sqlmanager_shared.TableTrigger{}
	for _, rows := range resMap {
		for _, row := range rows {
			if _, ok := fullTableNames[sqlmanager_shared.BuildTable(row.SchemaName, row.TableName)]; !ok {
				continue
			}
			output = append(output, &sqlmanager_shared.TableTrigger{
				Schema:      row.SchemaName,
				Table:       row.TableName,
				TriggerName: row.TriggerName,
				Definition:  wrapIdempotentTrigger(row.SchemaName, row.TableName, row.TriggerName, row.TriggerSchema, row.Timing, row.EventType, row.Orientation, row.Statement),
			})
		}
	}
	return output, nil
}

func (m *MysqlManager) GetSchemaInitStatements(
	ctx context.Context,
	tables []*sqlmanager_shared.SchemaTable,
) ([]*sqlmanager_shared.InitSchemaStatements, error) {
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(5)

	schemaStmts := []string{}
	errgrp.Go(func() error {
		uniqueSchemas := map[string]struct{}{}
		for _, table := range tables {
			uniqueSchemas[table.Schema] = struct{}{}
		}
		for schema := range uniqueSchemas {
			schemaStmts = append(schemaStmts, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS `%s`;", schema))
		}
		return nil
	})

	tableTriggerStmts := []string{}
	errgrp.Go(func() error {
		tableTriggers, err := m.GetSchemaTableTriggers(errctx, tables)
		if err != nil {
			return fmt.Errorf("unable to retrieve mysql schema table triggers: %w", err)
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
		initStatementCfgs, err := m.GetTableInitStatements(errctx, tables)
		if err != nil {
			return fmt.Errorf("unable to retrieve mysql schema table create statements: %w", err)
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
		{Label: SchemasLabel, Statements: schemaStmts},
		{Label: "data types"},
		{Label: "create table", Statements: createTables},
		{Label: "non-fk alter table", Statements: nonFkAlterStmts},
		{Label: "table index", Statements: idxStmts},
		{Label: "fk alter table", Statements: fkAlterStmts},
		{Label: "table triggers", Statements: tableTriggerStmts},
	}, nil
}

func (m *MysqlManager) getFunctionsByTables(ctx context.Context, schemas []string) ([]*sqlmanager_shared.DataType, error) {
	rows, err := m.querier.GetCustomFunctionsBySchemas(ctx, m.pool, schemas)
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
			Definition: wrapIdempotentFunction(row.FunctionName, row.ReturnDataType, row.Definition, row.IsDeterministic == 1),
		})
	}
	return output, nil
}

func wrapIdempotentConstraint(
	schema,
	table,
	constraintname,
	constraintStmt string,
) string {
	procedureName := fmt.Sprintf("NeosyncAddConstraint_%s", hashInput(schema, table, constraintname))[:64]
	stmt := fmt.Sprintf(`
CREATE PROCEDURE %[1]s()
BEGIN
    DECLARE constraint_exists INT DEFAULT 0;

    SELECT COUNT(*) INTO constraint_exists
    FROM information_schema.TABLE_CONSTRAINTS
    WHERE CONSTRAINT_SCHEMA = '%[2]s'
    AND TABLE_NAME = '%[3]s'
    AND CONSTRAINT_NAME = '%[4]s';

    IF constraint_exists = 0 THEN
        %[5]s
    END IF;
END;

CALL %[1]s();
DROP PROCEDURE %[1]s;
`, procedureName, schema, table, constraintname, constraintStmt)
	return strings.TrimSpace(stmt)
}

func hashInput(input ...string) string {
	hasher := sha256.New()
	for _, in := range input {
		hasher.Write([]byte(in))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func wrapIdempotentIndex(
	schema,
	table,
	constraintname string,
	cols []string,
) string {
	hashParams := []string{schema, table, constraintname}
	hashParams = append(hashParams, cols...)

	columnInput := []string{}
	for _, col := range cols {
		if strings.HasPrefix(col, "(") {
			columnInput = append(columnInput, col)
		} else {
			columnInput = append(columnInput, EscapeMysqlColumn(col))
		}
	}
	procedureName := fmt.Sprintf("NeosyncAddIndex_%s", hashInput(hashParams...))[:64]
	stmt := fmt.Sprintf(`
CREATE PROCEDURE %[1]s()
BEGIN
    DECLARE index_exists INT DEFAULT 0;

    SELECT COUNT(*) INTO index_exists
    FROM information_schema.statistics
    WHERE table_schema = '%[2]s'
    AND table_name = '%[3]s'
    AND index_name = '%[4]s';

    IF index_exists = 0 THEN
        CREATE INDEX %[5]s ON %[6]s.%[7]s(%[8]s);
    END IF;
END;

CALL %[1]s();
DROP PROCEDURE %[1]s;
`, procedureName, schema, table, constraintname, EscapeMysqlColumn(constraintname), EscapeMysqlColumn(schema), EscapeMysqlColumn(table), strings.Join(columnInput, ", "))
	return strings.TrimSpace(stmt)
}

func wrapIdempotentFunction(
	funcName,
	returnDataType,
	definition string,
	isDeterministic bool,
) string {
	deterministic := "DETERMINISTIC"
	if !isDeterministic {
		deterministic = "NOT DETERMINISTIC"
	}
	stmt := fmt.Sprintf(`
CREATE FUNCTION IF NOT EXISTS %s(%s)
RETURNS %s
%s
%s;
`, funcName, returnDataType, returnDataType, deterministic, definition)
	return strings.TrimSpace(stmt)
}

func wrapIdempotentTrigger(
	schema,
	tableName,
	triggerName,
	triggerSchema,
	timing,
	event_type,
	orientation,
	actionStmt string,
) string {
	stmt := fmt.Sprintf(`
CREATE TRIGGER IF NOT EXISTS %s.%s
%s %s ON %s.%s
FOR EACH %s
%s;
`, triggerSchema, triggerName, timing, event_type, EscapeMysqlColumn(schema), EscapeMysqlColumn(tableName), orientation, actionStmt)
	return strings.TrimSpace(stmt)
}

func (m *MysqlManager) BatchExec(ctx context.Context, batchSize int, statements []string, opts *sqlmanager_shared.BatchExecOpts) error {
	for i := 0; i < len(statements); i += batchSize {
		end := i + batchSize
		if end > len(statements) {
			end = len(statements)
		}

		batchCmd := strings.Join(statements[i:end], " ")
		if opts != nil && opts.Prefix != nil && *opts.Prefix != "" {
			batchCmd = fmt.Sprintf("%s %s", *opts.Prefix, batchCmd)
		}
		_, err := m.pool.ExecContext(ctx, batchCmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MysqlManager) GetTableRowCount(
	ctx context.Context,
	schema, table string,
	whereClause *string,
) (int64, error) {
	tableName := sqlmanager_shared.BuildTable(schema, table)
	builder := goqu.Dialect(sqlmanager_shared.MysqlDriver)
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
	err = m.pool.QueryRowContext(ctx, sql).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, err
}

func (m *MysqlManager) Exec(ctx context.Context, statement string) error {
	_, err := m.pool.ExecContext(ctx, statement)
	if err != nil {
		return err
	}
	return nil
}

func (m *MysqlManager) Close() {
	if m.pool != nil && m.close != nil {
		m.close()
	}
}

func BuildMysqlTruncateStatement(
	schema string,
	table string,
) (string, error) {
	builder := goqu.Dialect("mysql")
	sqltable := goqu.S(schema).Table(table)
	truncateStmt := builder.From(sqltable).Truncate()
	stmt, _, err := truncateStmt.ToSQL()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s;", stmt), nil
}

func EscapeMysqlColumns(cols []string) []string {
	outcols := make([]string, len(cols))
	for idx := range cols {
		outcols[idx] = EscapeMysqlColumn(cols[idx])
	}
	return outcols
}

func EscapeMysqlColumn(col string) string {
	return fmt.Sprintf("`%s`", col)
}

func EscapeMysqlDefaultColumn(defaultColumnValue string, defaultColumnType *string) (string, error) {
	defaultColumnTypes := []string{columnDefaultString, columnDefaultDefault}
	if defaultColumnType == nil {
		return defaultColumnValue, nil
	}
	if *defaultColumnType == columnDefaultString {
		return fmt.Sprintf("'%s'", defaultColumnValue), nil
	}
	if *defaultColumnType == columnDefaultDefault {
		return fmt.Sprintf("(%s)", defaultColumnValue), nil
	}
	return fmt.Sprintf("(%s)", defaultColumnValue), fmt.Errorf("unsupported default column type: %s, currently supported types are: %v", *defaultColumnType, defaultColumnTypes)
}

func GetMysqlColumnOverrideAndResetProperties(columnInfo *sqlmanager_shared.DatabaseSchemaRow) (needsOverride, needsReset bool) {
	needsOverride = false
	needsReset = false
	return
}
