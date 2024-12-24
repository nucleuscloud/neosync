package datasync_workflow

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dyntypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/mongoprovider"
	"github.com/nucleuscloud/neosync/internal/connection-manager/providers/sqlprovider"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	"github.com/nucleuscloud/neosync/internal/testutil"
	neosync_redis "github.com/nucleuscloud/neosync/worker/internal/redis"
	accountstatus_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/account-status"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	jobhooks_by_timing_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/jobhooks-by-timing"
	posttablesync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/post-table-sync"
	runsqlinittablestmts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/run-sql-init-table-stmts"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
	mssql_datatypes "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/mssql/data-types"
	mssql_simple "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/mssql/simple"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/errgroup"

	"connectrpc.com/connect"
	// tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/testsuite"
)

const neosyncDbMigrationsPath = "../../../../../backend/sql/postgresql/schema"

func getAllMssqlSyncTests() map[string][]*workflow_testdata.IntegrationTest {
	allTests := map[string][]*workflow_testdata.IntegrationTest{}
	simpleTests := mssql_simple.GetSyncTests()
	allDatatypesTests := mssql_datatypes.GetSyncTests()

	allTests["Simple"] = simpleTests
	allTests["DataTypes"] = allDatatypesTests
	return allTests
}

func (s *IntegrationTestSuite) Test_Workflow_Sync_Mssql() {
	tests := getAllMssqlSyncTests()
	for groupName, group := range tests {
		group := group
		s.T().Run(groupName, func(t *testing.T) {
			t.Parallel()
			for _, tt := range group {
				t.Run(tt.Name, func(t *testing.T) {
					t.Logf("running integration test: %s \n", tt.Name)
					// setup
					s.RunMysqlSqlFiles(s.mssql.source.pool, tt.Folder, tt.SourceFilePaths)
					s.RunMysqlSqlFiles(s.mssql.target.pool, tt.Folder, tt.TargetFilePaths)

					schemas := []*mgmtv1alpha1.MssqlSourceSchemaOption{}
					subsetMap := map[string]*mgmtv1alpha1.MssqlSourceSchemaOption{}
					for table, where := range tt.SubsetMap {
						schema, table := sqlmanager_shared.SplitTableKey(table)
						if _, exists := subsetMap[schema]; !exists {
							subsetMap[schema] = &mgmtv1alpha1.MssqlSourceSchemaOption{
								Schema: schema,
								Tables: []*mgmtv1alpha1.MssqlSourceTableOption{},
							}
						}
						w := where
						subsetMap[schema].Tables = append(subsetMap[schema].Tables, &mgmtv1alpha1.MssqlSourceTableOption{
							Table:       table,
							WhereClause: &w,
						})
					}

					for _, s := range subsetMap {
						schemas = append(schemas, s)
					}

					var subsetByForeignKeyConstraints bool
					destinationOptions := &mgmtv1alpha1.JobDestinationOptions{
						Config: &mgmtv1alpha1.JobDestinationOptions_MssqlOptions{
							MssqlOptions: &mgmtv1alpha1.MssqlDestinationConnectionOptions{},
						},
					}
					if tt.JobOptions != nil {
						if tt.JobOptions.SubsetByForeignKeyConstraints {
							subsetByForeignKeyConstraints = true
						}
						destinationOptions = &mgmtv1alpha1.JobDestinationOptions{
							Config: &mgmtv1alpha1.JobDestinationOptions_MssqlOptions{
								MssqlOptions: &mgmtv1alpha1.MssqlDestinationConnectionOptions{
									InitTableSchema: tt.JobOptions.InitSchema,
									TruncateTable: &mgmtv1alpha1.MssqlTruncateTableConfig{
										TruncateBeforeInsert: tt.JobOptions.Truncate,
									},
									SkipForeignKeyViolations: tt.JobOptions.SkipForeignKeyViolations,
								},
							},
						}
					}

					mux := http.NewServeMux()
					mux.Handle(mgmtv1alpha1connect.UserAccountServiceIsAccountStatusValidProcedure, connect.NewUnaryHandler(
						mgmtv1alpha1connect.UserAccountServiceIsAccountStatusValidProcedure,
						func(ctx context.Context, r *connect.Request[mgmtv1alpha1.IsAccountStatusValidRequest]) (*connect.Response[mgmtv1alpha1.IsAccountStatusValidResponse], error) {
							return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{IsValid: true}), nil
						},
					))
					mux.Handle(mgmtv1alpha1connect.JobServiceGetJobProcedure, connect.NewUnaryHandler(
						mgmtv1alpha1connect.JobServiceGetJobProcedure,
						func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetJobRequest]) (*connect.Response[mgmtv1alpha1.GetJobResponse], error) {
							return connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
								Job: &mgmtv1alpha1.Job{
									Id:        "115aaf2c-776e-4847-8268-d914e3c15968",
									AccountId: "225aaf2c-776e-4847-8268-d914e3c15988",
									Source: &mgmtv1alpha1.JobSource{
										Options: &mgmtv1alpha1.JobSourceOptions{
											Config: &mgmtv1alpha1.JobSourceOptions_Mssql{
												Mssql: &mgmtv1alpha1.MssqlSourceConnectionOptions{
													ConnectionId:                  "c9b6ce58-5c8e-4dce-870d-96841b19d988",
													Schemas:                       schemas,
													SubsetByForeignKeyConstraints: subsetByForeignKeyConstraints,
												},
											},
										},
									},
									Destinations: []*mgmtv1alpha1.JobDestination{
										{
											ConnectionId: "226add85-5751-4232-b085-a0ae93afc7ce",
											Options:      destinationOptions,
										},
									},
									Mappings:           tt.JobMappings,
									VirtualForeignKeys: tt.VirtualForeignKeys,
								}}), nil
						},
					))

					mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
						mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
						func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
							if r.Msg.GetId() == "c9b6ce58-5c8e-4dce-870d-96841b19d988" {
								return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
									Connection: &mgmtv1alpha1.Connection{
										Id:   "c9b6ce58-5c8e-4dce-870d-96841b19d988",
										Name: "source",
										ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
											Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{
												MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
													ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
														Url: s.mssql.source.url,
													},
												},
											},
										},
									},
								}), nil
							}
							if r.Msg.GetId() == "226add85-5751-4232-b085-a0ae93afc7ce" {
								return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
									Connection: &mgmtv1alpha1.Connection{
										Id:   "226add85-5751-4232-b085-a0ae93afc7ce",
										Name: "target",
										ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
											Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{
												MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
													ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
														Url: s.mssql.target.url,
													},
												},
											},
										},
									},
								}), nil
							}
							return nil, nil
						},
					))

					addRunContextProcedureMux(mux)
					addEmptyJobHooksProcedureMux(mux)
					srv := startHTTPServer(t, mux)
					env := executeWorkflow(t, srv, nil, "115aaf2c-776e-4847-8268-d914e3c15968", tt.ExpectError)
					require.Truef(t, env.IsWorkflowCompleted(), fmt.Sprintf("Workflow did not complete. Test: %s", tt.Name))
					err := env.GetWorkflowError()
					if tt.ExpectError {
						require.Error(t, err, "Did not received Temporal Workflow Error %s", tt.Name)
						return
					}
					require.NoError(t, err, "Received Temporal Workflow Error %s", tt.Name)

					for table, expected := range tt.Expected {
						rows, err := s.mssql.target.pool.QueryContext(s.ctx, fmt.Sprintf("select * from %s;", table))
						require.NoError(t, err)
						count := 0
						for rows.Next() {
							count++
						}
						assert.Equalf(t, expected.RowCount, count, fmt.Sprintf("Test: %s Table: %s", tt.Name, table))
					}

					// tear down
					s.RunMysqlSqlFiles(s.mssql.source.pool, tt.Folder, []string{"teardown.sql"})
					s.RunMysqlSqlFiles(s.mssql.target.pool, tt.Folder, []string{"teardown.sql"})
				})
			}
		})
	}
}

