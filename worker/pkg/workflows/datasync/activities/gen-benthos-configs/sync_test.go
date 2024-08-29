package genbenthosconfigs_activity

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/stretchr/testify/require"
)

func TestFilterForeignKeysMap(t *testing.T) {
	tests := []struct {
		name              string
		colTransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer
		foreignKeysMap    map[string][]*sqlmanager_shared.ForeignConstraint
		expected          map[string][]*sqlmanager_shared.ForeignConstraint
	}{
		{
			name:              "Empty input maps",
			colTransformerMap: map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{},
			foreignKeysMap:    map[string][]*sqlmanager_shared.ForeignConstraint{},
			expected:          map[string][]*sqlmanager_shared.ForeignConstraint{},
		},
		{
			name: "No matching tables",
			colTransformerMap: map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{
				"table1": {"col1": &mgmtv1alpha1.JobMappingTransformer{}},
			},
			foreignKeysMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"table2": {
					{
						Columns:     []string{"col1"},
						NotNullable: []bool{true},
						ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "ref_table", Columns: []string{"ref_col"}},
					},
				},
			},
			expected: map[string][]*sqlmanager_shared.ForeignConstraint{},
		},
		{
			name: "Filtered composite foreign keys",
			colTransformerMap: map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{
				"table1": {
					"col1": &mgmtv1alpha1.JobMappingTransformer{Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH},
					"col2": &mgmtv1alpha1.JobMappingTransformer{Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL},
				},
			},
			foreignKeysMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"table1": {
					{
						Columns:     []string{"col1", "col2"},
						NotNullable: []bool{true, false},
						ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "ref_table", Columns: []string{"ref_col1", "ref_col2"}},
					},
				},
			},
			expected: map[string][]*sqlmanager_shared.ForeignConstraint{
				"table1": {
					{
						Columns:     []string{"col1"},
						NotNullable: []bool{true},
						ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "ref_table", Columns: []string{"ref_col1"}},
					},
				},
			},
		},
		{
			name: "Filtered foreign keys",
			colTransformerMap: map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{
				"table1": {
					"col1": &mgmtv1alpha1.JobMappingTransformer{Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH},
				},
				"table2": {
					"col2": &mgmtv1alpha1.JobMappingTransformer{Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL},
				},
			},
			foreignKeysMap: map[string][]*sqlmanager_shared.ForeignConstraint{
				"table1": {
					{
						Columns:     []string{"col1"},
						NotNullable: []bool{true},
						ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "ref_table", Columns: []string{"ref_col1"}},
					},
				},
				"table2": {
					{
						Columns:     []string{"col2"},
						NotNullable: []bool{false},
						ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "ref_table", Columns: []string{"ref_col2"}},
					},
				},
			},
			expected: map[string][]*sqlmanager_shared.ForeignConstraint{
				"table1": {
					{
						Columns:     []string{"col1"},
						NotNullable: []bool{true},
						ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "ref_table", Columns: []string{"ref_col1"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterForeignKeysMap(tt.colTransformerMap, tt.foreignKeysMap)
			require.Equal(t, tt.expected, result)
		})
	}
}
