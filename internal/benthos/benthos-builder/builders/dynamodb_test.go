package benthosbuilder_builders

import (
	"reflect"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func Test_getWhereFromSourceTableOption(t *testing.T) {
	actual := getWhereFromSourceTableOption(nil)
	require.Nil(t, actual)

	actual = getWhereFromSourceTableOption(&mgmtv1alpha1.DynamoDBSourceTableOption{})
	require.Nil(t, actual)

	actual = getWhereFromSourceTableOption(&mgmtv1alpha1.DynamoDBSourceTableOption{WhereClause: nil})
	require.Nil(t, actual)

	where := "foo"
	actual = getWhereFromSourceTableOption(&mgmtv1alpha1.DynamoDBSourceTableOption{WhereClause: &where})
	require.NotNil(t, actual)
	require.Equal(t, where, *actual)
}

func Test_toDynamoDbSourceTableOptionMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []*mgmtv1alpha1.DynamoDBSourceTableOption
		expected map[string]*mgmtv1alpha1.DynamoDBSourceTableOption
	}{
		{
			name: "Single Option",
			input: []*mgmtv1alpha1.DynamoDBSourceTableOption{
				{
					Table: "Table1",
				},
			},
			expected: map[string]*mgmtv1alpha1.DynamoDBSourceTableOption{
				"Table1": {
					Table: "Table1",
				},
			},
		},
		{
			name: "Multiple Options",
			input: []*mgmtv1alpha1.DynamoDBSourceTableOption{
				{
					Table: "Table1",
				},
				{
					Table: "Table2",
				},
			},
			expected: map[string]*mgmtv1alpha1.DynamoDBSourceTableOption{
				"Table1": {
					Table: "Table1",
				},
				"Table2": {
					Table: "Table2",
				},
			},
		},
		{
			name:     "Empty Input",
			input:    []*mgmtv1alpha1.DynamoDBSourceTableOption{},
			expected: map[string]*mgmtv1alpha1.DynamoDBSourceTableOption{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toDynamoDbSourceTableOptionMap(tt.input)
			require.True(t, reflect.DeepEqual(result, tt.expected), "expected %v, got %v", tt.expected, result)
		})
	}
}
