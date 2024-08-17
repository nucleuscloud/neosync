package querybuilder2

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
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
	Args      []any
}

type QueryBuilder struct {
	tables          map[string]*TableInfo
	whereConditions map[string][]WhereCondition
	defaultSchema   string
	visitedTables   map[string]bool
}

func NewQueryBuilder(defaultSchema string) *QueryBuilder {
	return &QueryBuilder{
		tables:          make(map[string]*TableInfo),
		whereConditions: make(map[string][]WhereCondition),
		defaultSchema:   defaultSchema,
		visitedTables:   make(map[string]bool),
	}
}

func (qb *QueryBuilder) AddTable(table *TableInfo) {
	key := qb.getTableKey(table.Schema, table.Name)
	qb.tables[key] = table
}

func (qb *QueryBuilder) AddWhereCondition(schema, tableName, condition string, args ...any) {
	key := qb.getTableKey(schema, tableName)
	qb.whereConditions[key] = append(qb.whereConditions[key], WhereCondition{Condition: condition, Args: args})
}

func (qb *QueryBuilder) BuildQuery(schema, tableName string) (string, []any, error) {
	qb.visitedTables = make(map[string]bool) // Reset visited tables
	return qb.buildQueryRecursive(schema, tableName, false)
}

func (qb *QueryBuilder) buildQueryRecursive(schema, tableName string, isSubquery bool) (string, []any, error) {
	key := qb.getTableKey(schema, tableName)
	if qb.visitedTables[key] {
		return "", nil, nil // Avoid circular dependencies
	}
	qb.visitedTables[key] = true

	table, ok := qb.tables[key]
	if !ok {
		return "", nil, fmt.Errorf("table not found: %s", key)
	}

	query := squirrel.Select(qb.getQualifiedColumns(table)...).From(qb.getQualifiedTableName(table))

	// Add WHERE conditions for this table
	if conditions, ok := qb.whereConditions[key]; ok {
		for _, cond := range conditions {
			qualifiedCondition := qb.qualifyWhereCondition(table, cond.Condition)
			query = query.Where(qualifiedCondition, cond.Args...)
		}
	}

	var allArgs []any

	// Recursively build and join queries for related tables
	for _, fk := range table.ForeignKeys {
		subQuery, subArgs, err := qb.buildQueryRecursive(fk.ReferenceSchema, fk.ReferenceTable, true)
		if err != nil {
			return "", nil, err
		}
		if subQuery != "" {
			joinCondition := qb.buildJoinCondition(table, fk)
			query = query.JoinClause(fmt.Sprintf("INNER JOIN (%s) AS %s ON %s", subQuery, qb.getQualifiedTableName(&TableInfo{Schema: fk.ReferenceSchema, Name: fk.ReferenceTable}), joinCondition))
			allArgs = append(allArgs, subArgs...)
		}
	}

	// If this is a subquery, wrap it and return
	if isSubquery {
		sql, args, err := query.ToSql()
		if err != nil {
			return "", nil, err
		}
		return sql, append(allArgs, args...), nil
	}

	// This is the main query, convert to SQL and return
	sql, args, err := query.ToSql()
	if err != nil {
		return "", nil, err
	}

	return sql, append(allArgs, args...), nil
}

func (qb *QueryBuilder) qualifyWhereCondition(table *TableInfo, condition string) string {
	// Split the condition into parts
	parts := strings.Fields(condition)

	// Qualify the column name (assumed to be the first part)
	if len(parts) > 0 {
		parts[0] = fmt.Sprintf("%s.%s", qb.getQualifiedTableName(table), parts[0])
	}

	// Rejoin the condition
	return strings.Join(parts, " ")
}

func (qb *QueryBuilder) getQualifiedColumns(table *TableInfo) []string {
	qualifiedColumns := make([]string, len(table.Columns))
	for i, col := range table.Columns {
		qualifiedColumns[i] = fmt.Sprintf("%s.%s", qb.getQualifiedTableName(table), col)
	}
	return qualifiedColumns
}

func (qb *QueryBuilder) getQualifiedTableName(table *TableInfo) string {
	if table.Schema == "" || table.Schema == qb.defaultSchema {
		return table.Name
	}
	return fmt.Sprintf("%s.%s", table.Schema, table.Name)
}

func (qb *QueryBuilder) buildJoinCondition(table *TableInfo, fk ForeignKey) string {
	conditions := make([]string, len(fk.Columns))
	for i := range fk.Columns {
		conditions[i] = fmt.Sprintf("%s.%s = %s.%s",
			qb.getQualifiedTableName(table),
			fk.Columns[i],
			qb.getQualifiedTableName(&TableInfo{Schema: fk.ReferenceSchema, Name: fk.ReferenceTable}),
			fk.ReferenceColumns[i])
	}
	return strings.Join(conditions, " AND ")
}

func (qb *QueryBuilder) getTableKey(schema, tableName string) string {
	if schema == "" {
		schema = qb.defaultSchema
	}
	return fmt.Sprintf("%s.%s", schema, tableName)
}
