package querybuilder2

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
)

type TableInfo struct {
	Schema      string
	Name        string
	Columns     []string
	PrimaryKeys []string
	ForeignKeys []ForeignKey
}

type ForeignKey struct {
	Columns          []string
	ReferenceSchema  string
	ReferenceTable   string
	ReferenceColumns []string
}

type WhereCondition struct {
	Condition string
	Args      []interface{}
}

type QueryBuilder struct {
	tables                        map[string]*TableInfo
	whereConditions               map[string][]WhereCondition
	defaultSchema                 string
	visitedTables                 map[string]bool
	driver                        string
	subsetByForeignKeyConstraints bool
}

func NewQueryBuilder(defaultSchema, driver string, subsetByForeignKeyConstraints bool) *QueryBuilder {
	return &QueryBuilder{
		tables:                        make(map[string]*TableInfo),
		whereConditions:               make(map[string][]WhereCondition),
		defaultSchema:                 defaultSchema,
		visitedTables:                 make(map[string]bool),
		driver:                        driver,
		subsetByForeignKeyConstraints: subsetByForeignKeyConstraints,
	}
}

func (qb *QueryBuilder) AddTable(table *TableInfo) {
	key := qb.getTableKey(table.Schema, table.Name)
	qb.tables[key] = table
}

func (qb *QueryBuilder) getRequiredColumns(table *TableInfo) []string {
	columns := make([]string, 0)
	// Add primary keys
	columns = append(columns, table.PrimaryKeys...)
	// Add foreign key columns
	for _, fk := range table.ForeignKeys {
		columns = append(columns, fk.Columns...)
	}
	// Add columns used in WHERE conditions
	if conditions, ok := qb.whereConditions[qb.getTableKey(table.Schema, table.Name)]; ok {
		for _, cond := range conditions {
			parts := strings.Fields(cond.Condition)
			if len(parts) > 0 {
				columns = append(columns, parts[0])
			}
		}
	}
	// Remove duplicates
	return uniqueStrings(columns)
}

func (qb *QueryBuilder) AddWhereCondition(schema, tableName, condition string, args ...interface{}) {
	key := qb.getTableKey(schema, tableName)
	qb.whereConditions[key] = append(qb.whereConditions[key], WhereCondition{Condition: condition, Args: args})
}

func (qb *QueryBuilder) BuildQuery(schema, tableName string) (string, []interface{}, error) {
	qb.visitedTables = make(map[string]bool) // Reset visited tables
	return qb.buildQueryRecursive(schema, tableName, nil, nil, map[string]int{})
}

func (qb *QueryBuilder) buildQueryRecursive(schema, tableName string, parentTable *TableInfo, columnsToInclude []string, joinCount map[string]int) (string, []interface{}, error) {
	key := qb.getTableKey(schema, tableName)
	if qb.visitedTables[key] {
		return "", nil, nil // Avoid circular dependencies
	}
	qb.visitedTables[key] = true
	defer delete(qb.visitedTables, key) // Remove from visited after processing

	table, ok := qb.tables[key]
	if !ok {
		return "", nil, fmt.Errorf("table not found: %s", key)
	}

	if len(columnsToInclude) == 0 {
		columnsToInclude = qb.getRequiredColumns(table)
	}
	// If still no columns, select all columns
	if len(columnsToInclude) == 0 {
		columnsToInclude = table.Columns
	}
	// If still no columns, select '*'
	if len(columnsToInclude) == 0 {
		columnsToInclude = []string{"*"}
	}

	query := squirrel.Select(qb.getQualifiedColumns(table, columnsToInclude)...).From(qb.getQualifiedTableName(table))

	var allArgs []interface{}

	// Add WHERE conditions for this table
	if conditions, ok := qb.whereConditions[key]; ok {
		for _, cond := range conditions {
			qualifiedCondition := qb.qualifyWhereCondition(table, cond.Condition)
			query = query.Where(qualifiedCondition, cond.Args...)
			allArgs = append(allArgs, cond.Args...)
		}
	}

	// Only join and apply subsetting if subsetByForeignKeyConstraints is true
	if qb.subsetByForeignKeyConstraints {
		// Recursively build and join queries for related tables
		for _, fk := range table.ForeignKeys {
			subQuery, subArgs, err := qb.buildQueryRecursive(fk.ReferenceSchema, fk.ReferenceTable, table, fk.ReferenceColumns, joinCount)
			if err != nil {
				return "", nil, err
			}
			if subQuery != "" {
				joinCount[fk.ReferenceTable]++
				subQueryAlias := fmt.Sprintf("%s_%s_%d", fk.ReferenceSchema, fk.ReferenceTable, joinCount[fk.ReferenceTable])
				joinCondition := qb.buildJoinCondition(table, fk, joinCount[fk.ReferenceTable])
				query = query.JoinClause(fmt.Sprintf("INNER JOIN (%s) AS %s ON %s",
					subQuery,
					quoteIdentifier(qb.driver, subQueryAlias),
					joinCondition))
				allArgs = append(allArgs, subArgs...)
			}
		}

		// Apply subsetting based on parent table's where conditions
		if parentTable != nil {
			parentKey := qb.getTableKey(parentTable.Schema, parentTable.Name)
			if parentConditions, ok := qb.whereConditions[parentKey]; ok {
				for _, parentCond := range parentConditions {
					subsetCondition := qb.buildSubsetCondition(table, parentTable, parentCond.Condition)
					if subsetCondition != "" {
						query = query.Where(subsetCondition)
					}
				}
			}
		}
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return "", nil, err
	}

	return sql, append(allArgs, args...), nil
}

