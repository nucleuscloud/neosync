package datasync_workflow

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dyntypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/docker/go-connections/nat"
	_ "github.com/go-sql-driver/mysql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcmysql "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/mysql"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	testmssql "github.com/testcontainers/testcontainers-go/modules/mssql"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
)

type mssqlTest struct {
	pool          *sql.DB
	testcontainer *testmssql.MSSQLServerContainer
	source        *mssqlTestContainer
	target        *mssqlTestContainer
}

type mssqlTestContainer struct {
	pool *sql.DB
	url  string
}

type redisTest struct {
	url           string
	testcontainer *redis.RedisContainer
}

type mongodbTestContainer struct {
	testcontainer *testmongodb.MongoDBContainer
	client        *mongo.Client
	url           string
}

type mongodbTest struct {
	source *mongodbTestContainer
	target *mongodbTestContainer
}
type IntegrationTestSuite struct {
	suite.Suite

	ctx context.Context

	mysql    *tcmysql.MysqlTestSyncContainer
	postgres *tcpostgres.PostgresTestSyncContainer
	mssql    *mssqlTest
	redis    *redisTest
	dynamo   *dynamodbTest
	mongodb  *mongodbTest
}

func (s *IntegrationTestSuite) SetupMongoDb() (*mongodbTest, error) {
	var source *mongodbTestContainer
	var target *mongodbTestContainer

	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		sourcecontainer, err := createMongoTestContainer(s.ctx)
		if err != nil {
			return err
		}
		source = sourcecontainer
		return nil
	})

	errgrp.Go(func() error {
		targetcontainer, err := createMongoTestContainer(s.ctx)
		if err != nil {
			return err
		}
		target = targetcontainer
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}

	return &mongodbTest{
		source: source,
		target: target,
	}, nil
}

func createMongoTestContainer(
	ctx context.Context,
) (*mongodbTestContainer, error) {
	mongodbContainer, err := testmongodb.Run(ctx, "mongo:6")
	if err != nil {
		return nil, err
	}
	uri, err := mongodbContainer.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	return &mongodbTestContainer{
		testcontainer: mongodbContainer,
		client:        client,
		url:           uri,
	}, nil
}
func (s *IntegrationTestSuite) SetupMssql() (*mssqlTest, error) {
	mssqlcontainer, err := testmssql.Run(s.ctx,
		"mcr.microsoft.com/mssql/server:2022-latest",
		testmssql.WithAcceptEULA(),
		testmssql.WithPassword("mssqlPASSword1"),
	)
	if err != nil {
		return nil, err
	}
	connstr, err := mssqlcontainer.ConnectionString(s.ctx, "encrypt=disable")
	if err != nil {
		return nil, err
	}
	conn, err := sql.Open(sqlmanager_shared.MssqlDriver, connstr)
	if err != nil {
		return nil, err
	}

	source, err := createMssqlTest(s.ctx, mssqlcontainer, conn, "datasync_source")
	if err != nil {
		return nil, err
	}
	target, err := createMssqlTest(s.ctx, mssqlcontainer, conn, "datasync_target")
	if err != nil {
		return nil, err
	}

	return &mssqlTest{
		testcontainer: mssqlcontainer,
		pool:          conn,
		source:        source,
		target:        target,
	}, nil
}

func createMssqlTest(ctx context.Context, mssqlcontainer *testmssql.MSSQLServerContainer, conn *sql.DB, database string) (*mssqlTestContainer, error) {
	_, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s;", database))
	if err != nil {
		return nil, err
	}

	connStr, err := mssqlcontainer.ConnectionString(ctx, fmt.Sprintf("database=%s", database), "encrypt=disable")
	if err != nil {
		return nil, err
	}

	dbConn, err := sql.Open(sqlmanager_shared.MssqlDriver, connStr)
	if err != nil {
		return nil, err
	}

	return &mssqlTestContainer{
		pool: dbConn,
		url:  connStr,
	}, nil
}

func (s *IntegrationTestSuite) SetupPostgres() (*tcpostgres.PostgresTestSyncContainer, error) {
	container, err := tcpostgres.NewPostgresTestSyncContainer(s.ctx, []tcpostgres.Option{}, []tcpostgres.Option{})
	if err != nil {
		return nil, err
	}
	return container, nil
}

func (s *IntegrationTestSuite) SetupMysql() (*tcmysql.MysqlTestSyncContainer, error) {
	container, err := tcmysql.NewMysqlTestSyncContainer(s.ctx, []tcmysql.Option{}, []tcmysql.Option{})
	if err != nil {
		return nil, err
	}
	return container, nil
}