// Used if there is no plan for jobhooks in the test and just need to satisfy the impl
func addEmptyJobHooksProcedureMux(mux *http.ServeMux) {
	mux.Handle(mgmtv1alpha1connect.JobServiceGetActiveJobHooksByTimingProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetActiveJobHooksByTimingProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetActiveJobHooksByTimingRequest]) (*connect.Response[mgmtv1alpha1.GetActiveJobHooksByTimingResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.GetActiveJobHooksByTimingResponse{Hooks: []*mgmtv1alpha1.JobHook{}}), nil
		},
	))
}

func addRunContextProcedureMux(mux *http.ServeMux) {
	rcmap := map[string][]byte{}
	rcmu := sync.RWMutex{}
	mux.Handle(mgmtv1alpha1connect.JobServiceGetRunContextProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetRunContextProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetRunContextRequest]) (*connect.Response[mgmtv1alpha1.GetRunContextResponse], error) {
			rcmu.RLock()
			defer rcmu.RUnlock()
			val, ok := rcmap[toRunContextKeyString(r.Msg.GetId())]
			if !ok {
				return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("unable to find key: %s", toRunContextKeyString(r.Msg.GetId())))
			}
			return connect.NewResponse(&mgmtv1alpha1.GetRunContextResponse{Value: val}), nil
		},
	))

	mux.Handle(mgmtv1alpha1connect.JobServiceSetRunContextsProcedure, connect.NewClientStreamHandler(
		mgmtv1alpha1connect.JobServiceSetRunContextsProcedure,
		func(ctx context.Context, cs *connect.ClientStream[mgmtv1alpha1.SetRunContextsRequest]) (*connect.Response[mgmtv1alpha1.SetRunContextsResponse], error) {
			for cs.Receive() {
				req := cs.Msg()
				rcmu.Lock()
				rcmap[toRunContextKeyString(req.GetId())] = req.GetValue()
				rcmu.Unlock()
			}
			if err := cs.Err(); err != nil {
				return nil, connect.NewError(connect.CodeUnknown, err)
			}
			return connect.NewResponse(&mgmtv1alpha1.SetRunContextsResponse{}), nil
		},
	))

	mux.Handle(mgmtv1alpha1connect.JobServiceSetRunContextProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceSetRunContextProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.SetRunContextRequest]) (*connect.Response[mgmtv1alpha1.SetRunContextResponse], error) {
			rcmu.RLock()
			defer rcmu.RUnlock()
			rcmap[toRunContextKeyString(r.Msg.GetId())] = r.Msg.GetValue()

			return connect.NewResponse(&mgmtv1alpha1.SetRunContextResponse{}), nil
		},
	))
}

