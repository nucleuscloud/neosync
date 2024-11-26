package datasync_workflow

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dyntypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	accountstatus_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/account-status"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	jobhooks_by_timing_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/jobhooks-by-timing"
	runsqlinittablestmts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/run-sql-init-table-stmts"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
	testdata_javascripttransformers "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/javascript-transformers"
	mssql_datatypes "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/mssql/data-types"
	mssql_simple "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/mssql/simple"

	mysql_alltypes "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/mysql/all-types"
	mysql_compositekeys "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/mysql/composite-keys"
	mysql_initschema "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/mysql/init-schema"
	mysql_multipledbs "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/mysql/multiple-dbs"
	testdata_pgtypes "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/all-types"
	testdata_circulardependencies "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/circular-dependencies"
	testdata_doublereference "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/double-reference"
	testdata_subsetting "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/subsetting"
	testdata_virtualforeignkeys "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/virtual-foreign-keys"
	testdata_primarykeytransformer "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/primary-key-transformer"
	testdata_skipfkviolations "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/skip-fk-violations"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/errgroup"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	temporalmocks "go.temporal.io/sdk/mocks"
	"go.temporal.io/sdk/testsuite"
)

func getAllPostgresSyncTests() map[string][]*workflow_testdata.IntegrationTest {
	allTests := map[string][]*workflow_testdata.IntegrationTest{}
	drTests := testdata_doublereference.GetSyncTests()
	vfkTests := testdata_virtualforeignkeys.GetSyncTests()
	cdTests := testdata_circulardependencies.GetSyncTests()
	javascriptTests := testdata_javascripttransformers.GetSyncTests()
	pkTransformationTests := testdata_primarykeytransformer.GetSyncTests()
	subsettingTests := testdata_subsetting.GetSyncTests()
	pgTypesTests := testdata_pgtypes.GetSyncTests()
	skipFkViolationTests := testdata_skipfkviolations.GetSyncTests()

	allTests["Double_References"] = drTests
	allTests["Virtual_Foreign_Keys"] = vfkTests
	allTests["Circular_Dependencies"] = cdTests
	allTests["Javascript_Transformers"] = javascriptTests
	allTests["Primary_Key_Transformers"] = pkTransformationTests
	allTests["Subsetting"] = subsettingTests
	allTests["PG_Types"] = pgTypesTests
	allTests["Skip_ForeignKey_Violations"] = skipFkViolationTests
	return allTests
}

