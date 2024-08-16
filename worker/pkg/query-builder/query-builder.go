package querybuilder

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	utils "github.com/nucleuscloud/neosync/backend/pkg/utils"
	pg_query "github.com/pganalyze/pg_query_go/v5"
	pgquery "github.com/wasilibs/go-pgquery"
	"github.com/xwb1989/sqlparser"

	// import the dialect
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
)

const (
	innerJoin joinType = "INNER"
)

type joinType string

type sqlJoin struct {
	JoinType  joinType
	JoinTable string
	BaseTable string
	// Alias          *string
	JoinColumnsMap map[string]string // map of joinColumn to baseColumn
}

type SubsetReferenceKey struct {
	Table         string
	Columns       []string
	OriginalTable *string
}
type SubsetColumnConstraint struct {
	Columns     []string
	NotNullable []bool
	ForeignKey  *SubsetReferenceKey
}

func BuildSelectQuery(
	driver, table string,
	columns []string,
	whereClause *string,
) (string, error) {
	builder := goqu.Dialect(driver)
	sqltable := goqu.I(table)

	selectColumns := make([]any, len(columns))
	for i, col := range columns {
		selectColumns[i] = col
	}
	query := builder.From(sqltable).Select(selectColumns...)

	if whereClause != nil && *whereClause != "" {
		query = query.Where(goqu.L(*whereClause))
	}
	sql, _, err := query.ToSQL()
	if err != nil {
		return "", err
	}

	return formatSqlQuery(sql), nil
}

func BuildSelectLimitQuery(
	driver, table string,
	limit uint,
) (string, error) {
	builder := goqu.Dialect(driver)
	sqltable := goqu.I(table)
	sql, _, err := builder.From((sqltable)).Limit(limit).ToSQL()
	if err != nil {
		return "", err
	}
	return sql, nil
}

func BuildSelectJoinQuery(
	driver, table string,
	columns []string,
	joins []*sqlJoin,
	whereClauses []string,
	aliasReference map[string]string,
) (string, error) {
	builder := goqu.Dialect(driver)

	mainTableAlias := aliasReference[table]
	if mainTableAlias == "" {
		mainTableAlias = table
	}
	// Split the table name into schema and table parts
	schemaAndTable := strings.SplitN(table, ".", 2)
	var sqltable exp.Expression
	if len(schemaAndTable) == 2 {
		sqltable = goqu.S(schemaAndTable[0]).Table(schemaAndTable[1])
	} else {
		sqltable = goqu.T(table)
	}
	if mainTableAlias != table {
		sqltable = sqltable.(exp.IdentifierExpression).As(mainTableAlias)
	}

	selectColumns := make([]any, len(columns))
	for i, col := range columns {
		selectColumns[i] = buildSqlIdentifier(table, col)
	}
	query := builder.From(sqltable).Select(selectColumns...)

	// joins
	for _, j := range joins {
		if j == nil {
			continue
		}
		joinTableAlias := aliasReference[j.JoinTable]
		if joinTableAlias == "" {
			joinTableAlias = j.JoinTable
		}
		var joinTable exp.Expression
		joinSchemaAndTable := strings.SplitN(j.JoinTable, ".", 2)
		if len(joinSchemaAndTable) == 2 {
			joinTable = goqu.S(joinSchemaAndTable[0]).Table(joinSchemaAndTable[1])
		} else {
			joinTable = goqu.T(j.JoinTable)
		}
		if joinTableAlias != j.JoinTable {
			joinTable = joinTable.(exp.IdentifierExpression).As(joinTableAlias)
		}

		joinConditions := make([]exp.Expression, 0, len(j.JoinColumnsMap))
		for joinCol, baseCol := range j.JoinColumnsMap {
			joinConditions = append(joinConditions, goqu.I(buildSqlIdentifier(joinTableAlias, joinCol)).Eq(buildSqlIdentifier(mainTableAlias, baseCol)))
		}
		if j.JoinType == innerJoin {
			query = query.InnerJoin(
				joinTable,
				goqu.On(joinConditions...),
			)
		}
	}
	// where
	goquWhere := make([]exp.Expression, len(whereClauses))
	for i, w := range whereClauses {
		goquWhere[i] = goqu.L(w)
	}
	query = query.Where(goqu.And(goquWhere...))

	sql, _, err := query.ToSQL()
	if err != nil {
		return "", err
	}
	return formatSqlQuery(sql), nil
}

