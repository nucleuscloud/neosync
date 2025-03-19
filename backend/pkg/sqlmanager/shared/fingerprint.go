package sqlmanager_shared

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

// buildForeignKeyConstraintFingerprint creates a stable hash that
// excludes the constraint name but includes everything else
// that defines the constraint.
func BuildForeignKeyConstraintFingerprint(fk *ForeignKeyConstraint) string {
	// Make local copies of slices so we can sort without side effects.
	referencingCols := append([]string{}, fk.ReferencingColumns...)
	referencedCols := append([]string{}, fk.ReferencedColumns...)
	sort.Strings(referencingCols)
	sort.Strings(referencedCols)

	// Convert bool slice to string for hashing
	notNullableStr := boolSliceToString(fk.NotNullable)

	// Build a canonical string that includes:
	// referencing_schema, referencing_table, referencing_columns,
	// referenced_schema, referenced_table, referenced_columns,
	// constraint_type, notNullable, updateRule, deleteRule
	parts := []string{
		fk.ReferencingSchema,
		fk.ReferencingTable,
		strings.Join(referencingCols, ","),
		fk.ReferencedSchema,
		fk.ReferencedTable,
		strings.Join(referencedCols, ","),
		fk.ConstraintType,
		notNullableStr,
		ptrOrEmpty(fk.UpdateRule),
		ptrOrEmpty(fk.DeleteRule),
	}

	return BuildFingerprint(parts...)
}

// buildNonForeignKeyConstraintFingerprint creates a stable hash that
// excludes the constraint name but includes everything else
// that defines the constraint (columns, definition, etc.).
func BuildNonForeignKeyConstraintFingerprint(nf *NonForeignKeyConstraint) string {
	// Sort columns
	sortedCols := append([]string{}, nf.Columns...)
	sort.Strings(sortedCols)

	// For PK/Unique, the definition might be empty or irrelevant.
	// For CHECK constraints, definition typically is the check expression.
	// So we include definition as well, if present.
	parts := []string{
		nf.ConstraintType,
		nf.SchemaName,
		nf.TableName,
		strings.Join(sortedCols, ","),
		nf.Definition,
	}

	return BuildFingerprint(parts...)
}

// BuildFingerprint creates a stable hash for a table trigger that includes
// schema, table, trigger name, and trigger definition.
func BuildTriggerFingerprint(trigger *TableTrigger) string {
	parts := []string{
		trigger.Schema,
		trigger.Table,
		trigger.TriggerName,
		trigger.Definition,
	}

	if trigger.TriggerSchema != nil && *trigger.TriggerSchema != "" {
		parts = append(parts, *trigger.TriggerSchema)
	}

	return BuildFingerprint(parts...)
}

func BuildFingerprint(input ...string) string {
	h := sha256.New()
	for _, i := range input {
		h.Write([]byte(i))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func BuildTableColumnFingerprint(column *TableColumn) string {
	parts := []string{
		column.Schema,
		column.Table,
		column.Name,
		column.DataType,
		strconv.FormatBool(column.IsNullable),
		column.ColumnDefault,
		ptrOrEmpty(column.ColumnDefaultType),
		ptrOrEmpty(column.IdentityGeneration),
		ptrOrEmpty(column.GeneratedType),
		ptrOrEmpty(column.GeneratedExpression),
	}

	input := strings.Join(parts, "|")
	return sha256Hex(input)
}

// ptrOrEmpty returns the pointer's value if not nil, otherwise "".
func ptrOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// boolSliceToString converts a slice of booleans to a short
// string representation e.g. [true,false] => "1,0"
func boolSliceToString(vals []bool) string {
	if len(vals) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, v := range vals {
		if i > 0 {
			sb.WriteRune(',')
		}
		if v {
			sb.WriteRune('1')
		} else {
			sb.WriteRune('0')
		}
	}
	return sb.String()
}