func (s *IntegrationTestSuite) Test_Workflow_Sync_Postgres() {
	tests := getAllPostgresSyncTests()
	for groupName, group := range tests {
		group := group
		s.T().Run(groupName, func(t *testing.T) {
			t.Parallel()
			for _, tt := range group {
				t.Run(tt.Name, func(t *testing.T) {
					t.Logf("running integration test: %s \n", tt.Name)
					// setup
					err := s.postgres.Source.RunSqlFiles(s.ctx, &tt.Folder, tt.SourceFilePaths)
					require.NoError(t, err)
					err = s.postgres.Target.RunSqlFiles(s.ctx, &tt.Folder, tt.TargetFilePaths)
					require.NoError(t, err)

					schemas := []*mgmtv1alpha1.PostgresSourceSchemaOption{}
					subsetMap := map[string]*mgmtv1alpha1.PostgresSourceSchemaOption{}
					for table, where := range tt.SubsetMap {
						schema, table := sqlmanager_shared.SplitTableKey(table)
						if _, exists := subsetMap[schema]; !exists {
							subsetMap[schema] = &mgmtv1alpha1.PostgresSourceSchemaOption{
								Schema: schema,
								Tables: []*mgmtv1alpha1.PostgresSourceTableOption{},
							}
						}
						w := where
						subsetMap[schema].Tables = append(subsetMap[schema].Tables, &mgmtv1alpha1.PostgresSourceTableOption{
							Table:       table,
							WhereClause: &w,
						})
					}

					for _, s := range subsetMap {
						schemas = append(schemas, s)
					}

					var subsetByForeignKeyConstraints bool
					var destinationOptions *mgmtv1alpha1.JobDestinationOptions
					if tt.JobOptions != nil {
						if tt.JobOptions.SubsetByForeignKeyConstraints {
							subsetByForeignKeyConstraints = true
						}
						destinationOptions = &mgmtv1alpha1.JobDestinationOptions{
							Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
								PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
									InitTableSchema: tt.JobOptions.InitSchema,
									TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
										TruncateBeforeInsert: tt.JobOptions.Truncate,
									},
									SkipForeignKeyViolations: tt.JobOptions.SkipForeignKeyViolations,
								},
							},
						}
					}

					jobId := "115aaf2c-776e-4847-8268-d914e3c15968"
					srcConnId := "c9b6ce58-5c8e-4dce-870d-96841b19d988"
					destConnId := "226add85-5751-4232-b085-a0ae93afc7ce"

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
											Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
												Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
													ConnectionId:                  srcConnId,
													Schemas:                       schemas,
													SubsetByForeignKeyConstraints: subsetByForeignKeyConstraints,
												},
											},
										},
									},
									Destinations: []*mgmtv1alpha1.JobDestination{
										{
											ConnectionId: destConnId,
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
							if r.Msg.GetId() == srcConnId {
								return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
									Connection: &mgmtv1alpha1.Connection{
										Id:   srcConnId,
										Name: "source",
										ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
											Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
												PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
													ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
														Url: s.postgres.Source.URL,
													},
												},
											},
										},
									},
								}), nil
							}
							if r.Msg.GetId() == destConnId {
								return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
									Connection: &mgmtv1alpha1.Connection{
										Id:   destConnId,
										Name: "target",
										ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
											Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
												PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
													ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
														Url: s.postgres.Target.URL,
													},
												},
											},
										},
									},
								}), nil
							}
							return nil, connect.NewError(connect.CodeInternal, errors.New("invalid test connection id"))
						},
					))
					mux.Handle(mgmtv1alpha1connect.JobServiceGetActiveJobHooksByTimingProcedure, connect.NewUnaryHandler(
						mgmtv1alpha1connect.JobServiceGetActiveJobHooksByTimingProcedure,
						func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetActiveJobHooksByTimingRequest]) (*connect.Response[mgmtv1alpha1.GetActiveJobHooksByTimingResponse], error) {
							if r.Msg.GetJobId() != jobId {
								return nil, connect.NewError(connect.CodeInternal, errors.New("invalid test job id"))
							}
							hooks := []*mgmtv1alpha1.JobHook{}
							if r.Msg.Timing == mgmtv1alpha1.GetActiveJobHooksByTimingRequest_TIMING_PRESYNC {
								hooks = append(hooks, &mgmtv1alpha1.JobHook{
									Id:       uuid.NewString(),
									Name:     "test-presync-hook-1",
									JobId:    jobId,
									Enabled:  true,
									Priority: 0,
									Config: &mgmtv1alpha1.JobHookConfig{Config: &mgmtv1alpha1.JobHookConfig_Sql{Sql: &mgmtv1alpha1.JobHookConfig_JobSqlHook{
										Query:        "select 1",
										ConnectionId: srcConnId,
										Timing:       &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{}},
									}}},
								})
							} else if r.Msg.Timing == mgmtv1alpha1.GetActiveJobHooksByTimingRequest_TIMING_POSTSYNC {
								hooks = append(hooks, &mgmtv1alpha1.JobHook{
									Id:       uuid.NewString(),
									Name:     "test-postsync-hook-1",
									JobId:    jobId,
									Enabled:  true,
									Priority: 0,
									Config: &mgmtv1alpha1.JobHookConfig{Config: &mgmtv1alpha1.JobHookConfig_Sql{Sql: &mgmtv1alpha1.JobHookConfig_JobSqlHook{
										Query:        "select 1",
										ConnectionId: destConnId,
										Timing:       &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing{Timing: &mgmtv1alpha1.JobHookConfig_JobSqlHook_Timing_PreSync{}},
									}}},
								})
							}
							return connect.NewResponse(&mgmtv1alpha1.GetActiveJobHooksByTimingResponse{Hooks: hooks}), nil
						},
					))

					addRunContextProcedureMux(mux)
					srv := startHTTPServer(t, mux)
					env := executeWorkflow(t, srv, s.redis.url, jobId)
					require.Truef(t, env.IsWorkflowCompleted(), fmt.Sprintf("Workflow did not complete. Test: %s", tt.Name))
					err = env.GetWorkflowError()
					if tt.ExpectError {
						require.Error(t, err, "Did not receive Temporal Workflow Error %s", tt.Name)
						return
					}
					require.NoError(t, err, "Received Temporal Workflow Error %s", tt.Name)

					for table, expected := range tt.Expected {
						rows, err := s.postgres.Target.DB.Query(s.ctx, fmt.Sprintf("select * from %s;", table))
						require.NoError(t, err)
						count := 0
						for rows.Next() {
							count++
						}
						require.Equalf(t, expected.RowCount, count, fmt.Sprintf("Test: %s Table: %s", tt.Name, table))
					}

					// tear down
					err = s.postgres.Source.RunSqlFiles(s.ctx, &tt.Folder, []string{"teardown.sql"})
					require.NoError(t, err)
					err = s.postgres.Target.RunSqlFiles(s.ctx, &tt.Folder, []string{"teardown.sql"})
					require.NoError(t, err)
				})
			}
		})
	}
}

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
					var destinationOptions *mgmtv1alpha1.JobDestinationOptions
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
					env := executeWorkflow(t, srv, s.redis.url, "115aaf2c-776e-4847-8268-d914e3c15968")
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

