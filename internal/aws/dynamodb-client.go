package awsmanager

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDbClient struct {
	client dynamoDBAPIV2
}

type dynamoDBAPIV2 interface {
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	ListTables(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error)
}

func NewDynamoDbClient(api dynamoDBAPIV2) *DynamoDbClient {
	return &DynamoDbClient{client: api}
}

func (d *DynamoDbClient) ListAllTables(ctx context.Context, input *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) ([]string, error) {
	tableNames := []string{}
	done := false
	for !done {
		output, err := d.client.ListTables(ctx, input, optFns...)
		if err != nil {
			return nil, err
		}
		fmt.Println("hit table output")
		tableNames = append(tableNames, output.TableNames...)
		input.ExclusiveStartTableName = output.LastEvaluatedTableName
		done = output.LastEvaluatedTableName == nil
	}
	return tableNames, nil
}