func BuildSelectRecursiveQuery(
	driver, table string,
	columns []string,
	columnInfoMap map[string]*sqlmanager_shared.ColumnInfo,
	dependencies []*selfReferencingCircularDependency,
	joins []*sqlJoin,
	whereClauses []string,
) (string, error) {
	recursiveCteAlias := "related"
	var builder goqu.DialectWrapper
	if driver == sqlmanager_shared.MysqlDriver {
		opts := goqu.DefaultDialectOptions()
		opts.QuoteRune = '`'
		opts.SupportsWithCTERecursive = true
		opts.SupportsWithCTE = true
		dialectName := "custom-mysql-dialect"
		goqu.RegisterDialect(dialectName, opts)
		builder = goqu.Dialect(dialectName)
	} else {
		builder = goqu.Dialect(driver)
	}

	sqltable := goqu.I(table)

	selectColumns := make([]any, len(columns))
	for i, col := range columns {
		colInfo := columnInfoMap[col]
		if driver == sqlmanager_shared.PostgresDriver && colInfo != nil && colInfo.DataType == "json" {
			selectColumns[i] = goqu.L("to_jsonb(?)", goqu.I(buildSqlIdentifier(table, col))).As(col)
		} else {
			selectColumns[i] = buildSqlIdentifier(table, col)
		}
	}
	selectQuery := builder.From(sqltable).Select(selectColumns...)

	initialSelect := selectQuery
	// joins
	for _, j := range joins {
		if j == nil {
			continue
		}
		joinCondition := goqu.Ex{}
		for joinCol, baseCol := range j.JoinColumnsMap {
			joinCondition[buildSqlIdentifier(j.JoinTable, joinCol)] = goqu.I(buildSqlIdentifier(j.BaseTable, baseCol))
		}
		if j.JoinType == innerJoin {
			table := goqu.I(j.JoinTable)
			initialSelect = initialSelect.InnerJoin(
				table,
				goqu.On(joinCondition),
			)
		}
	}

	// where
	goquWhere := []exp.Expression{}
	for _, w := range whereClauses {
		goquWhere = append(goquWhere, goqu.L(w))
	}
	initialSelect = initialSelect.Where(goqu.And(goquWhere...))

	// inner join on foreign keys
	goquOnOrEx := []exp.Expression{}
	for _, d := range dependencies {
		goquOnAndEx := []exp.Expression{}
		for _, fk := range d.ForeignKeyColumns {
			if len(fk) > 1 {
				for idx, col := range fk {
					goquOnAndEx = append(goquOnAndEx, goqu.Ex{buildSqlIdentifier(table, d.PrimaryKeyColumns[idx]): goqu.I(buildSqlIdentifier(recursiveCteAlias, col))})
				}
				goquOnOrEx = append(goquOnOrEx, goqu.And(goquOnAndEx...))
			} else {
				goquOnOrEx = append(goquOnOrEx, goqu.Ex{buildSqlIdentifier(table, d.PrimaryKeyColumns[0]): goqu.I(buildSqlIdentifier(recursiveCteAlias, fk[0]))})
			}
		}
	}

	recursiveSelect := selectQuery
	recursiveSelect = recursiveSelect.InnerJoin(goqu.I(recursiveCteAlias), goqu.On(goqu.Or(goquOnOrEx...)))

	// union
	unionQuery := initialSelect.Union(recursiveSelect)

	selectCols := make([]any, len(columns))
	for i, col := range columns {
		selectCols[i] = col
	}
	recursiveQuery := builder.From(goqu.T(recursiveCteAlias)).WithRecursive(recursiveCteAlias, unionQuery).SelectDistinct(selectCols...)
	sql, _, err := recursiveQuery.ToSQL()
	if err != nil {
		return "", err
	}
	return formatSqlQuery(sql), nil
}

