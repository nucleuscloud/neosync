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
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	runsqlinittablestmts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/run-sql-init-table-stmts"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
	testdata_javascripttransformers "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/javascript-transformers"
	mysql_compositekeys "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/mysql/composite-keys"
	mysql_initschema "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/mysql/init-schema"
	mysql_multipledbs "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/mysql/multiple-dbs"
	testdata_circulardependencies "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/circular-dependencies"
	testdata_doublereference "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/double-reference"
	testdata_subsetting "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/subsetting"
	testdata_virtualforeignkeys "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/postgres/virtual-foreign-keys"
	testdata_primarykeytransformer "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata/primary-key-transformer"
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

	allTests["Double_References"] = drTests
	allTests["Virtual_Foreign_Keys"] = vfkTests
	allTests["Circular_Dependencies"] = cdTests
	allTests["Javascript_Transformers"] = javascriptTests
	allTests["Primary_Key_Transformers"] = pkTransformationTests
	allTests["Subsetting"] = subsettingTests
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
					s.RunPostgresSqlFiles(s.postgres.source.pool, tt.Folder, tt.SourceFilePaths)
					s.RunPostgresSqlFiles(s.postgres.target.pool, tt.Folder, tt.TargetFilePaths)

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
								},
							},
						}
					}

					mux := http.NewServeMux()
					mux.Handle(mgmtv1alpha1connect.JobServiceGetJobProcedure, connect.NewUnaryHandler(
						mgmtv1alpha1connect.JobServiceGetJobProcedure,
						func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetJobRequest]) (*connect.Response[mgmtv1alpha1.GetJobResponse], error) {
							return connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
								Job: &mgmtv1alpha1.Job{
									Id:        "115aaf2c-776e-4847-8268-d914e3c15968",
									AccountId: "225aaf2c-776e-4847-8268-d914e3c15988",
									Source: &mgmtv1alpha1.JobSource{
										Options: &mgmtv1alpha1.JobSourceOptions{
											Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
												Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
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
											Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
												PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
													ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
														Url: s.postgres.source.url,
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
														Url: s.postgres.target.url,
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
					srv := startHTTPServer(t, mux)
					executeWorkflow(t, srv, s.redis.url, "115aaf2c-776e-4847-8268-d914e3c15968", tt.Name)

					for table, expected := range tt.Expected {
						rows, err := s.postgres.target.pool.Query(s.ctx, fmt.Sprintf("select * from %s;", table))
						require.NoError(t, err)
						count := 0
						for rows.Next() {
							count++
						}
						require.Equalf(t, expected.RowCount, count, fmt.Sprintf("Test: %s Table: %s", tt.Name, table))
					}

					// tear down
					s.RunPostgresSqlFiles(s.postgres.source.pool, tt.Folder, []string{"teardown.sql"})
					s.RunPostgresSqlFiles(s.postgres.target.pool, tt.Folder, []string{"teardown.sql"})
				})
			}
		})
	}
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
}

func toRunContextKeyString(id *mgmtv1alpha1.RunContextKey) string {
	return fmt.Sprintf("%s.%s.%s", id.GetJobRunId(), id.GetExternalId(), id.GetAccountId())
}

func (s *IntegrationTestSuite) Test_Workflow_VirtualForeignKeys_Transform() {
	testFolder := "postgres/virtual-foreign-keys"
	// setup
	s.RunPostgresSqlFiles(s.postgres.source.pool, testFolder, []string{"source-setup.sql"})
	s.RunPostgresSqlFiles(s.postgres.target.pool, testFolder, []string{"target-setup.sql"})

	virtualForeignKeys := testdata_virtualforeignkeys.GetVirtualForeignKeys()
	jobmappings := testdata_virtualforeignkeys.GetDefaultSyncJobMappings()

	for _, m := range jobmappings {
		if m.Table == "countries" && m.Column == "country_id" {
			m.Transformer = &mgmtv1alpha1.JobMappingTransformer{
				Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
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
										Url: s.postgres.source.url,
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
										Url: s.postgres.target.url,
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
	srv := startHTTPServer(s.T(), mux)
	executeWorkflow(s.T(), srv, s.redis.url, "fd4d8660-31a0-48b2-9adf-10f11b94898f", "Virtual Foreign Key primary key transform")

	tables := []string{"regions", "countries", "locations", "departments", "dependents", "jobs", "employees"}
	for _, t := range tables {
		rows, err := s.postgres.target.pool.Query(s.ctx, fmt.Sprintf("select * from vfk_hr.%s;", t))
		require.NoError(s.T(), err)
		count := 0
		for rows.Next() {
			count++
		}
		require.Greater(s.T(), count, 0)
		require.NoError(s.T(), err)
	}

	rows := s.postgres.source.pool.QueryRow(s.ctx, "select count(*) from vfk_hr.countries where country_id = 'US';")
	var rowCount int
	err := rows.Scan(&rowCount)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, rowCount)

	rows = s.postgres.source.pool.QueryRow(s.ctx, "select count(*) from vfk_hr.locations where country_id = 'US';")
	err = rows.Scan(&rowCount)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 3, rowCount)

	rows = s.postgres.target.pool.QueryRow(s.ctx, "select count(*) from vfk_hr.countries where country_id = 'US';")
	err = rows.Scan(&rowCount)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 0, rowCount)

	rows = s.postgres.target.pool.QueryRow(s.ctx, "select count(*) from vfk_hr.countries where country_id = 'SU';")
	err = rows.Scan(&rowCount)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 1, rowCount)

	rows = s.postgres.target.pool.QueryRow(s.ctx, "select count(*) from vfk_hr.locations where country_id = 'SU';")
	err = rows.Scan(&rowCount)
	require.NoError(s.T(), err)
	require.Equal(s.T(), 3, rowCount)

	// tear down
	s.RunPostgresSqlFiles(s.postgres.source.pool, testFolder, []string{"teardown.sql"})
	s.RunPostgresSqlFiles(s.postgres.target.pool, testFolder, []string{"teardown.sql"})
}