func (s *IntegrationTestSuite) Test_Workflow_VirtualForeignKeys_Transform() {
	testFolder := "testdata/postgres/virtual-foreign-keys"
	// setup
	err := s.postgres.Source.RunSqlFiles(s.ctx, &testFolder, []string{"source-setup.sql"})
	require.NoError(s.T(), err)
	err = s.postgres.Target.RunSqlFiles(s.ctx, &testFolder, []string{"target-setup.sql"})
	require.NoError(s.T(), err)

	virtualForeignKeys := testdata_virtualforeignkeys.GetVirtualForeignKeys()
	jobmappings := testdata_virtualforeignkeys.GetDefaultSyncJobMappings()

	for _, m := range jobmappings {
		if m.Table == "countries" && m.Column == "country_id" {
			m.Transformer = &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
						TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: `if (value == 'US') { return 'SU'; } return value;`},
					},
				},
			}
		}
	}
	// neosync api mocks
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
					Id:        "fd4d8660-31a0-48b2-9adf-10f11b94898f",
					AccountId: "225aaf2c-776e-4847-8268-d914e3c15988",
					Source: &mgmtv1alpha1.JobSource{
						Options: &mgmtv1alpha1.JobSourceOptions{
							Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
								Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
									ConnectionId: "c9b6ce58-5c8e-4dce-870d-96841b19d988",
								},
							},
						},
					},
					Mappings:           jobmappings,
					VirtualForeignKeys: virtualForeignKeys,
					Destinations: []*mgmtv1alpha1.JobDestination{
						{
							ConnectionId: "226add85-5751-4232-b085-a0ae93afc7ce",
						},
					},
				},
			}), nil
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
							Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
								PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
									ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
										Url: s.postgres.Source.URL,
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
							Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
								PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
									ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
										Url: s.postgres.Target.URL,
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
	srv := startHTTPServer(s.T(), mux)
	testName := "Virtual Foreign Key primary key transform"
	env := executeWorkflow(s.T(), srv, s.redis.url, "fd4d8660-31a0-48b2-9adf-10f11b94898f")
	require.Truef(s.T(), env.IsWorkflowCompleted(), fmt.Sprintf("Workflow did not complete. Test: %s", testName))
	err = env.GetWorkflowError()
	require.NoError(s.T(), err, "Received Temporal Workflow Error %s", testName)

	tables := []string{"regions", "countries", "locations", "departments", "dependents", "jobs", "employees"}
	for _, t := range tables {
		rows, err := s.postgres.Target.DB.Query(s.ctx, fmt.Sprintf("select * from vfk_hr.%s;", t))
		require.NoError(s.T(), err)
		count := 0
		for rows.Next() {
			count++
		}
		require.Greater(s.T(), count, 0)
		require.NoError(s.T(), err)
	}

	rows := s.postgres.Source.DB.QueryRow(s.ctx, "select count(*) from vfk_hr.countries where country_id = 'US';")
	var rowCount int
	err = rows.Scan(&rowCount)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, rowCount)

	rows = s.postgres.Source.DB.QueryRow(s.ctx, "select count(*) from vfk_hr.locations where country_id = 'US';")
	err = rows.Scan(&rowCount)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 3, rowCount)

	rows = s.postgres.Target.DB.QueryRow(s.ctx, "select count(*) from vfk_hr.countries where country_id = 'US';")
	err = rows.Scan(&rowCount)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 0, rowCount)

	rows = s.postgres.Target.DB.QueryRow(s.ctx, "select count(*) from vfk_hr.countries where country_id = 'SU';")
	err = rows.Scan(&rowCount)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, rowCount)

	rows = s.postgres.Target.DB.QueryRow(s.ctx, "select count(*) from vfk_hr.locations where country_id = 'SU';")
	err = rows.Scan(&rowCount)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 3, rowCount)

	// tear down
	err = s.postgres.Source.RunSqlFiles(s.ctx, &testFolder, []string{"teardown.sql"})
	require.NoError(s.T(), err)
	err = s.postgres.Target.RunSqlFiles(s.ctx, &testFolder, []string{"teardown.sql"})
	require.NoError(s.T(), err)
}