func BuildInsertQuery(
	driver, table string,
	columns []string,
	values [][]any,
	onConflictDoNothing *bool,
) (string, error) {
	builder := goqu.Dialect(driver)
	sqltable := goqu.I(table)
	insertCols := make([]any, len(columns))
	for i, col := range columns {
		insertCols[i] = col
	}
	insert := builder.Insert(sqltable).Cols(insertCols...)
	for _, row := range values {
		insert = insert.Vals(row)
	}
	// adds on conflict do nothing to insert query
	if *onConflictDoNothing {
		insert = insert.OnConflict(goqu.DoNothing())
	}

	query, _, err := insert.ToSQL()
	if err != nil {
		return "", err
	}
	return query, nil
}

func BuildUpdateQuery(
	driver, table string,
	insertColumns []string,
	whereColumns []string,
	columnValueMap map[string]any,
) (string, error) {
	builder := goqu.Dialect(driver)
	sqltable := goqu.I(table)

	updateRecord := goqu.Record{}
	for _, col := range insertColumns {
		val := columnValueMap[col]
		if val == "DEFAULT" {
			updateRecord[col] = goqu.L("DEFAULT")
		} else {
			updateRecord[col] = val
		}
	}

	where := []exp.Expression{}
	for _, col := range whereColumns {
		val := columnValueMap[col]
		where = append(where, goqu.Ex{col: val})
	}

	update := builder.Update(sqltable).
		Set(updateRecord).
		Where(where...)

	query, _, err := update.ToSQL()
	if err != nil {
		return "", err
	}
	return query, nil
}

func BuildTruncateQuery(
	driver, table string,
) (string, error) {
	builder := goqu.Dialect(driver)
	sqltable := goqu.I(table)
	truncate := builder.Truncate(sqltable)
	query, _, err := truncate.ToSQL()
	if err != nil {
		return "", err
	}
	return query, nil
}

func print(label string, input any) {
	bits, _ := json.Marshal(input)
	fmt.Println("======================")
	fmt.Println(label)
	fmt.Println(string(bits))
	fmt.Println("======================")

}

// returns map of schema.table -> select query
func BuildSelectQueryMap(
	driver string,
	tableDependencies map[string][]*sqlmanager_shared.ForeignConstraint,
	runConfigs []*tabledependency.RunConfig,
	subsetByForeignKeyConstraints bool,
	groupedColumnInfo map[string]map[string]*sqlmanager_shared.ColumnInfo,
) (map[string]map[tabledependency.RunType]string, error) {
	// print("driver", driver)
	// print("table dependencies", tableDependencies)
	// print("runConfigs", runConfigs)
	// print("subsetByFKConstraints", subsetByForeignKeyConstraints)
	// print("grouped column info", groupedColumnInfo)

	insertRunConfigMap := map[string]*tabledependency.RunConfig{}
	for _, cfg := range runConfigs {
		if cfg.RunType == tabledependency.RunTypeInsert {
			insertRunConfigMap[cfg.Table] = cfg
		}
	}

	// map of table -> where clause
	tableWhereMap := map[string]string{}
	for table, config := range insertRunConfigMap {
		if config.WhereClause != nil && *config.WhereClause != "" {
			qualifiedWhere, err := qualifyWhereColumnNames(driver, *config.WhereClause, table)
			if err != nil {
				return nil, err
			}
			tableWhereMap[table] = qualifiedWhere
		}
	}

	if !subsetByForeignKeyConstraints || len(tableDependencies) == 0 || len(tableWhereMap) == 0 {
		queryRunTypeMap, err := buildQueryMapNoSubsetConstraints(driver, runConfigs)
		if err != nil {
			return nil, err
		}
		return queryRunTypeMap, nil
	}

	subsetConfigs, aliasReference, err := buildTableSubsetQueryConfigs(driver, tableDependencies, tableWhereMap, insertRunConfigMap)
	if err != nil {
		return nil, err
	}

	queryRunTypeMap := map[string]map[tabledependency.RunType]string{}
	for _, runConfig := range runConfigs {
		if _, ok := queryRunTypeMap[runConfig.Table]; !ok {
			queryRunTypeMap[runConfig.Table] = map[tabledependency.RunType]string{}
		}
		columns := runConfig.SelectColumns
		subsetConfig := subsetConfigs[runConfig.Table]
		columnInfoMap := groupedColumnInfo[runConfig.Table]
		sql, err := buildTableQuery(driver, runConfig.Table, columns, subsetConfig, columnInfoMap, aliasReference)
		if err != nil {
			return nil, err
		}
		queryRunTypeMap[runConfig.Table][runConfig.RunType] = sql
	}
	return queryRunTypeMap, nil
}