func (qb *QueryBuilder) qualifyWhereCondition(table *TableInfo, condition string) string {
	parts := strings.Fields(condition)
	if len(parts) > 0 {
		parts[0] = fmt.Sprintf("%s.%s", qb.getQualifiedTableName(table), quoteIdentifier(qb.driver, parts[0]))
	}
	return strings.Join(parts, " ")
}

func (qb *QueryBuilder) getQualifiedTableName(table *TableInfo) string {
	if table.Schema == "" || table.Schema == qb.defaultSchema {
		return quoteIdentifier(qb.driver, table.Name)
	}
	return fmt.Sprintf("%s.%s",
		quoteIdentifier(qb.driver, table.Schema),
		quoteIdentifier(qb.driver, table.Name))
}

func (qb *QueryBuilder) getQualifiedColumns(table *TableInfo, columnsToInclude []string) []string {
	qualifiedColumns := make([]string, len(columnsToInclude))
	for i, col := range columnsToInclude {
		qualifiedColumns[i] = fmt.Sprintf("%s.%s",
			qb.getQualifiedTableName(table),
			quoteIdentifier(qb.driver, col))
	}
	return qualifiedColumns
}

func (qb *QueryBuilder) buildJoinCondition(table *TableInfo, fk ForeignKey, joinCount int) string {
	conditions := make([]string, len(fk.Columns))
	for i := range fk.Columns {
		conditions[i] = fmt.Sprintf("%s.%s = %s.%s",
			qb.getQualifiedTableName(table),
			quoteIdentifier(qb.driver, fk.Columns[i]),
			quoteIdentifier(qb.driver, fmt.Sprintf("%s_%s_%d", fk.ReferenceSchema, fk.ReferenceTable, joinCount)),
			quoteIdentifier(qb.driver, fk.ReferenceColumns[i]))
	}
	return strings.Join(conditions, " AND ")
}

func (qb *QueryBuilder) buildSubsetCondition(childTable, parentTable *TableInfo, parentCondition string) string {
	for _, fk := range childTable.ForeignKeys {
		if fk.ReferenceSchema == parentTable.Schema && fk.ReferenceTable == parentTable.Name {
			parts := strings.Fields(parentCondition)
			if len(parts) > 2 && (parts[1] == "=" || parts[1] == "IN") {
				return fmt.Sprintf("%s.%s %s %s",
					qb.getQualifiedTableName(childTable),
					quoteIdentifier(qb.driver, fk.Columns[0]),
					parts[1],
					strings.Join(parts[2:], " "))
			}
		}
	}
	return ""
}

func (qb *QueryBuilder) getTableKey(schema, tableName string) string {
	if schema == "" {
		schema = qb.defaultSchema
	}
	return fmt.Sprintf("%s.%s", schema, tableName)
}

func quoteIdentifier(driver, identifier string) string {
	switch driver {
	case sqlmanager_shared.PostgresDriver:
		return fmt.Sprintf("\"%s\"", strings.Replace(identifier, "\"", "\"\"", -1))
	case sqlmanager_shared.MysqlDriver:
		return fmt.Sprintf("`%s`", strings.Replace(identifier, "`", "``", -1))
	default:
		return identifier
	}
}