func (s *IntegrationTestSuite) SetupRedis() (*redisTest, error) {
	redisContainer, err := redis.Run(
		s.ctx,
		"docker.io/redis:7",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
		testcontainers.WithWaitStrategy(
			wait.ForLog("* Ready to accept connections"),
			wait.ForExposedPort(),
		),
	)
	if err != nil {
		return nil, err
	}
	redisUrl, err := redisContainer.ConnectionString(s.ctx)
	if err != nil {
		return nil, err
	}
	return &redisTest{
		testcontainer: redisContainer,
		url:           redisUrl,
	}, nil
}

type dynamodbTest struct {
	container testcontainers.Container
	endpoint  string

	// Used by plugging in to Neosync resources so Benthos can wire up its aws config
	dtoAwsCreds *mgmtv1alpha1.AwsS3Credentials

	dynamoclient *dynamodb.Client
}

func (s *IntegrationTestSuite) SetupDynamoDB() (*dynamodbTest, error) {
	port := nat.Port("8000/tcp")
	container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
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

	mappedport, err := container.MappedPort(s.ctx, port)
	if err != nil {
		return nil, err
	}
	host, err := container.Host(s.ctx)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("http://%s:%d", host, mappedport.Int())
	fakeId := "fakeid"
	fakeSecret := "fakesecret"
	fakeToken := "faketoken"

	awscfg, err := awsmanager.GetAwsConfig(s.ctx, &awsmanager.AwsCredentialsConfig{
		Endpoint: endpoint,
		Id:       fakeId,
		Secret:   fakeSecret,
		Token:    fakeToken,
	})
	if err != nil {
		return nil, err
	}

	dtoAwsCreds := &mgmtv1alpha1.AwsS3Credentials{
		AccessKeyId:     &fakeId,
		SecretAccessKey: &fakeSecret,
		SessionToken:    &fakeToken,
	}

	return &dynamodbTest{
		container:   container,
		endpoint:    endpoint,
		dtoAwsCreds: dtoAwsCreds,
		dynamoclient: dynamodb.NewFromConfig(*awscfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = &endpoint
		}),
	}, nil
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	var postgresTest *tcpostgres.PostgresTestSyncContainer
	var mysqlTest *tcmysql.MysqlTestSyncContainer
	var mssqlTest *mssqlTest
	var redisTest *redisTest
	var dynamoTest *dynamodbTest
	var mongodbTest *mongodbTest

	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		p, err := s.SetupPostgres()
		if err != nil {
			return err
		}
		postgresTest = p
		return nil
	})

	errgrp.Go(func() error {
		m, err := s.SetupMysql()
		if err != nil {
			return err
		}
		mysqlTest = m
		return nil
	})

	errgrp.Go(func() error {
		m, err := s.SetupMssql()
		if err != nil {
			return err
		}
		mssqlTest = m
		return nil
	})

	errgrp.Go(func() error {
		r, err := s.SetupRedis()
		if err != nil {
			return err
		}
		redisTest = r
		return nil
	})

	errgrp.Go(func() error {
		d, err := s.SetupDynamoDB()
		if err != nil {
			return err
		}
		dynamoTest = d
		return nil
	})

	errgrp.Go(func() error {
		m, err := s.SetupMongoDb()
		if err != nil {
			return err
		}
		mongodbTest = m
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		panic(err)
	}

	s.postgres = postgresTest
	s.mysql = mysqlTest
	s.mssql = mssqlTest
	s.redis = redisTest
	s.dynamo = dynamoTest
	s.mongodb = mongodbTest
}

func (s *IntegrationTestSuite) RunMysqlSqlFiles(pool *sql.DB, testFolder string, files []string) {
	s.T().Logf("running mysql sql file. folder: %s \n", testFolder)
	for _, file := range files {
		sqlStr, err := os.ReadFile(fmt.Sprintf("./testdata/%s/%s", testFolder, file))
		if err != nil {
			panic(err)
		}

		_, err = pool.ExecContext(s.ctx, string(sqlStr))
		if err != nil {
			panic(err)
		}
	}
}