func buildTableQuery(
	driver, table string,
	columns []string,
	config *subsetQueryConfig,
	columnInfoMap map[string]*sqlmanager_shared.ColumnInfo,
	aliasReference map[string]string,
) (string, error) {
	if len(config.SelfReferencingCircularDependency) != 0 {
		sql, err := BuildSelectRecursiveQuery(
			driver,
			table,
			columns,
			columnInfoMap,
			config.SelfReferencingCircularDependency,
			config.Joins,
			config.WhereClauses,
		)
		if err != nil {
			return "", fmt.Errorf("unable to build recursive select query: %w", err)
		}
		return sql, err
	} else if len(config.Joins) == 0 {
		where := strings.Join(config.WhereClauses, " AND ")
		sql, err := BuildSelectQuery(driver, table, columns, &where)
		if err != nil {
			return "", fmt.Errorf("unable to build select query: %w", err)
		}
		return sql, nil
	} else {
		sql, err := BuildSelectJoinQuery(driver, table, columns, config.Joins, config.WhereClauses, aliasReference)
		if err != nil {
			return "", fmt.Errorf("unable to build select query with joins: %w", err)
		}
		return sql, nil
	}
}

func buildQueryMapNoSubsetConstraints(
	driver string,
	runConfigs []*tabledependency.RunConfig,
) (map[string]map[tabledependency.RunType]string, error) {
	queryRunTypeMap := map[string]map[tabledependency.RunType]string{}
	for _, config := range runConfigs {
		if _, ok := queryRunTypeMap[config.Table]; !ok {
			queryRunTypeMap[config.Table] = map[tabledependency.RunType]string{}
		}
		query, err := BuildSelectQuery(driver, config.Table, config.SelectColumns, config.WhereClause)
		if err != nil {
			return nil, fmt.Errorf("unable to build select query: %w", err)
		}
		queryRunTypeMap[config.Table][config.RunType] = query
	}
	return queryRunTypeMap, nil
}

type tableSubset struct {
	Joins        []*sqlJoin
	WhereClauses []string
}

// recusively builds join for subset table
func buildSubsetJoins(table string, data map[string][]*SubsetColumnConstraint, whereClauses map[string]string, visited map[string]bool) *tableSubset {
	joins := []*sqlJoin{}
	wheres := []string{}

	if seen := visited[table]; seen {
		return &tableSubset{
			Joins:        joins,
			WhereClauses: wheres,
		}
	}
	visited[table] = true

	if condition, exists := whereClauses[table]; exists {
		wheres = append(wheres, condition)
	}

	if columns, exists := data[table]; exists {
		for _, col := range columns {
			if col.ForeignKey.Table == "" && col.ForeignKey.Columns == nil {
				continue
			}
			if !visited[col.ForeignKey.Table] {
				// handle aliased table
				// var alias *string
				joinTable := col.ForeignKey.Table
				if col.ForeignKey.OriginalTable != nil && *col.ForeignKey.OriginalTable != "" {
					// alias = &col.ForeignKey.Table
					joinTable = *col.ForeignKey.OriginalTable
				}

				joinColMap := map[string]string{}
				for idx, c := range col.ForeignKey.Columns {
					joinColMap[c] = col.Columns[idx]
				}
				joins = append(joins, &sqlJoin{
					JoinType:  innerJoin,
					JoinTable: joinTable,
					BaseTable: table,
					// Alias:          alias,
					JoinColumnsMap: joinColMap,
				})

				sub := buildSubsetJoins(col.ForeignKey.Table, data, whereClauses, visited)
				joins = append(joins, sub.Joins...)
				wheres = append(wheres, sub.WhereClauses...)
			}
		}
	}
	return &tableSubset{
		Joins:        joins,
		WhereClauses: wheres,
	}
}

type subsetQueryConfig struct {
	Joins                             []*sqlJoin
	WhereClauses                      []string
	SelfReferencingCircularDependency []*selfReferencingCircularDependency
}

