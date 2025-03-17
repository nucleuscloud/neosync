package schemamanager_shared

import (
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_BuildOrderedForeignKeyConstraintsToDrop(t *testing.T) {
	logger := testutil.GetTestLogger(t)

	tests := []struct {
		name     string
		diff     *SchemaDifferences
		expected []*sqlmanager_shared.ForeignKeyConstraint
	}{
		{
			name: "No foreign keys to drop",
			diff: &SchemaDifferences{
				ExistsInDestination: &ExistsInDestination{
					ForeignKeyConstraints: []*sqlmanager_shared.ForeignKeyConstraint{},
				},
			},
			expected: nil,
		},
		{
			name: "Single foreign key",
			diff: &SchemaDifferences{
				ExistsInDestination: &ExistsInDestination{
					ForeignKeyConstraints: []*sqlmanager_shared.ForeignKeyConstraint{
						{
							ReferencedSchema:  "public",
							ReferencedTable:   "parent",
							ReferencingSchema: "public",
							ReferencingTable:  "child",
							ConstraintName:    "fk_child_parent",
							ConstraintType:    "FOREIGN KEY",
						},
					},
				},
			},
			expected: []*sqlmanager_shared.ForeignKeyConstraint{
				{
					ReferencedSchema:  "public",
					ReferencedTable:   "parent",
					ReferencingSchema: "public",
					ReferencingTable:  "child",
					ConstraintName:    "fk_child_parent",
					ConstraintType:    "FOREIGN KEY",
				},
			},
		},
		{
			name: "Self-referencing foreign key",
			diff: &SchemaDifferences{
				ExistsInDestination: &ExistsInDestination{
					ForeignKeyConstraints: []*sqlmanager_shared.ForeignKeyConstraint{
						{
							ReferencedSchema:  "public",
							ReferencedTable:   "self_ref",
							ReferencingSchema: "public",
							ReferencingTable:  "self_ref",
							ConstraintName:    "fk_self_ref",
							ConstraintType:    "FOREIGN KEY",
						},
					},
				},
			},
			expected: []*sqlmanager_shared.ForeignKeyConstraint{
				{
					ReferencedSchema:  "public",
					ReferencedTable:   "self_ref",
					ReferencingSchema: "public",
					ReferencingTable:  "self_ref",
					ConstraintName:    "fk_self_ref",
					ConstraintType:    "FOREIGN KEY",
				},
			},
		},
		{
			name: "Multiple foreign keys with cycle",
			diff: &SchemaDifferences{
				ExistsInDestination: &ExistsInDestination{
					ForeignKeyConstraints: []*sqlmanager_shared.ForeignKeyConstraint{
						{
							ReferencedSchema:  "public",
							ReferencedTable:   "parent",
							ReferencingSchema: "public",
							ReferencingTable:  "child",
							ConstraintName:    "fk_child_parent",
							ConstraintType:    "FOREIGN KEY",
						},
						{
							ReferencedSchema:  "public",
							ReferencedTable:   "child",
							ReferencingSchema: "public",
							ReferencingTable:  "parent",
							ConstraintName:    "fk_parent_child",
							ConstraintType:    "FOREIGN KEY",
						},
					},
				},
			},
			expected: []*sqlmanager_shared.ForeignKeyConstraint{
				{
					ReferencedSchema:  "public",
					ReferencedTable:   "parent",
					ReferencingSchema: "public",
					ReferencingTable:  "child",
					ConstraintName:    "fk_child_parent",
					ConstraintType:    "FOREIGN KEY",
				},
				{
					ReferencedSchema:  "public",
					ReferencedTable:   "child",
					ReferencingSchema: "public",
					ReferencingTable:  "parent",
					ConstraintName:    "fk_parent_child",
					ConstraintType:    "FOREIGN KEY",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildOrderedForeignKeyConstraintsToDrop(logger, tt.diff)
			assert.Equal(t, len(tt.expected), len(result), "expected %d foreign keys, got %d", len(tt.expected), len(result))
			for i, fk := range result {
				assert.Equal(t, tt.expected[i].ConstraintName, fk.ConstraintName, "expected foreign key %s, got %s", tt.expected[i].ConstraintName, fk.ConstraintName)
			}
		})
	}
}
