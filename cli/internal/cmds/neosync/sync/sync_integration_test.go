package sync_cmd

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/nucleuscloud/neosync/cli/internal/output"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcmysql "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/mysql"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	mysqlalltypes "github.com/nucleuscloud/neosync/internal/testutil/testdata/mysql/alltypes"
	pgalltypes "github.com/nucleuscloud/neosync/internal/testutil/testdata/postgres/alltypes"
	tcworkflow "github.com/nucleuscloud/neosync/worker/pkg/integration-test"
	"github.com/stretchr/testify/require"
)

const neosyncDbMigrationsPath = "../../../../../backend/sql/postgresql/schema"

func Test_Sync(t *testing.T) {
	t.Parallel()
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	ctx := context.Background()

	neosyncApi, err := tcneosyncapi.NewNeosyncApiTestClient(ctx, t, tcneosyncapi.WithMigrationsDirectory(neosyncDbMigrationsPath))
	if err != nil {
		t.Fatal(err)
	}

	connclient := neosyncApi.OSSUnauthenticatedLicensedClients.Connections()
	conndataclient := neosyncApi.OSSUnauthenticatedLicensedClients.ConnectionData()
	jobclient := neosyncApi.OSSUnauthenticatedLicensedClients.Jobs()

	dbManagers := tcworkflow.NewTestDatabaseManagers(t)
	connmanager := dbManagers.SqlConnManager
	sqlmanagerclient := dbManagers.SqlManager
	accountId := tcneosyncapi.CreatePersonalAccount(ctx, t, neosyncApi.OSSUnauthenticatedLicensedClients.Users())
	awsS3Config := testutil.GetTestAwsS3Config()
	s3Conn := tcneosyncapi.CreateS3Connection(
		ctx,
		t,
		connclient,
		accountId,
		"s3-conn",
		awsS3Config.Bucket,
		&awsS3Config.Region,
	)
	outputType := output.PlainOutput

	t.Run("postgres", func(t *testing.T) {
		t.Parallel()
		postgres, err := tcpostgres.NewPostgresTestSyncContainer(ctx, []tcpostgres.Option{}, []tcpostgres.Option{})
		if err != nil {
			t.Fatal(err)
		}

		testdataFolder := "../../../../../internal/testutil/testdata/postgres"
		sourceConn := tcneosyncapi.CreatePostgresConnection(ctx, t, connclient, accountId, "postgres-source", postgres.Source.URL)

		t.Run("postgres_sync", func(t *testing.T) {
			// can't be run in parallel yet
			// right now CLI sync and init schema takes everything in source and copies it to target since there are no job mappings defined by the user
			// so it can't be scoped to specific schema
			// t.Parallel()
			err = postgres.Source.RunCreateStmtsInSchema(ctx, &testdataFolder, []string{"humanresources/create-tables.sql"}, "humanresources")
			if err != nil {
				t.Fatal(err)
			}
			err = postgres.Source.RunCreateStmtsInSchema(ctx, &testdataFolder, []string{"alltypes/create-tables.sql"}, "alltypes")
			if err != nil {
				t.Fatal(err)
			}
			err = postgres.Target.CreateSchemas(ctx, []string{"humanresources", "alltypes"})
			if err != nil {
				t.Fatal(err)
			}

			testlogger := testutil.GetTestLogger(t)
			cmdConfig := &cmdConfig{
				Source: &sourceConfig{
					ConnectionId: sourceConn.Id,
				},
				Destination: &sqlDestinationConfig{
					ConnectionUrl:        postgres.Target.URL,
					Driver:               postgresDriver,
					InitSchema:           true,
					TruncateBeforeInsert: true,
					TruncateCascade:      true,
				},
				OutputType: &outputType,
				AccountId:  &accountId,
			}
			sync := &clisync{
				connectiondataclient: conndataclient,
				connectionclient:     connclient,
				sqlmanagerclient:     sqlmanagerclient,
				ctx:                  ctx,
				logger:               testlogger,
				cmd:                  cmdConfig,
				connmanager:          connmanager,
				session:              connectionmanager.NewUniqueSession(),
			}
			err := sync.configureAndRunSync()
			require.NoError(t, err)

			rowCount, err := postgres.Target.GetTableRowCount(ctx, "humanresources", "employees")
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)

			rowCount, err = postgres.Target.GetTableRowCount(ctx, "humanresources", "generated_table")
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)

			rowCount, err = postgres.Target.GetTableRowCount(ctx, "alltypes", "all_data_types")
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)

			rowCount, err = postgres.Target.GetTableRowCount(ctx, "alltypes", "time_time")
			require.NoError(t, err)
			require.Greater(t, rowCount, 0)

			var tsvectorVal, jsonVal, jsonbVal string
			row := postgres.Target.DB.QueryRow(ctx, "select tsvector_col::text, json_col::text, jsonb_col::text from alltypes.all_data_types where tsvector_col is not null and json_col is not null;")
			err = row.Scan(&tsvectorVal, &jsonVal, &jsonbVal)
			require.NoError(t, err)
			require.Equal(t, "'example' 'tsvector'", tsvectorVal)
			require.Equal(t, `{"name": "John", "age": 30}`, jsonVal)
			require.Equal(t, `{"age": 30, "name": "John"}`, jsonbVal) // Note: JSONB reorders keys

			err = verifyPostgresTimeTimeTableValues(t, ctx, postgres.Target.DB, "alltypes")
			require.NoError(t, err)
		})

		t.Run("S3_end_to_end", func(t *testing.T) {
			t.Parallel()
			ok := testutil.ShouldRunS3IntegrationTest()
			if !ok {
				return
			}

			alltypesSchema := "alltypes_s3_pg"
			err := postgres.Source.RunCreateStmtsInSchema(ctx, &testdataFolder, []string{"alltypes/create-tables.sql"}, alltypesSchema)
			if err != nil {
				t.Fatal(err)
			}

			err = postgres.Target.RunCreateStmtsInSchema(ctx, &testdataFolder, []string{"alltypes/create-tables.sql"}, alltypesSchema)
			if err != nil {
				t.Fatal(err)
			}

			neosyncApi.MockTemporalForCreateJob("cli-test-sync")
			job, err := jobclient.CreateJob(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobRequest{
				AccountId: accountId,
				JobName:   "S3 to PG",
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
							Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
								ConnectionId:                  sourceConn.Id,
								Schemas:                       []*mgmtv1alpha1.PostgresSourceSchemaOption{},
								SubsetByForeignKeyConstraints: true,
							},
						},
					},
				},
				Destinations: []*mgmtv1alpha1.CreateJobDestination{
					{
						ConnectionId: s3Conn.Id,
						Options: &mgmtv1alpha1.JobDestinationOptions{
							Config: &mgmtv1alpha1.JobDestinationOptions_AwsS3Options{
								AwsS3Options: &mgmtv1alpha1.AwsS3DestinationConnectionOptions{},
							},
						},
					},
				},
				Mappings: pgalltypes.GetDefaultSyncJobMappings(alltypesSchema),
			}))
			require.NoError(t, err)

			t.Run("Postgres_to_S3", func(t *testing.T) {
				testworkflow := tcworkflow.NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
				testworkflow.RequireActivitiesCompletedSuccessfully(t)
				testworkflow.ExecuteTestDataSyncWorkflow(job.Msg.GetJob().GetId())
				require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: pg_to_s3")
				err = testworkflow.TestEnv.GetWorkflowError()
				require.NoError(t, err, "Received Temporal Workflow Error", "testName", "pg_to_s3")
			})

			t.Run("S3_to_Postgres", func(t *testing.T) {
				testlogger := testutil.GetTestLogger(t)
				cmdConfig := &cmdConfig{
					Source: &sourceConfig{
						ConnectionId: s3Conn.Id,
						ConnectionOpts: &connectionOpts{
							JobId: &job.Msg.Job.Id,
						},
					},
					Destination: &sqlDestinationConfig{
						ConnectionUrl:        postgres.Target.URL,
						Driver:               postgresDriver,
						InitSchema:           false,
						TruncateBeforeInsert: true,
						TruncateCascade:      true,
					},
					OutputType: &outputType,
					AccountId:  &accountId,
				}
				sync := &clisync{
					connectiondataclient: conndataclient,
					connectionclient:     connclient,
					sqlmanagerclient:     sqlmanagerclient,
					ctx:                  ctx,
					logger:               testlogger,
					cmd:                  cmdConfig,
					connmanager:          connmanager,
					session:              connectionmanager.NewUniqueSession(),
				}
				err := sync.configureAndRunSync()
				require.NoError(t, err)
			})

			rowCount, err := postgres.Target.GetTableRowCount(ctx, alltypesSchema, "all_data_types")
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)

			rowCount, err = postgres.Target.GetTableRowCount(ctx, alltypesSchema, "json_data")
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)

			rowCount, err = postgres.Target.GetTableRowCount(ctx, alltypesSchema, "time_time")
			require.NoError(t, err)
			require.Greater(t, rowCount, 0)

			var tsvectorVal, jsonVal, jsonbVal string
			row := postgres.Target.DB.QueryRow(ctx, fmt.Sprintf("select tsvector_col::text, json_col::text, jsonb_col::text from %s.all_data_types where tsvector_col is not null and json_col is not null;", alltypesSchema))
			err = row.Scan(&tsvectorVal, &jsonVal, &jsonbVal)
			require.NoError(t, err)
			require.Equal(t, "'example' 'tsvector'", tsvectorVal)
			require.Equal(t, `{"age":30,"name":"John"}`, jsonVal)
			require.Equal(t, `{"age": 30, "name": "John"}`, jsonbVal) // Note: JSONB reorders keys

			err = verifyPostgresTimeTimeTableValues(t, ctx, postgres.Target.DB, alltypesSchema)
			require.NoError(t, err)
		})

		t.Cleanup(func() {
			err := postgres.TearDown(ctx)
			if err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("mysql", func(t *testing.T) {
		t.Parallel()
		mysql, err := tcmysql.NewMysqlTestSyncContainer(ctx, []tcmysql.Option{}, []tcmysql.Option{})
		if err != nil {
			t.Fatal(err)
		}

		testdataFolder := "../../../../../internal/testutil/testdata/mysql"
		sourceConn := tcneosyncapi.CreateMysqlConnection(ctx, t, connclient, accountId, "mysql-source", mysql.Source.URL)

		t.Run("mysql_sync", func(t *testing.T) {
			// can't be run in parallel yet
			// right now CLI sync and init schema takes everything in source and copies it to target since there are no job mappings defined by the user
			// so it can't be scoped to specific schema
			// t.Parallel()
			err = mysql.Source.RunCreateStmtsInDatabase(ctx, &testdataFolder, []string{"humanresources/create-tables.sql"}, "humanresources")
			if err != nil {
				t.Fatal(err)
			}
			err = mysql.Source.RunCreateStmtsInDatabase(ctx, &testdataFolder, []string{"alltypes/create-tables.sql"}, "alltypes")
			if err != nil {
				t.Fatal(err)
			}
			err = mysql.Target.CreateDatabases(ctx, []string{"humanresources", "alltypes"})
			if err != nil {
				t.Fatal(err)
			}
			testlogger := testutil.GetTestLogger(t)
			cmdConfig := &cmdConfig{
				Source: &sourceConfig{
					ConnectionId: sourceConn.Id,
				},
				Destination: &sqlDestinationConfig{
					ConnectionUrl:        mysql.Target.URL,
					Driver:               mysqlDriver,
					InitSchema:           true,
					TruncateBeforeInsert: true,
				},
				OutputType: &outputType,
				AccountId:  &accountId,
			}
			sync := &clisync{
				connectiondataclient: conndataclient,
				connectionclient:     connclient,
				sqlmanagerclient:     sqlmanagerclient,
				ctx:                  ctx,
				logger:               testlogger,
				cmd:                  cmdConfig,
				connmanager:          connmanager,
				session:              connectionmanager.NewUniqueSession(),
			}
			err := sync.configureAndRunSync()
			require.NoError(t, err)

			rowCount, err := mysql.Target.GetTableRowCount(ctx, "humanresources", "locations")
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)

			rowCount, err = mysql.Target.GetTableRowCount(ctx, "humanresources", "generated_table")
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)

			rowCount, err = mysql.Target.GetTableRowCount(ctx, "alltypes", "all_data_types")
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)
		})

		t.Run("S3_end_to_end", func(t *testing.T) {
			t.Parallel()
			ok := testutil.ShouldRunS3IntegrationTest()
			if !ok {
				return
			}

			alltypesSchema := "alltypes_s3_mysql"
			err := mysql.Source.RunCreateStmtsInDatabase(ctx, &testdataFolder, []string{"alltypes/create-tables.sql"}, alltypesSchema)
			if err != nil {
				t.Fatal(err)
			}

			err = mysql.Target.RunCreateStmtsInDatabase(ctx, &testdataFolder, []string{"alltypes/create-tables.sql"}, alltypesSchema)
			if err != nil {
				t.Fatal(err)
			}

			neosyncApi.MockTemporalForCreateJob("cli-test-sync")
			job, err := jobclient.CreateJob(ctx, connect.NewRequest(&mgmtv1alpha1.CreateJobRequest{
				AccountId: accountId,
				JobName:   "S3 to Mysql",
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Mysql{
							Mysql: &mgmtv1alpha1.MysqlSourceConnectionOptions{
								ConnectionId:                  sourceConn.Id,
								Schemas:                       []*mgmtv1alpha1.MysqlSourceSchemaOption{},
								SubsetByForeignKeyConstraints: true,
							},
						},
					},
				},
				Destinations: []*mgmtv1alpha1.CreateJobDestination{
					{
						ConnectionId: s3Conn.Id,
						Options: &mgmtv1alpha1.JobDestinationOptions{
							Config: &mgmtv1alpha1.JobDestinationOptions_AwsS3Options{
								AwsS3Options: &mgmtv1alpha1.AwsS3DestinationConnectionOptions{},
							},
						},
					},
				},
				Mappings: mysqlalltypes.GetDefaultSyncJobMappings(alltypesSchema),
			}))
			require.NoError(t, err)

			t.Run("Mysql_to_S3", func(t *testing.T) {
				testworkflow := tcworkflow.NewTestDataSyncWorkflowEnv(t, neosyncApi, dbManagers)
				testworkflow.RequireActivitiesCompletedSuccessfully(t)
				testworkflow.ExecuteTestDataSyncWorkflow(job.Msg.GetJob().GetId())
				require.Truef(t, testworkflow.TestEnv.IsWorkflowCompleted(), "Workflow did not complete. Test: mysql_to_s3")
				err = testworkflow.TestEnv.GetWorkflowError()
				require.NoError(t, err, "Received Temporal Workflow Error", "testName", "mysql_to_s3")
			})

			t.Run("S3_to_Mysql", func(t *testing.T) {
				testlogger := testutil.GetTestLogger(t)
				cmdConfig := &cmdConfig{
					Source: &sourceConfig{
						ConnectionId: s3Conn.Id,
						ConnectionOpts: &connectionOpts{
							JobId: &job.Msg.Job.Id,
						},
					},
					Destination: &sqlDestinationConfig{
						ConnectionUrl:        mysql.Target.URL,
						Driver:               mysqlDriver,
						InitSchema:           false,
						TruncateBeforeInsert: true,
					},
					OutputType: &outputType,
					AccountId:  &accountId,
				}
				sync := &clisync{
					connectiondataclient: conndataclient,
					connectionclient:     connclient,
					sqlmanagerclient:     sqlmanagerclient,
					ctx:                  ctx,
					logger:               testlogger,
					cmd:                  cmdConfig,
					connmanager:          connmanager,
					session:              connectionmanager.NewUniqueSession(),
				}
				err := sync.configureAndRunSync()
				require.NoError(t, err)
			})

			rowCount, err := mysql.Target.GetTableRowCount(ctx, alltypesSchema, "all_data_types")
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)

			rowCount, err = mysql.Target.GetTableRowCount(ctx, alltypesSchema, "json_data")
			require.NoError(t, err)
			require.Greater(t, rowCount, 1)
		})

		t.Cleanup(func() {
			err := mysql.TearDown(ctx)
			if err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Cleanup(func() {
		err = neosyncApi.TearDown(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func verifyPostgresTimeTimeTableValues(t *testing.T, ctx context.Context, db *pgxpool.Pool, schema string) error {
	rows, err := db.Query(ctx, fmt.Sprintf("select timestamp_col::text, date_col::text from %q.time_time;", schema))
	if err != nil {
		return err
	}
	defer rows.Close()

	expectedTimestamps := [][]byte{
		[]byte("2024-03-18 10:30:00"),
		[]byte("0001-01-01 00:00:00 BC"),
		[]byte("0002-01-01 00:00:00 BC"),
	}
	expectedDates := [][]byte{
		[]byte("2024-03-18"),
		[]byte("0001-01-01 BC"),
		[]byte("0002-01-01 BC"),
	}
	var actualTimestamps [][]byte
	var actualDates [][]byte

	for rows.Next() {
		var timestampCol, dateCol []byte
		err = rows.Scan(&timestampCol, &dateCol)
		if err != nil {
			return err
		}
		actualTimestamps = append(actualTimestamps, timestampCol)
		actualDates = append(actualDates, dateCol)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	require.ElementsMatch(t, expectedTimestamps, actualTimestamps, "Expected timestamp_col values to match")
	require.ElementsMatch(t, expectedDates, actualDates, "Expected date_col values to match")
	return nil
}