func buildTableSubsetQueryConfigs(
	driver string,
	tableConstraints map[string][]*sqlmanager_shared.ForeignConstraint,
	whereClauses map[string]string,
	runConfigMap map[string]*tabledependency.RunConfig) (map[string]*subsetQueryConfig, map[string]string, error) {
	configs := map[string]*subsetQueryConfig{}

	filteredConstraints := filterForeignKeysWithSubset(runConfigMap, tableConstraints, whereClauses)
	subset, err := buildAliasReferences(driver, filteredConstraints, whereClauses)
	if err != nil {
		return nil, nil, err
	}

	for table := range subset.ColumnConstraints {
		visited := map[string]bool{}
		tableSubset := buildSubsetJoins(table, subset.ColumnConstraints, subset.WhereClauses, visited)
		constraints := tableConstraints[table]
		selfRefCd := getSelfReferencingColumns(table, constraints)
		configs[table] = &subsetQueryConfig{
			Joins:                             tableSubset.Joins,
			WhereClauses:                      tableSubset.WhereClauses,
			SelfReferencingCircularDependency: selfRefCd,
		}
	}

	// Populate aliasReference based on the subset information
	return configs, subset.AliasReferences, nil
}

type subsetConstraints struct {
	ColumnConstraints map[string][]*SubsetColumnConstraint
	WhereClauses      map[string]string
	AliasReferences   map[string]string
}

func buildAliasReferences(driver string, constraints map[string][]*sqlmanager_shared.ForeignConstraint, whereClauses map[string]string) (*subsetConstraints, error) {
	updatedConstraints := map[string][]*SubsetColumnConstraint{}
	aliasReference := map[string]string{} // alias name to table name
	updatedWheres := map[string]string{}

	for table, where := range whereClauses {
		updatedWheres[table] = where
	}

	for table, fkConstrantDefs := range constraints {
		if len(fkConstrantDefs) == 0 {
			updatedConstraints[table] = []*SubsetColumnConstraint{}
		} else {
			updatedConstraints[table] = processAliasConstraints(table, fkConstrantDefs, updatedConstraints, aliasReference)
		}
	}

	if err := updateAliasReferences(driver, updatedConstraints, aliasReference, updatedWheres); err != nil {
		return nil, err
	}

	return &subsetConstraints{
		ColumnConstraints: updatedConstraints,
		WhereClauses:      updatedWheres,
		AliasReferences:   aliasReference,
	}, nil
}

// creates alias table reference if there is a double reference
func processAliasConstraints(
	table string,
	fkConstraintDefs []*sqlmanager_shared.ForeignConstraint,
	subsetConstraintsMap map[string][]*SubsetColumnConstraint,
	aliasReferenceMap map[string]string,
	// seenTables map[string]struct{},
) []*SubsetColumnConstraint {
	if _, exists := subsetConstraintsMap[table]; exists {
		return subsetConstraintsMap[table]
	}

	tableCount := map[string]int{}
	for _, colDef := range fkConstraintDefs {
		if colDef.ForeignKey != nil {
			tableCount[colDef.ForeignKey.Table]++
		}
	}

	newColDefs := []*SubsetColumnConstraint{}
	for _, colDef := range fkConstraintDefs {
		if colDef.ForeignKey.Table == table {
			continue // self reference skip
		}

		if count := tableCount[colDef.ForeignKey.Table]; count > 1 {
			// create aliased table
			newTable := fmt.Sprintf("%s_%s", strings.ReplaceAll(colDef.ForeignKey.Table, ".", "_"), strings.Join(colDef.Columns, "_"))
			alias := aliasHash(newTable)
			aliasReferenceMap[alias] = colDef.ForeignKey.Table
			newColDefs = append(newColDefs, &SubsetColumnConstraint{
				Columns: colDef.Columns,
				ForeignKey: &SubsetReferenceKey{
					Table:         alias,
					OriginalTable: &colDef.ForeignKey.Table,
					Columns:       colDef.ForeignKey.Columns,
				},
			})
		} else {
			newColDefs = append(newColDefs, &SubsetColumnConstraint{
				Columns: colDef.Columns,
				ForeignKey: &SubsetReferenceKey{
					Table:   colDef.ForeignKey.Table,
					Columns: colDef.ForeignKey.Columns,
				},
			})
		}
	}

	subsetConstraintsMap[table] = newColDefs
	return newColDefs
}

