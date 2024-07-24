package awsmanager

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_NewDynamoDbClient(t *testing.T) {
	mockApi := NewMockdynamoDBAPIV2(t)
	client := NewDynamoDbClient(mockApi)
	require.NotNil(t, client)
}

func Test_DynamoDbClient_ListAllTables(t *testing.T) {
	mockApi := NewMockdynamoDBAPIV2(t)
	client := NewDynamoDbClient(mockApi)

	mockApi.On("ListTables", mock.Anything, mock.Anything, mock.Anything).
		Return(&dynamodb.ListTablesOutput{TableNames: []string{"foo"}, LastEvaluatedTableName: aws.String("foo")}, nil).Once()
	mockApi.On("ListTables", mock.Anything, mock.Anything, mock.Anything).
		Return(&dynamodb.ListTablesOutput{TableNames: []string{"bar"}}, nil).Once()

	tableNames, err := client.ListAllTables(context.Background(), &dynamodb.ListTablesInput{})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"foo", "bar"}, tableNames)
}
