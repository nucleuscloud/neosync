package testutil_testdata

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FormatTimeForComparison(t time.Time) string {
	// Ensure we're working with UTC
	t = t.UTC()
	year := t.Year()

	// Handle BC dates (negative years and year 0)
	if year <= 0 {
		// Convert to PostgreSQL's BC format
		// PostgreSQL BC years start from 1, so we need to adjust
		bcYear := -year + 1
		return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d+00 BC",
			bcYear,
			t.Month(),
			t.Day(),
			t.Hour(),
			t.Minute(),
			t.Second())
	}

	// Handle AD dates (positive years)
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d+00",
		year,
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second())
}

func fetchSQLRows(ctx context.Context, db *sql.DB, schema, table, driver, idCol string) (map[string]map[string]any, error) {
	query := goqu.Dialect(driver).From(goqu.S(schema).Table(table)).Order(goqu.C(idCol).Asc())
	generatedSql, _, err := query.ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, generatedSql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make(map[string]map[string]any)
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]any)
		var uniqueKey string

		for i, col := range columns {
			val := valuePtrs[i].(*any)
			rowMap[col] = *val
			if col == idCol {
				uniqueKey = fmt.Sprintf("%v", *val)
			}
		}

		data[uniqueKey] = rowMap
	}

	return data, nil
}

func VerifySQLTableColumnValues(
	t *testing.T,
	ctx context.Context,
	source *sql.DB,
	target *sql.DB,
	schema, table, driver,
	idCol string,
) {
	// Fetch rows from both databases
	sourceRows, err := fetchSQLRows(ctx, source, schema, table, driver, idCol)
	require.NoErrorf(t, err, "Error fetching source rows from table %s", table)

	targetRows, err := fetchSQLRows(ctx, target, schema, table, driver, idCol)
	require.NoErrorf(t, err, "Error fetching target rows from table %s", table)

	query := goqu.Dialect(driver).From(goqu.S(schema).Table(table)).Limit(0)
	generatedSql, _, err := query.ToSQL()
	require.NoErrorf(t, err, "Error building query for table %s", table)

	rows, err := target.QueryContext(ctx, generatedSql)
	require.NoErrorf(t, err, "Error querying table %s", table)
	columns, err := rows.Columns()
	require.NoErrorf(t, err, "Error getting columns for table %s", table)
	colTypes, err := rows.ColumnTypes()
	require.NoErrorf(t, err, "Error getting column types for table %s", table)

	colTypesMap := make(map[string]string)
	for i, col := range columns {
		colTypesMap[col] = colTypes[i].DatabaseTypeName()
	}

	// Compare rows
	for key, sourceRow := range sourceRows {
		targetRow, exists := targetRows[key]
		assert.Truef(t, exists, "Row %s exists in source but not in target for table %s", key, table)

		if !exists {
			continue
		}

		for col, sourceValue := range sourceRow {
			targetValue := targetRow[col]
			colType := colTypesMap[col]
			if isJsonType(colType) && sourceValue != nil && targetValue != nil {
				// Parse JSON values for comparison
				var sourceJson, destJson any
				err := json.Unmarshal(sourceValue.([]byte), &sourceJson)
				assert.NoErrorf(t, err, "Error unmarshaling source JSON in table %s", table)
				err = json.Unmarshal(targetValue.([]byte), &destJson)
				assert.NoErrorf(t, err, "Error unmarshaling target JSON in table %s", table)
				assert.Equalf(t, sourceJson, destJson, "JSON difference in row %s, column %s for table %s", key, col, table)
			} else if isJsonArrayType(colType) && sourceValue != nil && targetValue != nil {
				// Handle Postgres array format "{{}}" and remove escaping and spaces
				var sourceStr, targetStr string
				switch s := sourceValue.(type) {
				case []byte:
					sourceStr = string(s)
				case string:
					sourceStr = s
				}
				switch t := targetValue.(type) {
				case []byte:
					targetStr = string(t)
				case string:
					targetStr = t
				}
				sourceStr = strings.ReplaceAll(strings.ReplaceAll(sourceStr, "\\", ""), " ", "")
				targetStr = strings.ReplaceAll(strings.ReplaceAll(targetStr, "\\", ""), " ", "")
				assert.Equalf(t, sourceStr, targetStr, "Array difference in row %s, column %s for table %s", key, col, table)
			} else {
				assert.Equalf(t, sourceValue, targetValue, "Difference in row %s, column %s for table %s: source=%v, target=%v", key, col, table, sourceValue, targetValue)
			}
		}
	}

	for key := range targetRows {
		_, exists := sourceRows[key]
		assert.Truef(t, exists, "Row %s exists in target but not in source for table %s", key, table)
	}
}

func isJsonType(colType string) bool {
	return strings.EqualFold(colType, "json") || strings.EqualFold(colType, "jsonb")
}

func isJsonArrayType(colType string) bool {
	return strings.EqualFold(colType, "json[]") || strings.EqualFold(colType, "jsonb[]") || strings.EqualFold(colType, "_json") || strings.EqualFold(colType, "_jsonb")
}