type aliasNode struct {
	alias string
	table string
	edges []*aliasNode
	color int // 0: white (unvisited), 1: gray (visiting), 2: black (visited)
}

func updateAliasReferences(
	driver string,
	updatedConstraints map[string][]*SubsetColumnConstraint,
	aliasReference map[string]string,
	updatedWheres map[string]string,
) error {
	// Build the graph
	graph := make(map[string]*aliasNode)
	for alias, table := range aliasReference {
		if _, exists := graph[alias]; !exists {
			graph[alias] = &aliasNode{alias: alias, table: table}
		}
		colDefs := updatedConstraints[table]
		for _, c := range colDefs {
			if c.ForeignKey != nil && c.ForeignKey.Table != "" {
				newAlias := aliasHash(fmt.Sprintf("%s_%s", alias, strings.ReplaceAll(c.ForeignKey.Table, ".", "_")))
				if _, exists := graph[newAlias]; !exists {
					graph[newAlias] = &aliasNode{alias: newAlias, table: c.ForeignKey.Table}
				}
				graph[alias].edges = append(graph[alias].edges, graph[newAlias])
			}
		}
	}

	// Perform topological sort
	sorted := make([]*aliasNode, 0, len(graph))
	visited := make(map[string]bool)

	var visit func(*aliasNode) error
	visit = func(node *aliasNode) error {
		if visited[node.alias] {
			return nil
		}
		if node.color == 1 {
			return fmt.Errorf("cyclic dependency detected at alias: %s", node.alias)
		}
		node.color = 1
		for _, edge := range node.edges {
			if err := visit(edge); err != nil {
				return err
			}
		}
		node.color = 2
		visited[node.alias] = true
		sorted = append(sorted, node)
		return nil
	}

	for _, node := range graph {
		if !visited[node.alias] {
			if err := visit(node); err != nil {
				// If we detect a cycle, we'll break it by removing the edge
				fmt.Printf("Warning: Cyclic dependency detected. Breaking cycle at %s\n", node.alias)
				node.edges = nil
				if err := visit(node); err != nil {
					return err
				}
			}
		}
	}

	// Process aliases in topological order
	for i := len(sorted) - 1; i >= 0; i-- {
		node := sorted[i]
		alias, table := node.alias, node.table

		colDefs := updatedConstraints[table]
		newColDefs := []*SubsetColumnConstraint{}

		for _, c := range colDefs {
			if c.ForeignKey == nil || c.ForeignKey.Table == "" {
				newColDefs = append(newColDefs, c)
				continue
			}

			newAlias := aliasHash(fmt.Sprintf("%s_%s", alias, strings.ReplaceAll(c.ForeignKey.Table, ".", "_")))
			newColDefs = append(newColDefs, &SubsetColumnConstraint{
				Columns: c.Columns,
				ForeignKey: &SubsetReferenceKey{
					Columns:       c.ForeignKey.Columns,
					OriginalTable: &c.ForeignKey.Table,
					Table:         newAlias,
				},
			})
		}

		updatedConstraints[alias] = newColDefs

		// Update the where clause
		if where, exists := updatedWheres[table]; exists && where != "" {
			aliasedWhere, err := qualifyWhereWithTableAlias(driver, where, alias)
			if err != nil {
				return err
			}
			// Store the aliased where clause with the alias as the key
			updatedWheres[alias] = aliasedWhere
			// Remove the original where clause
			delete(updatedWheres, table)
		}

		// Update aliasReference to use the original table name as the key
		aliasReference[table] = alias
		// Remove the old entry if the alias is different from the table name
		if alias != table {
			delete(aliasReference, alias)
		}
	}

	return nil
}

// Helper function to create a valid PostgreSQL alias
func createValidAlias(name string) string {
	// Remove any non-alphanumeric characters and prepend 'a_' to ensure it starts with a letter
	return "a_" + strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, name)
}

// Update the aliasHash function to use createValidAlias
func aliasHash(input string) string {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
	if len(hash) > 10 {
		hash = hash[:10]
	}
	return createValidAlias(hash)
}