func getAllMysqlSyncTests() map[string][]*workflow_testdata.IntegrationTest {
	allTests := map[string][]*workflow_testdata.IntegrationTest{}
	mdTests := mysql_multipledbs.GetSyncTests()
	initTests := mysql_initschema.GetSyncTests()
	compositeTests := mysql_compositekeys.GetSyncTests()
	allTests["Multiple_Dbs"] = mdTests
	allTests["Composite_Keys"] = compositeTests
	allTests["Init_Schema"] = initTests
	return allTests
}

func (s *IntegrationTestSuite) Test_Workflow_Mysql_Sync() {
	tests := getAllMysqlSyncTests()
	for groupName, group := range tests {
		group := group
		s.T().Run(groupName, func(t *testing.T) {
			t.Parallel()
			for _, tt := range group {
				t.Run(tt.Name, func(t *testing.T) {
					t.Logf("running integration test: %s \n", tt.Name)
					// setup
					s.RunMysqlSqlFiles(s.mysql.source.pool, tt.Folder, tt.SourceFilePaths)
					s.RunMysqlSqlFiles(s.mysql.target.pool, tt.Folder, tt.TargetFilePaths)

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
								},
							},
						}
					}

					mux := http.NewServeMux()
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
														Url: s.mysql.source.url,
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
														Url: s.mysql.target.url,
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
					srv := startHTTPServer(t, mux)
					executeWorkflow(t, srv, s.redis.url, "115aaf2c-776e-4847-8268-d914e3c15968", tt.Name)

					for table, expected := range tt.Expected {
						rows, err := s.mysql.target.pool.QueryContext(s.ctx, fmt.Sprintf("select * from %s;", table))
						require.NoError(t, err)
						count := 0
						for rows.Next() {
							count++
						}
						require.Equalf(t, expected.RowCount, count, fmt.Sprintf("Test: %s Table: %s", tt.Name, table))
					}

					// tear down
					s.RunMysqlSqlFiles(s.mysql.source.pool, tt.Folder, []string{"teardown.sql"})
					s.RunMysqlSqlFiles(s.mysql.target.pool, tt.Folder, []string{"teardown.sql"})
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
					srv := startHTTPServer(t, mux)
					executeWorkflow(t, srv, s.redis.url, jobId, tt.Name)

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
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
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
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
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
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
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
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
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
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
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
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
						},
					},
				},
			},
			JobOptions: &workflow_testdata.TestJobOptions{
				DefaultTransformers: &workflow_testdata.DefaultTransformers{
					Boolean: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL,
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{},
						},
					},
					Number: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64,
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
								TransformInt64Config: &mgmtv1alpha1.TransformInt64{RandomizationRangeMin: 10, RandomizationRangeMax: 1000},
							},
						},
					},
					String: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_STRING,
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
								TransformStringConfig: &mgmtv1alpha1.TransformString{},
							},
						},
					},
					Byte: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						Config: &mgmtv1alpha1.TransformerConfig{},
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

func executeWorkflow(
	t *testing.T,
	srv *httptest.Server,
	redisUrl string,
	jobId string,
	testName string,
) {
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(srv.Client(), srv.URL)
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(srv.Client(), srv.URL)
	transformerclient := mgmtv1alpha1connect.NewTransformersServiceClient(srv.Client(), srv.URL)
	sqlconnector := &sqlconnect.SqlOpenConnector{}
	redisconfig := &shared.RedisConfig{
		Url:  redisUrl,
		Kind: "simple",
		Tls: &shared.RedisTlsConfig{
			Enabled: false,
		},
	}
	temporalClientMock := temporalmocks.NewClient(t)
	pgpoolmap := &sync.Map{}
	mysqlpoolmap := &sync.Map{}
	mssqlpoolmap := &sync.Map{}
	pgquerier := pg_queries.New()
	mysqlquerier := mysql_queries.New()
	mssqlquerier := mssql_queries.New()
	sqlmanager := sql_manager.NewSqlManager(pgpoolmap, pgquerier, mysqlpoolmap, mysqlquerier, mssqlpoolmap, mssqlquerier, sqlconnector)

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
	syncActivity := sync_activity.New(connclient, jobclient, &sync.Map{}, temporalClientMock, activityMeter, sync_activity.NewBenthosStreamManager(), disableReaper)
	retrieveActivityOpts := syncactivityopts_activity.New(jobclient)
	runSqlInitTableStatements := runsqlinittablestmts_activity.New(jobclient, connclient, sqlmanager)
	env.RegisterWorkflow(Workflow)
	env.RegisterActivity(syncActivity.Sync)
	env.RegisterActivity(retrieveActivityOpts.RetrieveActivityOptions)
	env.RegisterActivity(runSqlInitTableStatements.RunSqlInitTableStatements)
	env.RegisterActivity(syncrediscleanup_activity.DeleteRedisHash)
	env.RegisterActivity(genbenthosActivity.GenerateBenthosConfigs)
	env.SetTestTimeout(600 * time.Second) // increase the test timeout

	env.ExecuteWorkflow(Workflow, &WorkflowRequest{JobId: jobId})
	require.Truef(t, env.IsWorkflowCompleted(), fmt.Sprintf("Workflow did not complete. Test: %s", testName))

	err := env.GetWorkflowError()
	require.NoError(t, err, "Received Temporal Workflow Error", "testName", testName)
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}
