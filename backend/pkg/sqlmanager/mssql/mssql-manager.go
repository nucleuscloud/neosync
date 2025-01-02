package sqlmanager_mssql

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/doug-martin/goqu/v9"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	ee_sqlmanager_mssql "github.com/nucleuscloud/neosync/internal/ee/mssql-manager"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
)

type Manager struct {
	querier mssql_queries.Querier
	db      mysql_queries.DBTX
	close   func()

	ee_sqlmanager_mssql.Manager
}

func NewManager(querier mssql_queries.Querier, db mysql_queries.DBTX, closer func()) *Manager {
	return &Manager{querier: querier, db: db, close: closer, Manager: *ee_sqlmanager_mssql.NewManager(querier, db, closer)}
}

const defaultIdentity string = "IDENTITY(1,1)"

func (m *Manager) GetDatabaseSchema(ctx context.Context) ([]*sqlmanager_shared.DatabaseSchemaRow, error) {
	dbSchemas, err := m.querier.GetDatabaseSchema(ctx, m.db)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, err
	} else if err != nil && neosyncdb.IsNoRows(err) {
		return []*sqlmanager_shared.DatabaseSchemaRow{}, nil
	}

	output := []*sqlmanager_shared.DatabaseSchemaRow{}
	for _, row := range dbSchemas {
		charMaxLength := -1
		if row.CharacterMaximumLength.Valid {
			charMaxLength = int(row.CharacterMaximumLength.Int32)
		}
		numericPrecision := -1
		if row.NumericPrecision.Valid {
			numericPrecision = int(row.NumericPrecision.Int16)
		}
		numericScale := -1
		if row.NumericScale.Valid {
			numericScale = int(row.NumericScale.Int16)
		}

		var identityGeneration *string
		if row.IsIdentity {
			syntax := defaultIdentity
			identityGeneration = &syntax
		}
		var generatedType *string
		if row.GenerationExpression.Valid {
			generatedType = &row.GenerationExpression.String
		}

		var identitySeed *int
		if row.IdentitySeed.Valid {
			seed := int(row.IdentitySeed.Int32)
			identitySeed = &seed
		}

		var identityIncrement *int
		if row.IdentityIncrement.Valid {
			increment := int(row.IdentityIncrement.Int32)
			identityIncrement = &increment
		}

		output = append(output, &sqlmanager_shared.DatabaseSchemaRow{
			TableSchema:            row.TableSchema,
			TableName:              row.TableName,
			ColumnName:             row.ColumnName,
			DataType:               row.DataType,
			ColumnDefault:          row.ColumnDefault, // todo: make sure this is valid for the other funcs
			IsNullable:             row.IsNullable != "NO",
			GeneratedType:          generatedType,
			OrdinalPosition:        int(row.OrdinalPosition),
			CharacterMaximumLength: charMaxLength,
			NumericPrecision:       numericPrecision,
			NumericScale:           numericScale,
			IdentityGeneration:     identityGeneration,
			IdentitySeed:           identitySeed,
			IdentityIncrement:      identityIncrement,
		})
	}

	return output, nil
}

func (m *Manager) GetSchemaColumnMap(ctx context.Context) (map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow, error) {
	dbSchemas, err := m.GetDatabaseSchema(ctx)
	if err != nil {
		return nil, err
	}
	result := sqlmanager_shared.GetUniqueSchemaColMappings(dbSchemas)
	return result, nil
}

func (m *Manager) GetTableConstraintsBySchema(ctx context.Context, schemas []string) (*sqlmanager_shared.TableConstraints, error) {
	if len(schemas) == 0 {
		return &sqlmanager_shared.TableConstraints{}, nil
	}
	rows, err := m.querier.GetTableConstraintsBySchemas(ctx, m.db, schemas)
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
		constraintCols := splitAndStrip(row.ConstraintColumns, ", ")

		switch row.ConstraintType {
		case "FOREIGN KEY":
			if row.ReferencedColumns.Valid && row.ReferencedTable.Valid {
				fkCols := splitAndStrip(row.ReferencedColumns.String, ", ")

				ccNullability := splitAndStrip(row.ConstraintColumnsNullability, ", ")
				notNullable := []bool{}
				for _, nullability := range ccNullability {
					notNullable = append(notNullable, nullability == "NOT NULL")
				}
				if len(constraintCols) != len(fkCols) {
					return nil, fmt.Errorf("length of columns was not equal to length of foreign key cols: %d %d", len(constraintCols), len(fkCols))
				}
				if len(constraintCols) != len(notNullable) {
					return nil, fmt.Errorf("length of columns was not equal to length of not nullable cols: %d %d", len(constraintCols), len(notNullable))
				}

				if isInvalidCircularSelfReferencingFk(row, constraintCols, fkCols) {
					continue
				}

				foreignKeyMap[tableName] = append(foreignKeyMap[tableName], &sqlmanager_shared.ForeignConstraint{
					Columns:     constraintCols,
					NotNullable: notNullable,
					ForeignKey: &sqlmanager_shared.ForeignKey{
						Table:   sqlmanager_shared.BuildTable(row.ReferencedSchema.String, row.ReferencedTable.String),
						Columns: fkCols,
					},
				})
			}

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
	}, nil
}