func getAllMysqlSyncTests() map[string][]*workflow_testdata.IntegrationTest {
	allTests := map[string][]*workflow_testdata.IntegrationTest{}
	mdTests := mysql_multipledbs.GetSyncTests()
	initTests := mysql_initschema.GetSyncTests()
	compositeTests := mysql_compositekeys.GetSyncTests()
	dataTypesTests := mysql_alltypes.GetSyncTests()
	allTests["Multiple_Dbs"] = mdTests
	allTests["Composite_Keys"] = compositeTests
	allTests["Init_Schema"] = initTests
	allTests["DataTypes"] = dataTypesTests
	return allTests
}

func (s *IntegrationTestSuite) Test_Workflow_Sync_Mysql() {
	tests := getAllMysqlSyncTests()
	for groupName, group := range tests {
		group := group
		s.T().Run(groupName, func(t *testing.T) {
			t.Parallel()
			for _, tt := range group {
				t.Run(tt.Name, func(t *testing.T) {
					t.Logf("running integration test: %s \n", tt.Name)
					// setup
					err := s.mysql.Source.RunSqlFiles(s.ctx, &tt.Folder, tt.SourceFilePaths)
					require.NoError(t, err)
					err = s.mysql.Target.RunSqlFiles(s.ctx, &tt.Folder, tt.TargetFilePaths)
					require.NoError(t, err)

					schemas := []*mgmtv1alpha1.MysqlSourceSchemaOption{}
					subsetMap := map[string]*mgmtv1alpha1.MysqlSourceSchemaOption{}
					for table, where := range tt.SubsetMap {
						schema, table := sqlmanager_shared.SplitTableKey(table)
						if _, exists := subsetMap[schema]; !exists {
							subsetMap[schema] = &mgmtv1alpha1.MysqlSourceSchemaOption{
								Schema: schema,
								Tables: []*mgmtv1alpha1.MysqlSourceTableOption{},
							}
						}
						w := where
						subsetMap[schema].Tables = append(subsetMap[schema].Tables, &mgmtv1alpha1.MysqlSourceTableOption{
							Table:       table,
							WhereClause: &w,
						})
					}

					for _, s := range subsetMap {
						schemas = append(schemas, s)
					}

					var subsetByForeignKeyConstraints bool
					var destinationOptions *mgmtv1alpha1.JobDestinationOptions
					if tt.JobOptions != nil {
						if tt.JobOptions.SubsetByForeignKeyConstraints {
							subsetByForeignKeyConstraints = true
						}
						destinationOptions = &mgmtv1alpha1.JobDestinationOptions{
							Config: &mgmtv1alpha1.JobDestinationOptions_MysqlOptions{
								MysqlOptions: &mgmtv1alpha1.MysqlDestinationConnectionOptions{
									InitTableSchema: tt.JobOptions.InitSchema,
									TruncateTable: &mgmtv1alpha1.MysqlTruncateTableConfig{
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
											Config: &mgmtv1alpha1.JobSourceOptions_Mysql{
												Mysql: &mgmtv1alpha1.MysqlSourceConnectionOptions{
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
											Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
												MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
													ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
														Url: s.mysql.Source.URL,
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
											Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
												MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
													ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
														Url: s.mysql.Target.URL,
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
					env := executeWorkflow(t, srv, s.redis.url, "115aaf2c-776e-4847-8268-d914e3c15968")
					require.Truef(t, env.IsWorkflowCompleted(), fmt.Sprintf("Workflow did not complete. Test: %s", tt.Name))
					err = env.GetWorkflowError()
					if tt.ExpectError {
						require.Error(t, err, "Did not received Temporal Workflow Error %s", tt.Name)
						return
					}
					require.NoError(t, err, "Received Temporal Workflow Error %s", tt.Name)

					for table, expected := range tt.Expected {
						rows, err := s.mysql.Target.DB.QueryContext(s.ctx, fmt.Sprintf("select * from %s;", table))
						require.NoError(t, err)
						count := 0
						for rows.Next() {
							count++
						}
						require.Equalf(t, expected.RowCount, count, fmt.Sprintf("Test: %s Table: %s", tt.Name, table))
					}

					// tear down
					err = s.mysql.Source.RunSqlFiles(s.ctx, &tt.Folder, []string{"teardown.sql"})
					require.NoError(t, err)
					err = s.mysql.Target.RunSqlFiles(s.ctx, &tt.Folder, []string{"teardown.sql"})
					require.NoError(t, err)
				})
			}
		})
	}
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
					env := executeWorkflow(t, srv, s.redis.url, jobId)
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
					env := executeWorkflow(t, srv, s.redis.url, jobId)
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

func (s *IntegrationTestSuite) Test_Workflow_Generate() {
	// setup
	testName := "Generate Job"
	folder := "testdata/generate-job"
	err := s.postgres.Target.RunSqlFiles(s.ctx, &folder, []string{"setup.sql"})
	require.NoError(s.T(), err)

	connectionId := "226add85-5751-4232-b085-a0ae93afc7ce"
	schema := "generate_job"
	table := "regions"

	destinationOptions := &mgmtv1alpha1.JobDestinationOptions{
		Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
			PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
				InitTableSchema: false,
				TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
					TruncateBeforeInsert: false,
				},
				SkipForeignKeyViolations: false,
			},
		},
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
							Config: &mgmtv1alpha1.JobSourceOptions_Generate{
								Generate: &mgmtv1alpha1.GenerateSourceOptions{
									Schemas: []*mgmtv1alpha1.GenerateSourceSchemaOption{{Schema: schema, Tables: []*mgmtv1alpha1.GenerateSourceTableOption{
										{Table: table, RowCount: 10},
									}}},
									FkSourceConnectionId: &connectionId,
								},
							},
						},
					},
					Destinations: []*mgmtv1alpha1.JobDestination{
						{
							ConnectionId: connectionId,
							Options:      destinationOptions,
						},
					},
					Mappings: []*mgmtv1alpha1.JobMapping{
						{Schema: schema, Table: table, Column: "region_id", Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{}},
						}},
						{Schema: schema, Table: table, Column: "region_name", Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
									GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
								},
							},
						}},
					},
				}}), nil
		},
	))

	mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
			if r.Msg.GetId() == "226add85-5751-4232-b085-a0ae93afc7ce" {
				return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
					Connection: &mgmtv1alpha1.Connection{
						Id:   "226add85-5751-4232-b085-a0ae93afc7ce",
						Name: "target",
						ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
							Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
								PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
									ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
										Url: s.postgres.Target.URL,
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
	srv := startHTTPServer(s.T(), mux)
	env := executeWorkflow(s.T(), srv, s.redis.url, "115aaf2c-776e-4847-8268-d914e3c15968")
	require.Truef(s.T(), env.IsWorkflowCompleted(), fmt.Sprintf("Workflow did not complete. Test: %s", testName))
	err = env.GetWorkflowError()
	require.NoError(s.T(), err, "Received Temporal Workflow Error %s", testName)

	rows, err := s.postgres.Target.DB.Query(s.ctx, fmt.Sprintf("select * from %s.%s;", schema, table))
	require.NoError(s.T(), err)
	count := 0
	for rows.Next() {
		count++
	}
	require.Equalf(s.T(), 10, count, fmt.Sprintf("Test: %s Table: %s", testName, table))

	// tear down
	err = s.postgres.Target.RunSqlFiles(s.ctx, &folder, []string{"teardown.sql"})
	require.NoError(s.T(), err)
}

type fakeEELicense struct{}

func (f *fakeEELicense) IsValid() bool {
	return true
}

func executeWorkflow(
	t *testing.T,
	srv *httptest.Server,
	redisUrl string,
	jobId string,
) *testsuite.TestWorkflowEnvironment {
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(srv.Client(), srv.URL)
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	transformerclient := mgmtv1alpha1connect.NewTransformersServiceClient(srv.Client(), srv.URL)
	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(srv.Client(), srv.URL)
	redisconfig := &shared.RedisConfig{
		Url:  redisUrl,
		Kind: "simple",
		Tls: &shared.RedisTlsConfig{
			Enabled: false,
		},
	}
	temporalClientMock := temporalmocks.NewClient(t)

	sqlmanager := sql_manager.NewSqlManager()

	// temporal workflow
	testSuite := &testsuite.WorkflowTestSuite{}
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
	disableReaper := true
	syncActivity := sync_activity.New(connclient, jobclient, &sqlconnect.SqlOpenConnector{}, &sync.Map{}, temporalClientMock, activityMeter, sync_activity.NewBenthosStreamManager(), disableReaper)
	retrieveActivityOpts := syncactivityopts_activity.New(jobclient)
	runSqlInitTableStatements := runsqlinittablestmts_activity.New(jobclient, connclient, sqlmanager, &fakeEELicense{})
	accountStatusActivity := accountstatus_activity.New(userclient)
	jobhookTimingActivity := jobhooks_by_timing_activity.New(jobclient, connclient, sqlmanager, &fakeEELicense{})

	env.RegisterWorkflow(Workflow)
	env.RegisterActivity(syncActivity.Sync)
	env.RegisterActivity(retrieveActivityOpts.RetrieveActivityOptions)
	env.RegisterActivity(runSqlInitTableStatements.RunSqlInitTableStatements)
	env.RegisterActivity(syncrediscleanup_activity.DeleteRedisHash)
	env.RegisterActivity(genbenthosActivity.GenerateBenthosConfigs)
	env.RegisterActivity(accountStatusActivity.CheckAccountStatus)
	env.RegisterActivity(jobhookTimingActivity.RunJobHooksByTiming)
	env.SetTestTimeout(600 * time.Second) // increase the test timeout

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
