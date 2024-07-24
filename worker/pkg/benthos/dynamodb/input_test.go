package neosync_benthos_dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/require"
)

func Test_isTableActive(t *testing.T) {
	type testcase struct {
		input    *dynamodb.DescribeTableOutput
		expected bool
	}

	testcases := []testcase{
		{nil, false},
		{&dynamodb.DescribeTableOutput{}, false},
		{&dynamodb.DescribeTableOutput{Table: nil}, false},
		{&dynamodb.DescribeTableOutput{Table: &types.TableDescription{}}, false},
		{&dynamodb.DescribeTableOutput{Table: &types.TableDescription{TableStatus: types.TableStatusArchived}}, false},
		{&dynamodb.DescribeTableOutput{Table: &types.TableDescription{TableStatus: types.TableStatusActive}}, true},
	}

	for _, tc := range testcases {
		actual := isTableActive(tc.input)
		require.Equal(t, tc.expected, actual)
	}
}
