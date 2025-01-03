package testcontainers_dynamodb

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dyntypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/docker/go-connections/nat"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/sync/errgroup"
)

type DynamoDBTestSyncContainer struct {
	Source *DynamoDBTestContainer
	Target *DynamoDBTestContainer
}

func NewDynamoDBTestSyncContainer(ctx context.Context, t *testing.T, sourceOpts, destOpts []Option) (*DynamoDBTestSyncContainer, error) {
	tc := &DynamoDBTestSyncContainer{}
	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		d, err := NewDynamoDBTestContainer(ctx, t, sourceOpts...)
		if err != nil {
			return err
		}
		tc.Source = d
		return nil
	})

	errgrp.Go(func() error {
		d, err := NewDynamoDBTestContainer(ctx, t, destOpts...)
		if err != nil {
			return err
		}
		tc.Target = d
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}

	return tc, nil
}

func (d *DynamoDBTestSyncContainer) TearDown(ctx context.Context) error {
	if d.Source != nil {
		if d.Source.TestContainer != nil {
			err := d.Source.TestContainer.Terminate(ctx)
			if err != nil {
				return err
			}
		}
	}
	if d.Target != nil {
		if d.Target.TestContainer != nil {
			err := d.Target.TestContainer.Terminate(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Holds the DynamoDB test container and client
type DynamoDBTestContainer struct {
	Client        *dynamodb.Client
	URL           string
	TestContainer testcontainers.Container
	Credentials   *mgmtv1alpha1.AwsS3Credentials
	awsId         string
	awsSecret     string
	awsToken      string
}

// Option is a functional option for configuring the DynamoDB Test Container
type Option func(*DynamoDBTestContainer)

// WithAwsId sets the AWS access key ID
func WithAwsId(id string) Option {
	return func(d *DynamoDBTestContainer) {
		d.awsId = id
	}
}

// WithAwsSecret sets the AWS secret access key
func WithAwsSecret(secret string) Option {
	return func(d *DynamoDBTestContainer) {
		d.awsSecret = secret
	}
}

// WithAwsToken sets the AWS session token
func WithAwsToken(token string) Option {
	return func(d *DynamoDBTestContainer) {
		d.awsToken = token
	}
}

// NewDynamoDBTestContainer initializes a new DynamoDB Test Container with functional options
func NewDynamoDBTestContainer(ctx context.Context, t *testing.T, opts ...Option) (*DynamoDBTestContainer, error) {
	d := &DynamoDBTestContainer{
		awsId:     "fakeid",     // default value
		awsSecret: "fakesecret", // default value
		awsToken:  "faketoken",  // default value
	}
	for _, opt := range opts {
		opt(d)
	}
	return d.Setup(ctx, t)
}

// Creates and starts a DynamoDB test container
func (d *DynamoDBTestContainer) Setup(ctx context.Context, t *testing.T) (*DynamoDBTestContainer, error) {
	port := nat.Port("8000/tcp")
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "amazon/dynamodb-local:2.5.2",
			ExposedPorts: []string{string(port)},
			WaitingFor:   wait.ForListeningPort(port),
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}

	mappedport, err := container.MappedPort(ctx, port)
	if err != nil {
		return nil, err
	}
	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("http://%s:%d", host, mappedport.Int())

	awscfg, err := awsmanager.GetAwsConfig(ctx, &awsmanager.AwsCredentialsConfig{
		Endpoint: endpoint,
		Id:       d.awsId,
		Secret:   d.awsSecret,
		Token:    d.awsToken,
	})
	if err != nil {
		return nil, err
	}

	dtoAwsCreds := &mgmtv1alpha1.AwsS3Credentials{
		AccessKeyId:     &d.awsId,
		SecretAccessKey: &d.awsSecret,
		SessionToken:    &d.awsToken,
	}

	client := dynamodb.NewFromConfig(*awscfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = &endpoint
	})

	return &DynamoDBTestContainer{
		Client:        client,
		URL:           endpoint,
		TestContainer: container,
		Credentials:   dtoAwsCreds,
		awsId:         d.awsId,
		awsSecret:     d.awsSecret,
		awsToken:      d.awsToken,
	}, nil
}

// Terminates the container
func (d *DynamoDBTestContainer) TearDown(ctx context.Context) error {
	if d.TestContainer != nil {
		return d.TestContainer.Terminate(ctx)
	}
	return nil
}

func (d *DynamoDBTestContainer) SetupDynamoDbTable(ctx context.Context, tableName, primaryKey string) error {
	out, err := d.Client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:            &tableName,
		KeySchema:            []dyntypes.KeySchemaElement{{KeyType: dyntypes.KeyTypeHash, AttributeName: &primaryKey}},
		AttributeDefinitions: []dyntypes.AttributeDefinition{{AttributeName: &primaryKey, AttributeType: dyntypes.ScalarAttributeTypeS}},
		BillingMode:          dyntypes.BillingModePayPerRequest,
	})
	if err != nil {
		return err
	}
	if out.TableDescription.TableStatus == dyntypes.TableStatusActive {
		return nil
	}
	if out.TableDescription.TableStatus == dyntypes.TableStatusCreating {
		return d.waitUntilDynamoTableExists(ctx, tableName)
	}
	return fmt.Errorf("%s dynamo table created but unexpected table status: %s", tableName, out.TableDescription.TableStatus)
}

func (d *DynamoDBTestContainer) waitUntilDynamoTableExists(ctx context.Context, tableName string) error {
	input := &dynamodb.DescribeTableInput{TableName: &tableName}
	for {
		out, err := d.Client.DescribeTable(ctx, input)
		if err != nil && !awsmanager.IsNotFound(err) {
			return err
		}
		if err != nil && awsmanager.IsNotFound(err) {
			continue
		}
		if out.Table.TableStatus == dyntypes.TableStatusActive {
			return nil
		}
	}
}

func (d *DynamoDBTestContainer) DestroyDynamoDbTable(ctx context.Context, tableName string) error {
	_, err := d.Client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: &tableName,
	})
	if err != nil {
		return err
	}
	return d.waitUntilDynamoTableDestroy(ctx, tableName)
}

func (d *DynamoDBTestContainer) waitUntilDynamoTableDestroy(ctx context.Context, tableName string) error {
	input := &dynamodb.DescribeTableInput{TableName: &tableName}
	for {
		_, err := d.Client.DescribeTable(ctx, input)
		if err != nil && !awsmanager.IsNotFound(err) {
			return err
		}
		if err != nil && awsmanager.IsNotFound(err) {
			return nil
		}
	}
}

func (d *DynamoDBTestContainer) InsertDynamoDBRecords(ctx context.Context, tableName string, data []map[string]dyntypes.AttributeValue) error {
	writeRequests := make([]dyntypes.WriteRequest, len(data))
	for i, record := range data {
		writeRequests[i] = dyntypes.WriteRequest{
			PutRequest: &dyntypes.PutRequest{
				Item: record,
			},
		}
	}

	_, err := d.Client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]dyntypes.WriteRequest{
			tableName: writeRequests,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