// Checks if a foreign key constraint is self-referencing (points to the same table)
// and all constraint columns match their referenced columns, indicating a circular reference.
// example  public.users.id has a foreign key to public.users.id
func isInvalidCircularSelfReferencingFk(row *mssql_queries.GetTableConstraintsBySchemasRow, constraintColumns, referencedColumns []string) bool {
	// Check if the foreign key references the same table
	isSameTable := row.SchemaName == row.ReferencedSchema.String &&
		row.TableName == row.ReferencedTable.String
	if !isSameTable {
		return false
	}

	// Check if all constraint columns exist in referenced columns
	for _, column := range constraintColumns {
		if !slices.Contains(referencedColumns, column) {
			return false
		}
	}

	return true
}

func (m *Manager) GetRolePermissionsMap(ctx context.Context) (map[string][]string, error) {
	rows, err := m.querier.GetRolePermissions(ctx, m.db)
	if err != nil && !neosyncdb.IsNoRows(err) {
		return nil, fmt.Errorf("unable to retrieve mssql role permissions: %w", err)
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

func splitAndStrip(input, delim string) []string {
	output := []string{}

	for _, piece := range strings.Split(input, delim) {
		if strings.TrimSpace(piece) != "" {
			output = append(output, piece)
		}
	}

	return output
}

func (m *Manager) BatchExec(ctx context.Context, batchSize int, statements []string, opts *sqlmanager_shared.BatchExecOpts) error {
	// mssql does not support batching statements
	total := len(statements)
	for idx, stmt := range statements {
		err := m.Exec(ctx, stmt)
		if err != nil {
			return fmt.Errorf("failed to execute batch statement %d/%d: %w", idx+1, total, err)
		}
	}
	return nil
}

func (m *Manager) GetTableRowCount(
	ctx context.Context,
	schema, table string,
	whereClause *string,
) (int64, error) {
	tableName := sqlmanager_shared.BuildTable(schema, table)
	builder := goqu.Dialect(sqlmanager_shared.MssqlDriver)

	query := builder.From(goqu.I(tableName)).Select(goqu.COUNT("*"))
	if whereClause != nil && *whereClause != "" {
		query = query.Where(goqu.L(*whereClause))
	}
	sql, _, err := query.ToSQL()
	if err != nil {
		return 0, fmt.Errorf("unable to build table row count statement for mssql: %w", err)
	}
	var count int64
	err = m.db.QueryRowContext(ctx, sql).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unable to query table row count for mssql: %w", err)
	}
	return count, err
}

func (m *Manager) Exec(ctx context.Context, statement string) error {
	_, err := m.db.ExecContext(ctx, statement)
	return err
}

func (m *Manager) Close() {
	if m.db != nil && m.close != nil {
		m.close()
	}
}

func GetMssqlColumnOverrideAndResetProperties(columnInfo *sqlmanager_shared.DatabaseSchemaRow) (needsOverride, needsReset bool) {
	needsOverride = false
	needsReset = false

	// check if the column is an idenitity type
	if columnInfo.IdentityGeneration != nil && *columnInfo.IdentityGeneration != "" {
		needsOverride = true
		needsReset = true
		return
	}

	// check if column default is sequence
	if columnInfo.ColumnDefault != "" && gotypeutil.CaseInsensitiveContains(columnInfo.ColumnDefault, "NEXT VALUE") {
		needsReset = true
		return
	}

	return
}

func BuildMssqlDeleteStatement(
	schema, table string,
) (string, error) {
	dialect := goqu.Dialect("sqlserver")
	ds := dialect.Delete(goqu.S(schema).Table(table))
	sql, _, err := ds.ToSQL()
	if err != nil {
		return "", err
	}
	return sql + ";", nil
}

// Resets current identity value back to the initial count
func BuildMssqlIdentityColumnResetStatement(
	schema, table string, identitySeed, identityIncrement *int,
) string {
	if identitySeed != nil && identityIncrement != nil {
		return fmt.Sprintf("DBCC CHECKIDENT ('%s.%s', RESEED, %d);", schema, table, *identitySeed)
	}
	return BuildMssqlIdentityColumnResetCurrent(schema, table)
}

// If the current identity value for a table is less than the maximum identity value stored in the identity column
// It is reset using the maximum value in the identity column.
func BuildMssqlIdentityColumnResetCurrent(
	schema, table string,
) string {
	return fmt.Sprintf("DBCC CHECKIDENT ('%s.%s', RESEED)", schema, table)
}

// Allows explicit values to be inserted into the identity column of a table.
func BuildMssqlSetIdentityInsertStatement(
	schema, table string,
	enable bool,
) string {
	enabledKeyword := "OFF"
	if enable {
		enabledKeyword = "ON"
	}
	return fmt.Sprintf("SET IDENTITY_INSERT %q.%q %s;", schema, table, enabledKeyword)
}