// check if table or any parent table has a where clause
func shouldSubsetTable(table string, data map[string][]*sqlmanager_shared.ForeignConstraint, whereClauses map[string]string, visited map[string]bool) bool {
	if _, exists := whereClauses[table]; exists {
		return true
	}
	if visited[table] {
		return false
	}
	visited[table] = true

	if columns, exists := data[table]; exists {
		for _, col := range columns {
			if col.ForeignKey.Table != "" {
				if shouldSubsetTable(col.ForeignKey.Table, data, whereClauses, visited) {
					return true
				}
			}
		}
	}
	return false
}

// filters out any foreign keys that are not involved in the subset (where clauses) and tables not in the map
func filterForeignKeysWithSubset(runConfigMap map[string]*tabledependency.RunConfig, constraints map[string][]*sqlmanager_shared.ForeignConstraint, whereClauses map[string]string) map[string][]*sqlmanager_shared.ForeignConstraint {
	tableSubsetMap := map[string]bool{} // map of which tables to subset
	for table := range runConfigMap {
		visited := map[string]bool{}
		tableSubsetMap[table] = shouldSubsetTable(table, constraints, whereClauses, visited)
	}

	filteredConstraints := map[string][]*sqlmanager_shared.ForeignConstraint{}
	for table := range runConfigMap {
		filteredConstraints[table] = []*sqlmanager_shared.ForeignConstraint{}
		if tableConstraints, ok := constraints[table]; ok {
			for _, colDef := range tableConstraints {
				if exists := tableSubsetMap[colDef.ForeignKey.Table]; exists {
					filteredConstraints[table] = append(filteredConstraints[table], colDef)
				}
			}
		}
	}
	return filteredConstraints
}

// func aliasHash(input string) string {
// 	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
// 	if len(hash) > 14 {
// 		hash = hash[:14]
// 	}
// 	return hash
// }

func qualifyWhereWithTableAlias(driver, where, alias string) (string, error) {
	query := goqu.Dialect(driver).From(goqu.T(alias)).Select("*").Where(goqu.L(where))
	sql, _, err := query.ToSQL()
	if err != nil {
		return "", err
	}
	var updatedSql string
	switch driver {
	case sqlmanager_shared.MysqlDriver:
		sql, err := qualifyMysqlWhereColumnNames(sql, nil, alias)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	case sqlmanager_shared.PostgresDriver:
		sql, err := qualifyPostgresWhereColumnNames(sql, nil, alias)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	default:
		return "", errors.New("unsupported sql driver type")
	}
	index := strings.Index(strings.ToLower(updatedSql), "where")
	if index == -1 {
		// "where" not found
		return "", fmt.Errorf("unable to qualify where column names")
	}
	startIndex := index + len("where")
	return strings.TrimSpace(updatedSql[startIndex:]), nil
}

func qualifyWhereColumnNames(driver, where, table string) (string, error) {
	sql, err := BuildSelectQuery(driver, table, []string{"*"}, &where)
	if err != nil {
		return "", err
	}
	schema, tableName := sqlmanager_shared.SplitTableKey(table)
	var updatedSql string
	switch driver {
	case sqlmanager_shared.MysqlDriver:
		sql, err := qualifyMysqlWhereColumnNames(sql, &schema, tableName)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	case sqlmanager_shared.PostgresDriver:
		sql, err := qualifyPostgresWhereColumnNames(sql, &schema, tableName)
		if err != nil {
			return "", err
		}
		updatedSql = sql
	default:
		return "", errors.New("unsupported sql driver type")
	}
	index := strings.Index(strings.ToLower(updatedSql), "where")
	if index == -1 {
		// "where" not found
		return "", fmt.Errorf("unable to qualify where column names")
	}
	startIndex := index + len("where")

	return strings.TrimSpace(updatedSql[startIndex:]), nil
}

func getPrimaryToForeignTableMapFromRunConfigs(runConfigs []*tabledependency.RunConfig) map[string][]string {
	dpMap := map[string]map[string]struct{}{}

	for _, cfg := range runConfigs {
		if _, exists := dpMap[cfg.Table]; !exists {
			dpMap[cfg.Table] = map[string]struct{}{}
		}
		for _, dep := range cfg.DependsOn {
			if _, exists := dpMap[dep.Table]; !exists {
				dpMap[dep.Table] = map[string]struct{}{}
			}
			dpMap[dep.Table][cfg.Table] = struct{}{}
		}
	}
	tableDependencyMap := map[string][]string{}
	for table, fkTables := range dpMap {
		for t := range fkTables {
			tableDependencyMap[table] = append(tableDependencyMap[table], t)
		}
	}

	return tableDependencyMap
}