func toRunContextKeyString(id *mgmtv1alpha1.RunContextKey) string {
	return fmt.Sprintf("%s.%s.%s", id.GetJobRunId(), id.GetExternalId(), id.GetAccountId())
}

func (s *IntegrationTestSuite) Test_Workflow_DynamoDB_Sync() {
	tests := getAllDynamoDBSyncTests()
	for groupName, group := range tests {
		group := group
		s.T().Run(groupName, func(t *testing.T) {
			// cannot run in parallel until we have each test create/delete its own tables.
			// t.Parallel()
			for _, tt := range group {
				t.Run(tt.Name, func(t *testing.T) {
					t.Logf("running integration test: %s \n", tt.Name)
					// setup
					sourceTableName := "test-sync-source"
					destTableName := "test-sync-dest"
					errgrp, errctx := errgroup.WithContext(s.ctx)
					errgrp.Go(func() error { return s.SetupDynamoDbTable(errctx, sourceTableName, "id") })
					errgrp.Go(func() error { return s.SetupDynamoDbTable(errctx, destTableName, "id") })
					err := errgrp.Wait()
					require.NoError(t, err)

					err = s.InsertDynamoDBRecords(sourceTableName, []map[string]dyntypes.AttributeValue{
						{
							"id": &dyntypes.AttributeValueMemberS{Value: "1"},
							"a":  &dyntypes.AttributeValueMemberBOOL{Value: true},
							"NestedMap": &dyntypes.AttributeValueMemberM{
								Value: map[string]dyntypes.AttributeValue{
									"Level1": &dyntypes.AttributeValueMemberM{
										Value: map[string]dyntypes.AttributeValue{
											"Level2": &dyntypes.AttributeValueMemberM{
												Value: map[string]dyntypes.AttributeValue{
													"Attribute1": &dyntypes.AttributeValueMemberS{Value: "Value1"},
													"NumberSet":  &dyntypes.AttributeValueMemberNS{Value: []string{"1", "2", "3"}},
													"BinaryData": &dyntypes.AttributeValueMemberB{Value: []byte("U29tZUJpbmFyeURhdGE=")},
													"Level3": &dyntypes.AttributeValueMemberM{
														Value: map[string]dyntypes.AttributeValue{
															"Attribute2": &dyntypes.AttributeValueMemberS{Value: "Value2"},
															"StringSet":  &dyntypes.AttributeValueMemberSS{Value: []string{"Item1", "Item2", "Item3"}},
															"BinarySet": &dyntypes.AttributeValueMemberBS{
																Value: [][]byte{
																	[]byte("U29tZUJpbmFyeQ=="),
																	[]byte("QW5vdGhlckJpbmFyeQ=="),
																},
															},
															"Level4": &dyntypes.AttributeValueMemberM{
																Value: map[string]dyntypes.AttributeValue{
																	"Attribute3":     &dyntypes.AttributeValueMemberS{Value: "Value3"},
																	"Boolean":        &dyntypes.AttributeValueMemberBOOL{Value: true},
																	"MoreBinaryData": &dyntypes.AttributeValueMemberB{Value: []byte("TW9yZUJpbmFyeURhdGE=")},
																	"MoreBinarySet": &dyntypes.AttributeValueMemberBS{
																		Value: [][]byte{
																			[]byte("TW9yZUJpbmFyeQ=="),
																			[]byte("QW5vdGhlck1vcmVCaW5hcnk="),
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						{
							"id": &dyntypes.AttributeValueMemberS{Value: "2"},
							"a":  &dyntypes.AttributeValueMemberBOOL{Value: false},
						},
					})
					require.NoError(t, err)

					jobId := "115aaf2c-776e-4847-8268-d914e3c15968"
					sourceConnectionId := "c9b6ce58-5c8e-4dce-870d-96841b19d988"
					destConnectionId := "226add85-5751-4232-b085-a0ae93afc7ce"

					sourceTableOpts := []*mgmtv1alpha1.DynamoDBSourceTableOption{}
					for table, where := range tt.SubsetMap {
						where := where
						sourceTableOpts = append(sourceTableOpts, &mgmtv1alpha1.DynamoDBSourceTableOption{
							Table:       table,
							WhereClause: &where,
						})
					}

					destOpts := &mgmtv1alpha1.DynamoDBDestinationConnectionOptions{
						TableMappings: []*mgmtv1alpha1.DynamoDBDestinationTableMapping{{SourceTable: sourceTableName, DestinationTable: destTableName}},
					}

					mux := http.NewServeMux()
					mux.Handle(mgmtv1alpha1connect.UserAccountServiceIsAccountStatusValidProcedure, connect.NewUnaryHandler(
						mgmtv1alpha1connect.UserAccountServiceIsAccountStatusValidProcedure,
						func(ctx context.Context, r *connect.Request[mgmtv1alpha1.IsAccountStatusValidRequest]) (*connect.Response[mgmtv1alpha1.IsAccountStatusValidResponse], error) {
							return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{IsValid: true}), nil
						},
					))
					mux.Handle(mgmtv1alpha1connect.JobServiceGetJobProcedure, connect.NewUnaryHandler(
						mgmtv1alpha1connect.JobServiceGetJobProcedure,
						func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetJobRequest]) (*connect.Response[mgmtv1alpha1.GetJobResponse], error) {
							return connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
								Job: &mgmtv1alpha1.Job{
									Id:        jobId,
									AccountId: "225aaf2c-776e-4847-8268-d914e3c15988",
									Source: &mgmtv1alpha1.JobSource{
										Options: &mgmtv1alpha1.JobSourceOptions{
											Config: &mgmtv1alpha1.JobSourceOptions_Dynamodb{
												Dynamodb: &mgmtv1alpha1.DynamoDBSourceConnectionOptions{
													ConnectionId: sourceConnectionId,
													Tables:       sourceTableOpts,
												},
											},
										},
									},
									Destinations: []*mgmtv1alpha1.JobDestination{
										{
											ConnectionId: destConnectionId,
											Options: &mgmtv1alpha1.JobDestinationOptions{
												Config: &mgmtv1alpha1.JobDestinationOptions_DynamodbOptions{
													DynamodbOptions: destOpts,
												},
											},
										},
									},
									Mappings:           tt.JobMappings,
									VirtualForeignKeys: tt.VirtualForeignKeys,
								}}), nil
						},
					))

					mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
						mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
						func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
							if r.Msg.GetId() == sourceConnectionId {
								return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
									Connection: &mgmtv1alpha1.Connection{
										Id:   sourceConnectionId,
										Name: "source",
										ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
											Config: &mgmtv1alpha1.ConnectionConfig_DynamodbConfig{
												DynamodbConfig: &mgmtv1alpha1.DynamoDBConnectionConfig{
													Credentials: s.dynamo.dtoAwsCreds,
													Endpoint:    &s.dynamo.endpoint,
												},
											},
										},
									},
								}), nil
							}
							if r.Msg.GetId() == destConnectionId {
								return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
									Connection: &mgmtv1alpha1.Connection{
										Id:   destConnectionId,
										Name: "target",
										ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
											Config: &mgmtv1alpha1.ConnectionConfig_DynamodbConfig{
												DynamodbConfig: &mgmtv1alpha1.DynamoDBConnectionConfig{
													Credentials: s.dynamo.dtoAwsCreds,
													Endpoint:    &s.dynamo.endpoint,
												},
											},
										},
									},
								}), nil
							}
							return nil, fmt.Errorf("unknown test connection")
						},
					))
					addRunContextProcedureMux(mux)
					addEmptyJobHooksProcedureMux(mux)
					srv := startHTTPServer(t, mux)
					env := executeWorkflow(t, srv, &s.redis.url, jobId, tt.ExpectError)
					require.Truef(t, env.IsWorkflowCompleted(), fmt.Sprintf("Workflow did not complete. Test: %s", tt.Name))
					err = env.GetWorkflowError()
					if tt.ExpectError {
						require.Error(t, err, "Did not received Temporal Workflow Error %s", tt.Name)
						return
					}
					require.NoError(t, err, "Received Temporal Workflow Error %s", tt.Name)

					for table, expected := range tt.Expected {
						out, err := s.dynamo.dynamoclient.Scan(s.ctx, &dynamodb.ScanInput{
							TableName: &table,
						})
						require.NoError(t, err)
						require.Equal(t, expected.RowCount, int(out.Count), fmt.Sprintf("Test: %s Table: %s", tt.Name, table))
					}

					// tear down
					errgrp, errctx = errgroup.WithContext(s.ctx)
					errgrp.Go(func() error { return s.DestroyDynamoDbTable(errctx, sourceTableName) })
					errgrp.Go(func() error { return s.DestroyDynamoDbTable(errctx, destTableName) })
					err = errgrp.Wait()
					require.NoError(t, err)
				})
			}
		})
	}
}

