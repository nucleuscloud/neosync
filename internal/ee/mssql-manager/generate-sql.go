package ee_sqlmanager_mssql

import (
	"fmt"
	"strings"

	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
)

// Creates idempotent create table statement
func generateCreateTableStatement(rows []*mssql_queries.GetDatabaseTableSchemasBySchemasAndTablesRow) string {
	if len(rows) == 0 {
		return ""
	}

	tableSchema := rows[0].TableSchema
	tableName := rows[0].TableName

	var sb strings.Builder

	// Create table if not exists
	sb.WriteString(fmt.Sprintf("IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[%s].[%s]') AND type in (N'U'))\nBEGIN\n",
		tableSchema, tableName))
	sb.WriteString(fmt.Sprintf("CREATE TABLE [%s].[%s] (\n", tableSchema, tableName))

	primaryKeys := []string{}
	var periodDefinition *string
	var temporalDefinition *string
	// Process each column
	for i, row := range rows {
		if row.IsPrimary {
			primaryKeys = append(primaryKeys, row.ColumnName)
		}
		if row.PeriodDefinition.Valid {
			periodDefinition = &row.PeriodDefinition.String
		}
		if row.TemporalDefinition.Valid {
			temporalDefinition = &row.TemporalDefinition.String
		}

		sb.WriteString(fmt.Sprintf("    [%s] ", row.ColumnName))

		if !row.IsComputed || !row.GenerationExpression.Valid {
			switch {
			case row.CharacterMaximumLength.Valid:
				if row.CharacterMaximumLength.Int32 == -1 {
					sb.WriteString(fmt.Sprintf("%s(MAX)", row.DataType))
				} else {
					sb.WriteString(fmt.Sprintf("%s(%d)", row.DataType, row.CharacterMaximumLength.Int32))
				}
				// Types with precision only (float)
			case strings.EqualFold(row.DataType, "float") && row.NumericPrecision.Valid:
				sb.WriteString(fmt.Sprintf("%s(%d)", row.DataType, row.NumericPrecision.Int16))

			// Types with scale only (datetime2, datetimeoffset, time)
			case (strings.EqualFold(row.DataType, "datetime2") ||
				strings.EqualFold(row.DataType, "datetimeoffset") ||
				strings.EqualFold(row.DataType, "time")) &&
				row.NumericScale.Valid:
				sb.WriteString(fmt.Sprintf("%s(%d)", row.DataType, row.NumericScale.Int16))

			// Types with both precision and scale (decimal, numeric)
			case (strings.EqualFold(row.DataType, "decimal") ||
				strings.EqualFold(row.DataType, "numeric")) &&
				row.NumericPrecision.Valid && row.NumericScale.Valid:
				sb.WriteString(fmt.Sprintf("%s(%d,%d)", row.DataType,
					row.NumericPrecision.Int16, row.NumericScale.Int16))
			default:
				sb.WriteString(row.DataType)
			}
		}

		// Add identity specification
		if row.IsIdentity {
			seed := int32(1)
			increment := int32(1)
			if row.IdentitySeed.Valid {
				seed = row.IdentitySeed.Int32
			}
			if row.IdentityIncrement.Valid {
				increment = row.IdentityIncrement.Int32
			}
			sb.WriteString(fmt.Sprintf(" IDENTITY(%d,%d)", seed, increment))
		}

		// Add computed column specification
		if row.IsComputed && row.GenerationExpression.Valid {
			sb.WriteString(fmt.Sprintf(" AS %s", row.GenerationExpression.String))
			if row.IsPersisted {
				sb.WriteString(" PERSISTED")
			}
		}

		// Add generated always specification
		if row.GeneratedAlwaysType.Valid {
			sb.WriteString(fmt.Sprintf(" %s", row.GeneratedAlwaysType.String))
		}

		// Add nullability
		if !row.IsComputed {
			if row.IsNullable {
				sb.WriteString(" NULL")
			} else {
				sb.WriteString(" NOT NULL")
			}
		}

		// Add default constraint
		if row.ColumnDefault.Valid {
			sb.WriteString(fmt.Sprintf(" DEFAULT %s", row.ColumnDefault.String))
		}

		// Add comma if not last column
		if i < len(rows)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	if len(primaryKeys) > 0 {
		sb.WriteString(fmt.Sprintf("CONSTRAINT pk_%s PRIMARY KEY (%s)", tableName, strings.Join(primaryKeys, ",")))
	}

	if periodDefinition != nil && *periodDefinition != "" {
		sb.WriteString(fmt.Sprintf(", \n %s", *periodDefinition))
	}
	if temporalDefinition != nil && *temporalDefinition != "" {
		sb.WriteString(") \n WITH (SYSTEM_VERSIONING = ON)")
	} else {
		sb.WriteString(")")
	}

	// Close the CREATE TABLE statement
	sb.WriteString("\nEND")

	return sb.String()
}

// Creates idempotent create index statement
func generateCreateIndexStatement(record *mssql_queries.GetIndicesBySchemasAndTablesRow) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`
IF NOT EXISTS (
	SELECT * 
	FROM sys.indexes 
	WHERE name = N'%s' 
	AND object_id = OBJECT_ID(N'[%s].[%s]')
)
BEGIN
	%s
END`, record.IndexName, record.SchemaName, record.TableName, record.IndexDefinition))

	return sb.String()
}