type selfReferencingCircularDependency struct {
	PrimaryKeyColumns []string
	ForeignKeyColumns [][]string
}

func getSelfReferencingColumns(table string, constraints []*sqlmanager_shared.ForeignConstraint) []*selfReferencingCircularDependency {
	result := []*selfReferencingCircularDependency{}
	resultMap := map[string]*selfReferencingCircularDependency{}

	for _, c := range constraints {
		if c.ForeignKey.Table != table {
			continue
		}
		pkHash := utils.ToSha256(fmt.Sprintf("%s%s", c.ForeignKey.Table, strings.Join(c.ForeignKey.Columns, "")))
		if _, exists := resultMap[pkHash]; !exists {
			resultMap[pkHash] = &selfReferencingCircularDependency{
				PrimaryKeyColumns: c.ForeignKey.Columns,
				ForeignKeyColumns: [][]string{},
			}
		}
		resultMap[pkHash].ForeignKeyColumns = append(resultMap[pkHash].ForeignKeyColumns, c.Columns)
	}

	for _, r := range resultMap {
		result = append(result, r)
	}

	return result
}

func formatSqlQuery(sql string) string {
	return fmt.Sprintf("%s;", sql)
}

func buildSqlIdentifier(identifiers ...string) string {
	return strings.Join(identifiers, ".")
}

func qualifyPostgresWhereColumnNames(sql string, schema *string, table string) (string, error) {
	tree, err := pgquery.Parse(sql)
	if err != nil {
		return "", err
	}

	for _, stmt := range tree.GetStmts() {
		selectStmt := stmt.GetStmt().GetSelectStmt()

		if selectStmt.WhereClause != nil {
			updatePostgresExpr(schema, table, selectStmt.WhereClause)
		}
	}
	updatedSql, err := pgquery.Deparse(tree)
	if err != nil {
		return "", err
	}
	return updatedSql, nil
}

func updatePostgresExpr(schema *string, table string, node *pg_query.Node) {
	switch expr := node.Node.(type) {
	case *pg_query.Node_SubLink:
		updatePostgresExpr(schema, table, node.GetSubLink().GetTestexpr())
	case *pg_query.Node_BoolExpr:
		for _, arg := range expr.BoolExpr.GetArgs() {
			updatePostgresExpr(schema, table, arg)
		}
	case *pg_query.Node_AExpr:
		updatePostgresExpr(schema, table, expr.AExpr.GetLexpr())
		updatePostgresExpr(schema, table, expr.AExpr.Rexpr)
	case *pg_query.Node_ColumnDef:
	case *pg_query.Node_ColumnRef:
		col := node.GetColumnRef()
		isQualified := false
		var colName *string
		// find col name and check if already has schema + table name
		for _, f := range col.Fields {
			val := f.GetString_().GetSval()
			if schema != nil && val == *schema {
				continue
			}
			if val == table {
				isQualified = true
				break
			}
			colName = &val
		}
		if !isQualified && colName != nil && *colName != "" {
			fields := []*pg_query.Node{}
			if schema != nil && *schema != "" {
				fields = append(fields, pg_query.MakeStrNode(*schema))
			}
			fields = append(fields, []*pg_query.Node{
				pg_query.MakeStrNode(table),
				pg_query.MakeStrNode(*colName),
			}...)
			col.Fields = fields
		}
	}
}

func qualifyMysqlWhereColumnNames(sql string, schema *string, table string) (string, error) {
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return "", err
	}

	switch stmt := stmt.(type) { //nolint:gocritic
	case *sqlparser.Select:
		err = sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
			switch node := node.(type) { //nolint:gocritic
			case *sqlparser.ComparisonExpr:
				if col, ok := node.Left.(*sqlparser.ColName); ok {
					s := ""
					if schema != nil && *schema != "" {
						s = *schema
					}
					col.Qualifier.Qualifier = sqlparser.NewTableIdent(s)
					col.Qualifier.Name = sqlparser.NewTableIdent(table)
				}
				return false, nil
			}
			return true, nil
		}, stmt)
		if err != nil {
			return "", err
		}
	}

	return sqlparser.String(stmt), nil
}