func getAllDynamoDBSyncTests() map[string][]*workflow_testdata.IntegrationTest {
	allTests := map[string][]*workflow_testdata.IntegrationTest{}

	allTests["Standard Sync"] = []*workflow_testdata.IntegrationTest{
		{
			Name: "Passthrough Sync",
			JobMappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "aws",
					Table:  "test-sync-source",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
						},
					},
				},
				{
					Schema: "aws",
					Table:  "test-sync-source",
					Column: "a",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
						},
					},
				},
			},
			JobOptions: &workflow_testdata.TestJobOptions{},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"test-sync-source": {RowCount: 2},
				"test-sync-dest":   {RowCount: 2},
			},
		},
		{
			Name: "Subset Sync",
			JobMappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "aws",
					Table:  "test-sync-source",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
						},
					},
				},
				{
					Schema: "aws",
					Table:  "test-sync-source",
					Column: "a",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
						},
					},
				},
			},
			SubsetMap: map[string]string{
				"test-sync-source": "id = '1'",
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"test-sync-source": {RowCount: 2},
				"test-sync-dest":   {RowCount: 1},
			},
		},
		{
			Name: "Default Transformer Sync",
			JobMappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "aws",
					Table:  "test-sync-source",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
						},
					},
				},
				{
					Schema: "aws",
					Table:  "test-sync-source",
					Column: "a",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
						},
					},
				},
			},
			JobOptions: &workflow_testdata.TestJobOptions{
				DefaultTransformers: &workflow_testdata.DefaultTransformers{
					Boolean: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{},
						},
					},
					Number: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
								TransformInt64Config: &mgmtv1alpha1.TransformInt64{RandomizationRangeMin: gotypeutil.ToPtr(int64(10)), RandomizationRangeMax: gotypeutil.ToPtr(int64(1000))},
							},
						},
					},
					String: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
								TransformStringConfig: &mgmtv1alpha1.TransformString{},
							},
						},
					},
					Byte: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
						},
					},
				},
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"test-sync-source": {RowCount: 2},
				"test-sync-dest":   {RowCount: 2},
			},
		},
	}
	return allTests
}