// Creates idempotent create trigger statement
func generateCreateTriggerStatement(triggerName, schemaName, definition string) string {
	var sb strings.Builder
	def := strings.ReplaceAll(definition, "'", "''")
	sb.WriteString(fmt.Sprintf(`
IF NOT EXISTS (
	SELECT *
	FROM
		sys.triggers t
	JOIN sys.objects o ON o.object_id = t.object_id
	where t.name = N'%s' AND SCHEMA_NAME(o.schema_id) = N'%s'
)
BEGIN
	Exec('%s')
END`, triggerName, schemaName, def))

	return sb.String()
}

// Creates idempotent create sequence statement
func generateCreateSequenceStatement(record *mssql_queries.GetCustomSequencesBySchemasRow) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`
IF NOT EXISTS (
	SELECT * 
	FROM sys.sequences 
	WHERE name = N'%s' 
	AND SCHEMA_NAME(schema_id) = N'%s'
)
BEGIN
	%s
END`, record.SequenceName, record.SchemaName, record.Definition))

	return sb.String()
}

// Creates idempotent create database object statement
func generateCreateDatabaseObjectStatement(name, schema, definition string) string {
	var sb strings.Builder
	def := strings.ReplaceAll(definition, "'", "''")
	sb.WriteString(fmt.Sprintf(`
IF NOT EXISTS (
	SELECT * 
	FROM sys.objects 
  WHERE name = N'%s'
  AND schema_id = SCHEMA_ID(N'%s')
)
BEGIN
  Exec('%s')
END`, name, schema, def))

	return sb.String()
}

// Creates idempotent create type statement
func generateCreateDataTypeStatement(record *mssql_queries.GetDataTypesBySchemasRow) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`
IF NOT EXISTS (
    SELECT * 
    FROM sys.types t
    JOIN sys.schemas s ON t.schema_id = s.schema_id
    WHERE t.name = N'%s' 
    AND s.name = N'%s'
)
BEGIN
	%s
END`, record.TypeName, record.SchemaName, record.Definition))

	return sb.String()
}

// Creates idempotent alter table add constraint statement
func generateAddConstraintStatement(constraint *mssql_queries.GetTableConstraintsBySchemasRow) string {
	var sb strings.Builder

	// Start IF NOT EXISTS check
	sb.WriteString(fmt.Sprintf(`
IF NOT EXISTS (
	SELECT * FROM sys.objects o
	JOIN sys.objects oo ON oo.object_id = o.parent_object_id 
	WHERE SCHEMA_NAME(oo.schema_id) = N'%s' AND oo.name = N'%s' AND o.type = '%s'
)
BEGIN
  ALTER TABLE [%s].[%s]
  ADD CONSTRAINT [%s] `,
		constraint.SchemaName,
		constraint.TableName,
		getConstraintTypeCode(constraint.ConstraintType),
		constraint.SchemaName,
		constraint.TableName,
		constraint.ConstraintName))

	// Add constraint definition based on type
	switch constraint.ConstraintType {
	case "PRIMARY KEY":
		sb.WriteString("PRIMARY KEY ")
		sb.WriteString(fmt.Sprintf("(%s)", escapeColumnList(constraint.ConstraintColumns)))

	case "UNIQUE":
		sb.WriteString("UNIQUE ")
		sb.WriteString(fmt.Sprintf("(%s)", escapeColumnList(constraint.ConstraintColumns)))

	case "FOREIGN KEY":
		sb.WriteString(fmt.Sprintf("FOREIGN KEY (%s) ", escapeColumnList(constraint.ConstraintColumns)))
		if constraint.ReferencedSchema.Valid && constraint.ReferencedTable.Valid && constraint.ReferencedColumns.Valid {
			sb.WriteString(fmt.Sprintf("REFERENCES [%s].[%s] (%s)",
				constraint.ReferencedSchema.String,
				constraint.ReferencedTable.String,
				escapeColumnList(constraint.ReferencedColumns.String)))
		}
		if constraint.FKActions.Valid {
			sb.WriteString(" " + constraint.FKActions.String)
		}

	case "CHECK":
		if constraint.CheckClause.Valid {
			sb.WriteString("CHECK ")
			sb.WriteString(constraint.CheckClause.String)
		}
	}

	sb.WriteString("\nEND")

	return sb.String()
}

// Escapes column list
func escapeColumnList(columns string) string {
	parts := strings.Split(columns, ",")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		parts[i] = fmt.Sprintf("[%s]", part)
	}
	return strings.Join(parts, ", ")
}

// Gets SQL Server constraint type code
func getConstraintTypeCode(constraintType string) string {
	switch constraintType {
	case "PRIMARY KEY":
		return "PK"
	case "UNIQUE":
		return "UQ"
	case "FOREIGN KEY":
		return "F"
	case "CHECK":
		return "C"
	default:
		return ""
	}
}
