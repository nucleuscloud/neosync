package awsmanager

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDbClient struct {
	client dynamoDBAPIV2
}

type dynamoDBAPIV2 interface {
	DescribeTable(
		ctx context.Context,
		params *dynamodb.DescribeTableInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.DescribeTableOutput, error)
	Scan(
		ctx context.Context,
		params *dynamodb.ScanInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.ScanOutput, error)
	ListTables(
		ctx context.Context,
		params *dynamodb.ListTablesInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.ListTablesOutput, error)
}

func NewDynamoDbClient(api dynamoDBAPIV2) *DynamoDbClient {
	return &DynamoDbClient{client: api}
}

func (d *DynamoDbClient) ListAllTables(
	ctx context.Context,
	input *dynamodb.ListTablesInput,
	optFns ...func(*dynamodb.Options),
) ([]string, error) {
	tableNames := []string{}
	done := false
	for !done {
		output, err := d.client.ListTables(ctx, input, optFns...)
		if err != nil {
			return nil, err
		}
		tableNames = append(tableNames, output.TableNames...)
		input.ExclusiveStartTableName = output.LastEvaluatedTableName
		done = output.LastEvaluatedTableName == nil
	}
	return tableNames, nil
}

type DynamoDbTableKey struct {
	HashKey  string
	RangeKey string
}

func (d *DynamoDbClient) GetTableKey(
	ctx context.Context,
	tableName string,
) (*DynamoDbTableKey, error) {
	describeTableOutput, err := d.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe table %s: %w", tableName, err)
	}

	if describeTableOutput == nil || describeTableOutput.Table == nil {
		return nil, fmt.Errorf("invalid DescribeTableOutput for table %s", tableName)
	}

	dynamoKey := &DynamoDbTableKey{}
	for _, key := range describeTableOutput.Table.KeySchema {
		if key.KeyType == types.KeyTypeHash {
			dynamoKey.HashKey = *key.AttributeName
		}
		if key.KeyType == types.KeyTypeRange {
			dynamoKey.RangeKey = *key.AttributeName
		}
	}

	return dynamoKey, nil
}

func (d *DynamoDbClient) ScanTable(
	ctx context.Context,
	tableName string,
	lastEvaluatedKey map[string]types.AttributeValue,
) (*dynamodb.ScanOutput, error) {
	input := &dynamodb.ScanInput{
		TableName:         &tableName,
		ExclusiveStartKey: lastEvaluatedKey,
	}

	output, err := d.client.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan table %s: %w", tableName, err)
	}

	return output, nil
}