func (s *IntegrationTestSuite) Test_Workflow_MongoDB_Sync() {
	tests := getAllMongoDBSyncTests()
	for groupName, group := range tests {
		group := group
		s.T().Run(groupName, func(t *testing.T) {
			// cannot run in parallel until we have each test create/delete its own tables.
			// t.Parallel()
			for _, tt := range group {
				t.Run(tt.Name, func(t *testing.T) {
					t.Logf("running integration test: %s \n", tt.Name)
					// setup
					dbName := "data"
					collectionName := "test-sync"

					doc := bson.D{
						{Key: "_id", Value: primitive.NewObjectID()},
						{Key: "string", Value: "Hello, MongoDB!"},
						{Key: "bool", Value: true},
						{Key: "int32", Value: int32(42)},
						{Key: "int64", Value: int64(92233720)},
						{Key: "double", Value: 3.14159},
						{Key: "decimal128", Value: primitive.NewDecimal128(3, 14159)},
						{Key: "date", Value: primitive.NewDateTimeFromTime(time.Now())},
						{Key: "timestamp", Value: primitive.Timestamp{T: 1645553494, I: 1}},
						{Key: "null", Value: primitive.Null{}},
						{Key: "regex", Value: primitive.Regex{Pattern: "^test", Options: "i"}},
						{Key: "array", Value: bson.A{"apple", "banana", "cherry"}},
						{Key: "embedded_document", Value: bson.D{
							{Key: "name", Value: "John Doe"},
							{Key: "age", Value: 30},
						}},
						{Key: "binary", Value: primitive.Binary{Subtype: 0x80, Data: []byte("binary data")}},
						{Key: "undefined", Value: primitive.Undefined{}},
						{Key: "object_id", Value: primitive.NewObjectID()},
						{Key: "min_key", Value: primitive.MinKey{}},
						{Key: "max_key", Value: primitive.MaxKey{}},
					}
					docs := []any{doc}

					count, err := s.InsertMongoDbRecords(s.mongodb.source.client, dbName, collectionName, docs)
					require.NoError(t, err)
					require.Greater(t, count, 0)

					jobId := "115aaf2c-776e-4847-8268-d914e3c15968"
					sourceConnectionId := "c9b6ce58-5c8e-4dce-870d-96841b19d988"
					destConnectionId := "226add85-5751-4232-b085-a0ae93afc7ce"

					mux := http.NewServeMux()
					mux.Handle(mgmtv1alpha1connect.UserAccountServiceIsAccountStatusValidProcedure, connect.NewUnaryHandler(
						mgmtv1alpha1connect.UserAccountServiceIsAccountStatusValidProcedure,
						func(ctx context.Context, r *connect.Request[mgmtv1alpha1.IsAccountStatusValidRequest]) (*connect.Response[mgmtv1alpha1.IsAccountStatusValidResponse], error) {
							return connect.NewResponse(&mgmtv1alpha1.IsAccountStatusValidResponse{IsValid: true}), nil
						},
					))
					mux.Handle(mgmtv1alpha1connect.JobServiceGetJobProcedure, connect.NewUnaryHandler(
						mgmtv1alpha1connect.JobServiceGetJobProcedure,
						func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetJobRequest]) (*connect.Response[mgmtv1alpha1.GetJobResponse], error) {
							return connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
								Job: &mgmtv1alpha1.Job{
									Id:        jobId,
									AccountId: "225aaf2c-776e-4847-8268-d914e3c15988",
									Source: &mgmtv1alpha1.JobSource{
										Options: &mgmtv1alpha1.JobSourceOptions{
											Config: &mgmtv1alpha1.JobSourceOptions_Mongodb{
												Mongodb: &mgmtv1alpha1.MongoDBSourceConnectionOptions{
													ConnectionId: sourceConnectionId,
												},
											},
										},
									},
									Destinations: []*mgmtv1alpha1.JobDestination{
										{
											ConnectionId: destConnectionId,
											Options: &mgmtv1alpha1.JobDestinationOptions{
												Config: &mgmtv1alpha1.JobDestinationOptions_MongodbOptions{
													MongodbOptions: &mgmtv1alpha1.MongoDBDestinationConnectionOptions{},
												},
											},
										},
									},
									Mappings:           tt.JobMappings,
									VirtualForeignKeys: tt.VirtualForeignKeys,
								}}), nil
						},
					))

					mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
						mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
						func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
							if r.Msg.GetId() == sourceConnectionId {
								return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
									Connection: &mgmtv1alpha1.Connection{
										Id:   sourceConnectionId,
										Name: "source",
										ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
											Config: &mgmtv1alpha1.ConnectionConfig_MongoConfig{
												MongoConfig: &mgmtv1alpha1.MongoConnectionConfig{
													ConnectionConfig: &mgmtv1alpha1.MongoConnectionConfig_Url{
														Url: s.mongodb.source.url,
													},
												},
											},
										},
									},
								}), nil
							}
							if r.Msg.GetId() == destConnectionId {
								return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
									Connection: &mgmtv1alpha1.Connection{
										Id:   destConnectionId,
										Name: "target",
										ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
											Config: &mgmtv1alpha1.ConnectionConfig_MongoConfig{
												MongoConfig: &mgmtv1alpha1.MongoConnectionConfig{
													ConnectionConfig: &mgmtv1alpha1.MongoConnectionConfig_Url{
														Url: s.mongodb.target.url,
													},
												},
											},
										},
									},
								}), nil
							}
							return nil, fmt.Errorf("unknown test connection")
						},
					))
					addRunContextProcedureMux(mux)
					addEmptyJobHooksProcedureMux(mux)
					srv := startHTTPServer(t, mux)
					env := executeWorkflow(t, srv, &s.redis.url, jobId, tt.ExpectError)
					require.Truef(t, env.IsWorkflowCompleted(), fmt.Sprintf("Workflow did not complete. Test: %s", tt.Name))
					err = env.GetWorkflowError()
					if tt.ExpectError {
						require.Error(t, err, "Did not received Temporal Workflow Error %s", tt.Name)
						return
					}
					require.NoError(t, err, "Received Temporal Workflow Error %s", tt.Name)

					for table, expected := range tt.Expected {
						col := s.mongodb.target.client.Database(dbName).Collection(collectionName)
						cursor, err := col.Find(s.ctx, bson.D{})
						require.NoError(t, err)
						var results []bson.M
						for cursor.Next(s.ctx) {
							var doc bson.M
							err = cursor.Decode(&doc)
							require.NoError(t, err)
							results = append(results, doc)
						}
						cursor.Close(s.ctx)
						require.Equal(t, expected.RowCount, len(results), fmt.Sprintf("Test: %s Table: %s", tt.Name, table))
					}

					// tear down
					errgrp, errctx := errgroup.WithContext(s.ctx)
					errgrp.Go(func() error { return s.DropMongoDbCollection(errctx, s.mongodb.source.client, dbName, collectionName) })
					errgrp.Go(func() error { return s.DropMongoDbCollection(errctx, s.mongodb.target.client, dbName, collectionName) })
					err = errgrp.Wait()
					require.NoError(t, err)
				})
			}
		})
	}
}