func (s *IntegrationTestSuite) SetupDynamoDbTable(ctx context.Context, tableName, primaryKey string) error {
	s.T().Logf("Creating DynamoDB table: %s\n", tableName)

	out, err := s.dynamo.dynamoclient.CreateTable(ctx, &dynamodb.CreateTableInput{
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
		return s.waitUntilDynamoTableExists(ctx, tableName)
	}
	return fmt.Errorf("%s dynamo table created but unexpected table status: %s", tableName, out.TableDescription.TableStatus)
}

func (s *IntegrationTestSuite) waitUntilDynamoTableExists(ctx context.Context, tableName string) error {
	input := &dynamodb.DescribeTableInput{TableName: &tableName}
	for {
		out, err := s.dynamo.dynamoclient.DescribeTable(ctx, input)
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

func (s *IntegrationTestSuite) DestroyDynamoDbTable(ctx context.Context, tableName string) error {
	s.T().Logf("Destroying DynamoDB table: %s\n", tableName)

	_, err := s.dynamo.dynamoclient.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: &tableName,
	})
	if err != nil {
		return err
	}
	return s.waitUntilDynamoTableDestroy(ctx, tableName)
}

func (s *IntegrationTestSuite) waitUntilDynamoTableDestroy(ctx context.Context, tableName string) error {
	input := &dynamodb.DescribeTableInput{TableName: &tableName}
	for {
		_, err := s.dynamo.dynamoclient.DescribeTable(ctx, input)
		if err != nil && !awsmanager.IsNotFound(err) {
			return err
		}
		if err != nil && awsmanager.IsNotFound(err) {
			return nil
		}
	}
}

func (s *IntegrationTestSuite) InsertDynamoDBRecords(tableName string, data []map[string]dyntypes.AttributeValue) error {
	s.T().Logf("Inserting %d DynamoDB Records into table: %s\n", len(data), tableName)

	writeRequests := make([]dyntypes.WriteRequest, len(data))
	for i, record := range data {
		writeRequests[i] = dyntypes.WriteRequest{
			PutRequest: &dyntypes.PutRequest{
				Item: record,
			},
		}
	}

	_, err := s.dynamo.dynamoclient.BatchWriteItem(s.ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]dyntypes.WriteRequest{
			tableName: writeRequests,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *IntegrationTestSuite) InsertMongoDbRecords(client *mongo.Client, database, collection string, documents []any) (int, error) {
	db := client.Database(database)
	col := db.Collection(collection)

	result, err := col.InsertMany(s.ctx, documents)
	if err != nil {
		return 0, fmt.Errorf("failed to insert mongodb records: %v", err)
	}

	return len(result.InsertedIDs), nil
}

func (s *IntegrationTestSuite) DropMongoDbCollection(ctx context.Context, client *mongo.Client, database, collection string) error {
	db := client.Database(database)
	collections, err := db.ListCollectionNames(ctx, map[string]any{"name": collection})
	if err != nil {
		return err
	}
	if len(collections) == 0 {
		return nil
	}
	return db.Collection(collection).Drop(ctx)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	// postgres
	if s.postgres != nil {
		err := s.postgres.TearDown(s.ctx)
		if err != nil {
			panic(err)
		}
	}

	// mssql
	if s.mssql != nil {
		if s.mssql.source.pool != nil {
			s.mssql.source.pool.Close()
		}
		if s.mssql.target.pool != nil {
			s.mssql.target.pool.Close()
		}
		if s.mssql.pool != nil {
			s.mssql.pool.Close()
		}
		if s.mssql.testcontainer != nil {
			err := s.mssql.testcontainer.Terminate(s.ctx)
			if err != nil {
				panic(err)
			}
		}
	}

	// mysql
	if s.mysql != nil {
		err := s.mysql.TearDown(s.ctx)
		if err != nil {
			panic(err)
		}
	}

	// redis
	if s.redis != nil {
		if s.redis.testcontainer != nil {
			if err := s.redis.testcontainer.Terminate(s.ctx); err != nil {
				panic(err)
			}
		}
	}

	// localstack
	if s.dynamo != nil {
		if s.dynamo.container != nil {
			if err := s.dynamo.container.Terminate(s.ctx); err != nil {
				panic(err)
			}
		}
	}

	// mongodb
	if s.mongodb != nil {
		if s.mongodb.source.client != nil {
			if err := s.mongodb.source.client.Disconnect(s.ctx); err != nil {
				panic(err)
			}
		}
		if s.mongodb.source.testcontainer != nil {
			if err := s.mongodb.source.testcontainer.Terminate(s.ctx); err != nil {
				panic(err)
			}
		}
		if s.mongodb.target.client != nil {
			if err := s.mongodb.target.client.Disconnect(s.ctx); err != nil {
				panic(err)
			}
		}
		if s.mongodb.target.testcontainer != nil {
			if err := s.mongodb.target.testcontainer.Terminate(s.ctx); err != nil {
				panic(err)
			}
		}
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	suite.Run(t, new(IntegrationTestSuite))
}
