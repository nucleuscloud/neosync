package connectiondata

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	aws_manager "github.com/nucleuscloud/neosync/internal/aws"
	neosync_dynamodb "github.com/nucleuscloud/neosync/internal/dynamodb"
)

type AwsDynamodbConnectionDataService struct {
	logger     *slog.Logger
	awsmanager aws_manager.NeosyncAwsManagerClient
	connection *mgmtv1alpha1.Connection
	connconfig *mgmtv1alpha1.DynamoDBConnectionConfig
}

func NewAwsDynamodbConnectionDataService(
	logger *slog.Logger,
	awsmanager aws_manager.NeosyncAwsManagerClient,
	connection *mgmtv1alpha1.Connection,
) *AwsDynamodbConnectionDataService {
	return &AwsDynamodbConnectionDataService{
		logger:     logger,
		awsmanager: awsmanager,
		connection: connection,
		connconfig: connection.GetConnectionConfig().GetDynamodbConfig(),
	}
}

func (s *AwsDynamodbConnectionDataService) StreamData(
	ctx context.Context,
	stream *connect.ServerStream[mgmtv1alpha1.GetConnectionDataStreamResponse],
	config *mgmtv1alpha1.ConnectionStreamConfig,
	schema, table string,
) error {
	dynamoclient, err := s.awsmanager.NewDynamoDbClient(ctx, s.connconfig)
	if err != nil {
		return fmt.Errorf("unable to create dynamodb client from connection: %w", err)
	}
	var lastEvaluatedKey map[string]dynamotypes.AttributeValue

	for {
		output, err := dynamoclient.ScanTable(ctx, table, lastEvaluatedKey)
		if err != nil {
			return fmt.Errorf("failed to scan table %s: %w", table, err)
		}

		for _, item := range output.Items {
			itemBits, err := neosync_dynamodb.ConvertDynamoItemToGoMap(item)
			if err != nil {
				return err
			}

			var itemBytes bytes.Buffer
			enc := gob.NewEncoder(&itemBytes)
			if err := enc.Encode(itemBits); err != nil {
				return fmt.Errorf("unable to encode dynamodb item using gob: %w", err)
			}
			if err := stream.Send(&mgmtv1alpha1.GetConnectionDataStreamResponse{RowBytes: itemBytes.Bytes()}); err != nil {
				return fmt.Errorf("failed to send stream response: %w", err)
			}
		}

		lastEvaluatedKey = output.LastEvaluatedKey
		if lastEvaluatedKey == nil {
			break
		}
	}
	return nil
}

func (s *AwsDynamodbConnectionDataService) GetSchema(
	ctx context.Context,
	config *mgmtv1alpha1.ConnectionSchemaConfig,
) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	dynclient, err := s.awsmanager.NewDynamoDbClient(ctx, s.connconfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create dynamodb client from connection: %w", err)
	}
	tableNames, err := dynclient.ListAllTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve dynamodb tables: %w", err)
	}
	schemas := []*mgmtv1alpha1.DatabaseColumn{}
	for _, tableName := range tableNames {
		schemas = append(schemas, &mgmtv1alpha1.DatabaseColumn{
			Schema: "dynamodb",
			Table:  tableName,
		})
	}
	return schemas, nil
}

func (s *AwsDynamodbConnectionDataService) GetInitStatements(
	ctx context.Context,
	options *mgmtv1alpha1.InitStatementOptions,
) (*mgmtv1alpha1.GetConnectionInitStatementsResponse, error) {
	return nil, errors.ErrUnsupported
}

func (s *AwsDynamodbConnectionDataService) GetTableConstraints(
	ctx context.Context,
) (*mgmtv1alpha1.GetConnectionTableConstraintsResponse, error) {
	return nil, errors.ErrUnsupported
}

func (s *AwsDynamodbConnectionDataService) GetTableSchema(ctx context.Context, schema, table string) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	return nil, errors.ErrUnsupported
}

func (s *AwsDynamodbConnectionDataService) GetTableRowCount(ctx context.Context, schema, table string, whereClause *string) (int64, error) {
	return 0, errors.ErrUnsupported
}