func getAllMongoDBSyncTests() map[string][]*workflow_testdata.IntegrationTest {
	allTests := map[string][]*workflow_testdata.IntegrationTest{}
	allTests["Standard Sync"] = []*workflow_testdata.IntegrationTest{
		{
			Name: "Passthrough Sync",
			JobMappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "data",
					Table:  "test-sync",
					Column: "string",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
						},
					},
				},
				{
					Schema: "data",
					Table:  "test-sync",
					Column: "bool",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
						},
					},
				},
			},
			JobOptions: &workflow_testdata.TestJobOptions{},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"test-sync": {RowCount: 1},
			},
		},
		{
			Name: "Transform Sync",
			JobMappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "data",
					Table:  "test-sync",
					Column: "string",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
								TransformStringConfig: &mgmtv1alpha1.TransformString{
									PreserveLength: gotypeutil.ToPtr(true),
								},
							},
						},
					},
				},
				{
					Schema: "data",
					Table:  "test-sync",
					Column: "embedded_document.name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
								GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
							},
						},
					},
				},
				{
					Schema: "data",
					Table:  "test-sync",
					Column: "decimal128",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformFloat64Config{
								TransformFloat64Config: &mgmtv1alpha1.TransformFloat64{
									RandomizationRangeMin: gotypeutil.ToPtr(float64(0)),
									RandomizationRangeMax: gotypeutil.ToPtr(float64(300)),
								},
							},
						},
					},
				},
				{
					Schema: "data",
					Table:  "test-sync",
					Column: "int64",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
								TransformInt64Config: &mgmtv1alpha1.TransformInt64{
									RandomizationRangeMin: gotypeutil.ToPtr(int64(0)),
									RandomizationRangeMax: gotypeutil.ToPtr(int64(300)),
								},
							},
						},
					},
				},
				{
					Schema: "data",
					Table:  "test-sync",
					Column: "timestamp",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{
								GenerateUnixtimestampConfig: &mgmtv1alpha1.GenerateUnixTimestamp{},
							},
						},
					},
				},
			},
			JobOptions: &workflow_testdata.TestJobOptions{},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"test-sync": {RowCount: 1},
			},
		},
	}
	return allTests
}

