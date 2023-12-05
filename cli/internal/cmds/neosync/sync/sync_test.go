package sync_cmd

import (
	"math"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func Test_groupConfigsByDependency(t *testing.T) {
	configs := []*benthosConfigResponse{
		{
			Name:      "ConfigA",
			DependsOn: []string{},
		},
		{
			Name:      "ConfigB",
			DependsOn: []string{"ConfigA"},
		},
		{
			Name:      "ConfigC",
			DependsOn: []string{"ConfigA"},
		},
		{
			Name:      "ConfigD",
			DependsOn: []string{"ConfigB", "ConfigC"},
		},
	}

	expected := [][]*benthosConfigResponse{
		{
			{
				Name:      "ConfigA",
				DependsOn: []string{},
			},
		},
		{
			{
				Name:      "ConfigB",
				DependsOn: []string{"ConfigA"},
			},
			{
				Name:      "ConfigC",
				DependsOn: []string{"ConfigA"},
			},
		},
		{
			{
				Name:      "ConfigD",
				DependsOn: []string{"ConfigB", "ConfigC"},
			},
		},
	}
	groups := groupConfigsByDependency(configs)

	for i, group := range groups {
		assert.Equal(t, len(group), len(expected[i]))
		for j, cfg := range group {
			assert.Equal(t, cfg.Name, expected[i][j].Name)
			assert.ElementsMatch(t, cfg.DependsOn, expected[i][j].DependsOn)
		}

	}
}

func Test_groupConfigsByDependency_NoDependency(t *testing.T) {
	configs := []*benthosConfigResponse{
		{
			Name:      "ConfigA",
			DependsOn: []string{},
		},
		{
			Name:      "ConfigB",
			DependsOn: []string{},
		},
	}

	expected := [][]*benthosConfigResponse{
		{
			{
				Name:      "ConfigA",
				DependsOn: []string{},
			},
			{
				Name:      "ConfigB",
				DependsOn: []string{},
			},
		},
	}
	groups := groupConfigsByDependency(configs)

	for i, group := range groups {
		assert.Equal(t, len(group), len(expected[i]))
		for j, cfg := range group {
			assert.Equal(t, cfg.Name, expected[i][j].Name)
			assert.ElementsMatch(t, cfg.DependsOn, expected[i][j].DependsOn)
		}
	}
}

func Test_buildPlainInsertArgs(t *testing.T) {
	assert.Empty(t, buildPlainInsertArgs(nil))
	assert.Empty(t, buildPlainInsertArgs([]string{}))
	assert.Equal(t, buildPlainInsertArgs([]string{"foo", "bar", "baz"}), "root = [this.foo, this.bar, this.baz]")
}

func Test_clampInt(t *testing.T) {
	assert.Equal(t, clampInt(0, 1, 2), 1)
	assert.Equal(t, clampInt(1, 1, 2), 1)
	assert.Equal(t, clampInt(2, 1, 2), 2)
	assert.Equal(t, clampInt(3, 1, 2), 2)
	assert.Equal(t, clampInt(1, 1, 1), 1)

	assert.Equal(t, clampInt(1, 3, 2), 3, "low is evaluated first, order is relevant")

}

func Test_computeMaxPgBatchCount(t *testing.T) {
	assert.Equal(t, computeMaxPgBatchCount(65535), 1)
	assert.Equal(t, computeMaxPgBatchCount(65536), 1, "anything over max should clamp to 1")
	assert.Equal(t, computeMaxPgBatchCount(math.MaxInt), 1, "anything over pgmax should clamp to 1")
	assert.Equal(t, computeMaxPgBatchCount(1), 65535)
	assert.Equal(t, computeMaxPgBatchCount(0), 65535)
}

func Test_getSchemaTables(t *testing.T) {
	tests := []struct {
		name     string
		schemas  []*mgmtv1alpha1.DatabaseColumn
		expected []*SqlTable
	}{
		{
			name: "single table single column",
			schemas: []*mgmtv1alpha1.DatabaseColumn{
				{Schema: "schema1", Table: "table1", Column: "column1"},
			},
			expected: []*SqlTable{
				{Schema: "schema1", Table: "table1", Columns: []string{"column1"}},
			},
		},
		{
			name: "single table multiple columns",
			schemas: []*mgmtv1alpha1.DatabaseColumn{
				{Schema: "schema1", Table: "table1", Column: "column1"},
				{Schema: "schema1", Table: "table1", Column: "column2"},
			},
			expected: []*SqlTable{
				{Schema: "schema1", Table: "table1", Columns: []string{"column1", "column2"}},
			},
		},
		{
			name: "multiple tables and columns",
			schemas: []*mgmtv1alpha1.DatabaseColumn{
				{Schema: "schema1", Table: "table1", Column: "column1"},
				{Schema: "schema1", Table: "table2", Column: "column2"},
				{Schema: "schema2", Table: "table1", Column: "column3"},
			},
			expected: []*SqlTable{
				{Schema: "schema1", Table: "table1", Columns: []string{"column1"}},
				{Schema: "schema1", Table: "table2", Columns: []string{"column2"}},
				{Schema: "schema2", Table: "table1", Columns: []string{"column3"}},
			},
		},
	}

	for _, tt := range tests {
		result := getSchemaTables(tt.schemas)
		assert.ElementsMatch(t, tt.expected, result)
	}
}