type fakeEELicense struct{}

func (f *fakeEELicense) IsValid() bool {
	return true
}

func executeWorkflow(
	t testing.TB,
	srv *httptest.Server,
	redisUrl *string,
	jobId string,
	expectActivityErr bool,
) *testsuite.TestWorkflowEnvironment {
	t.Helper()
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(srv.Client(), srv.URL)
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	transformerclient := mgmtv1alpha1connect.NewTransformersServiceClient(srv.Client(), srv.URL)
	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(srv.Client(), srv.URL)
	var redisconfig *shared.RedisConfig
	if redisUrl != nil && *redisUrl != "" {
		redisconfig = &shared.RedisConfig{
			Url:  *redisUrl,
			Kind: "simple",
			Tls: &shared.RedisTlsConfig{
				Enabled: false,
			},
		}
	}
	redisclient, err := neosync_redis.GetRedisClient(redisconfig)
	if err != nil {
		t.Fatal(err)
	}

	sqlconnmanager := connectionmanager.NewConnectionManager(sqlprovider.NewProvider(&sqlconnect.SqlOpenConnector{}), connectionmanager.WithReaperPoll(10*time.Second))
	go sqlconnmanager.Reaper(testutil.GetConcurrentTestLogger(t))
	mongoconnmanager := connectionmanager.NewConnectionManager(mongoprovider.NewProvider())
	go mongoconnmanager.Reaper(testutil.GetConcurrentTestLogger(t))

	t.Cleanup(func() {
		sqlconnmanager.Shutdown(testutil.GetConcurrentTestLogger(t))
		mongoconnmanager.Shutdown(testutil.GetConcurrentTestLogger(t))
	})

	sqlmanager := sql_manager.NewSqlManager(
		sql_manager.WithConnectionManager(sqlconnmanager),
	)

	// temporal workflow
	testSuite := &testsuite.WorkflowTestSuite{}
	testSuite.SetLogger(log.NewStructuredLogger(testutil.GetConcurrentTestLogger(t)))
	env := testSuite.NewTestWorkflowEnvironment()

	// register activities
	genbenthosActivity := genbenthosconfigs_activity.New(
		jobclient,
		connclient,
		transformerclient,
		sqlmanager,
		redisconfig,
		false,
	)
	var activityMeter metric.Meter

	syncActivity := sync_activity.New(connclient, jobclient, sqlconnmanager, mongoconnmanager, activityMeter, sync_activity.NewBenthosStreamManager())
	retrieveActivityOpts := syncactivityopts_activity.New(jobclient)
	runSqlInitTableStatements := runsqlinittablestmts_activity.New(jobclient, connclient, sqlmanager, &fakeEELicense{})
	accountStatusActivity := accountstatus_activity.New(userclient)
	jobhookTimingActivity := jobhooks_by_timing_activity.New(jobclient, connclient, sqlmanager, &fakeEELicense{})
	posttableSyncActivity := posttablesync_activity.New(jobclient, sqlmanager, connclient)
	redisCleanUpActivity := syncrediscleanup_activity.New(redisclient)

	env.RegisterWorkflow(Workflow)
	env.RegisterActivity(syncActivity.Sync)
	env.RegisterActivity(retrieveActivityOpts.RetrieveActivityOptions)
	env.RegisterActivity(runSqlInitTableStatements.RunSqlInitTableStatements)
	env.RegisterActivity(redisCleanUpActivity.DeleteRedisHash)
	env.RegisterActivity(genbenthosActivity.GenerateBenthosConfigs)
	env.RegisterActivity(accountStatusActivity.CheckAccountStatus)
	env.RegisterActivity(jobhookTimingActivity.RunJobHooksByTiming)
	env.RegisterActivity(posttableSyncActivity.RunPostTableSync)
	env.SetTestTimeout(600 * time.Second) // increase the test timeout

	env.SetOnActivityCompletedListener(func(activityInfo *activity.Info, result converter.EncodedValue, err error) {
		if !expectActivityErr {
			require.NoError(t, err, "Activity %s failed", activityInfo.ActivityType.Name)
		}
		if activityInfo.ActivityType.Name == "RunPostTableSync" && result.HasValue() {
			var postTableSyncResp posttablesync_activity.RunPostTableSyncResponse
			decodeErr := result.Get(&postTableSyncResp)
			require.NoError(t, decodeErr, "Failed to decode result for activity %s", activityInfo.ActivityType.Name)

			require.Emptyf(t, postTableSyncResp.Errors, "Post table sync activity returned errors: %v", formatPostTableSyncErrors(postTableSyncResp.Errors))
		}
	})

	env.SetStartWorkflowOptions(client.StartWorkflowOptions{ID: jobId})
	env.ExecuteWorkflow(Workflow, &WorkflowRequest{JobId: jobId})
	return env
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}

func formatPostTableSyncErrors(errors []*posttablesync_activity.PostTableSyncError) []string {
	formatted := []string{}
	for _, err := range errors {
		for _, e := range err.Errors {
			formatted = append(formatted, fmt.Sprintf("statement: %s  error: %s", e.Statement, e.Error))
		}
	}
	return formatted
}
