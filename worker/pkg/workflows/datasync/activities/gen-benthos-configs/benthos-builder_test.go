package genbenthosconfigs_activity

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/benthosdev/benthos/v4/public/service"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	_ "github.com/benthosdev/benthos/v4/public/components/aws"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	_ "github.com/benthosdev/benthos/v4/public/components/javascript"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/benthosdev/benthos/v4/public/components/pure/extended"
	_ "github.com/benthosdev/benthos/v4/public/components/redis"
	_ "github.com/benthosdev/benthos/v4/public/components/sql"
	neosync_benthos_error "github.com/nucleuscloud/neosync/worker/internal/benthos/error"
	benthos_metrics "github.com/nucleuscloud/neosync/worker/internal/benthos/metrics"
	_ "github.com/nucleuscloud/neosync/worker/internal/benthos/redis"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/internal/benthos/sql"
	_ "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers"

	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
)

const (
	mockJobId = "b1767636-3992-4cb4-9bf2-4bb9bddbf43c"
	mockRunId = "26444272-0bb0-4325-ae60-17dcd9744785"
)

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Generate_Pg(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Generate{
							Generate: &mgmtv1alpha1.GenerateSourceOptions{
								Schemas: []*mgmtv1alpha1.GenerateSourceSchemaOption{
									{
										Schema: "public",
										Tables: []*mgmtv1alpha1.GenerateSourceTableOption{
											{
												Table:    "users",
												RowCount: 10,
											},
										},
									},
								},
							},
						},
					},
				},
				Mappings: []*mgmtv1alpha1.JobMapping{
					{
						Schema: "public",
						Table:  "users",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UUID,
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
										IncludeHyphens: true,
									},
								},
							},
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SSN,
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
									GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
								},
							},
						},
					},
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
						Options: &mgmtv1alpha1.JobDestinationOptions{
							Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
								PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
									TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
										TruncateBeforeInsert: true,
									},
									OnConflict: &mgmtv1alpha1.PostgresOnConflictConfig{
										DoNothing: true,
									},
								},
							},
						},
					},
				},
			},
		}), nil)

	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "456",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "456",
			Name: "stage",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-stage-url",
						},
					},
				},
			},
		},
	}), nil)

	bbuilder := newBenthosBuilder(mockSqlManager, mockJobClient, mockConnectionClient, mockTransformerClient, mockJobId, mockRunId, nil, false)
	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	require.Nil(t, err)
	require.NotEmpty(t, resp.BenthosConfigs)
	require.Len(t, resp.BenthosConfigs, 1)
	bc := resp.BenthosConfigs[0]
	require.Equal(t, bc.Name, "public.users")
	require.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(`
input:
    label: ""
    generate:
        mapping: root = {}
        interval: ""
        count: 10
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - retry:
                    output:
                        label: ""
                        pooled_sql_insert:
                            driver: postgres
                            dsn: ${DESTINATION_0_CONNECTION_DSN}
                            schema: public
                            table: users
                            columns:
                                - id
                                - name
                            on_conflict_do_nothing: true
                            truncate_on_retry: true
                            args_mapping: root = [this."id", this."name"]
                            batching:
                                count: 100
                                byte_size: 0
                                period: 5s
                                check: ""
                                processors: []
                        processors:
                            - mutation: |-
                                root."id" = generate_uuid(include_hyphens:true)
                                root."name" = generate_ssn()
                            - catch:
                                - error:
                                    error_msg: ${! error()}
                    max_retries: 10
                    backoff: {}
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
		strings.TrimSpace(string(out)),
	)

	// create a new streambuilder instance so we can access the SetYaml method
	benthosenv := service.NewEnvironment()
	err = neosync_benthos_sql.RegisterPooledSqlInsertOutput(benthosenv, nil, false)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlRawInput(benthosenv, nil, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorProcessor(benthosenv, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorOutput(benthosenv, nil)
	require.NoError(t, err)
	newSB := benthosenv.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	require.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Metrics(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				AccountId: "test-account-id",
				Id:        "test-job-id",
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Generate{
							Generate: &mgmtv1alpha1.GenerateSourceOptions{
								Schemas: []*mgmtv1alpha1.GenerateSourceSchemaOption{
									{
										Schema: "public",
										Tables: []*mgmtv1alpha1.GenerateSourceTableOption{
											{
												Table:    "users",
												RowCount: 10,
											},
										},
									},
								},
							},
						},
					},
				},
				Mappings: []*mgmtv1alpha1.JobMapping{
					{
						Schema: "public",
						Table:  "users",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UUID,
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
										IncludeHyphens: true,
									},
								},
							},
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SSN,
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
									GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
								},
							},
						},
					},
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
					},
				},
			},
		}), nil)

	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "456",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "456",
			Name: "stage",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-stage-url",
						},
					},
				},
			},
		},
	}), nil)

	bbuilder := newBenthosBuilder(mockSqlManager, mockJobClient, mockConnectionClient, mockTransformerClient, mockJobId, mockRunId, nil, true)
	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	require.Nil(t, err)
	require.NotEmpty(t, resp.BenthosConfigs)
	require.Len(t, resp.BenthosConfigs, 1)
	bc := resp.BenthosConfigs[0]
	require.Equal(t, bc.Name, "public.users")
	require.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(`
input:
    label: ""
    generate:
        mapping: root = {}
        interval: ""
        count: 10
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - retry:
                    output:
                        label: ""
                        pooled_sql_insert:
                            driver: postgres
                            dsn: ${DESTINATION_0_CONNECTION_DSN}
                            schema: public
                            table: users
                            columns:
                                - id
                                - name
                            on_conflict_do_nothing: false
                            truncate_on_retry: false
                            args_mapping: root = [this."id", this."name"]
                            batching:
                                count: 100
                                byte_size: 0
                                period: 5s
                                check: ""
                                processors: []
                        processors:
                            - mutation: |-
                                root."id" = generate_uuid(include_hyphens:true)
                                root."name" = generate_ssn()
                            - catch:
                                - error:
                                    error_msg: ${! error()}
                    max_retries: 10
                    backoff: {}
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
metrics:
    otel_collector: {}
    mapping: |-
        meta neosyncAccountId = "test-account-id"
        meta neosyncJobId = "test-job-id"
        meta temporalWorkflowId = "${TEMPORAL_WORKFLOW_ID}"
        meta temporalRunId = "${TEMPORAL_RUN_ID}"
        meta tableSchema = "public"
        meta tableName = "users"
        meta jobType = "generate"
`),
		strings.TrimSpace(string(out)),
	)

	// create a new streambuilder instance so we can access the SetYaml method
	benthosenv := service.NewEnvironment()
	err = neosync_benthos_sql.RegisterPooledSqlInsertOutput(benthosenv, nil, false)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlRawInput(benthosenv, nil, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorProcessor(benthosenv, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorOutput(benthosenv, nil)
	require.NoError(t, err)
	err = benthos_metrics.RegisterOtelMetricsExporter(benthosenv, nil)
	require.NoError(t, err)
	newSB := benthosenv.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	require.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Generate_Pg_Pg(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Generate{
							Generate: &mgmtv1alpha1.GenerateSourceOptions{
								Schemas: []*mgmtv1alpha1.GenerateSourceSchemaOption{
									{
										Schema: "public",
										Tables: []*mgmtv1alpha1.GenerateSourceTableOption{
											{
												Table:    "users",
												RowCount: 10,
											},
										},
									},
								},
							},
						},
					},
				},
				Mappings: []*mgmtv1alpha1.JobMapping{
					{
						Schema: "public",
						Table:  "users",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{
									GenerateDefaultConfig: &mgmtv1alpha1.GenerateDefault{},
								},
							},
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SSN,
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
									GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
								},
							},
						},
					},
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
					},
				},
			},
		}), nil)

	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "456",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "456",
			Name: "stage",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-stage-url",
						},
					},
				},
			},
		},
	}), nil)

	bbuilder := newBenthosBuilder(mockSqlManager, mockJobClient, mockConnectionClient, mockTransformerClient, mockJobId, mockRunId, nil, false)
	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	require.Nil(t, err)
	require.NotEmpty(t, resp.BenthosConfigs)
	require.Len(t, resp.BenthosConfigs, 1)
	bc := resp.BenthosConfigs[0]
	require.Equal(t, bc.Name, "public.users")
	require.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    generate:
        mapping: root = {}
        interval: ""
        count: 10
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - retry:
                    output:
                        label: ""
                        pooled_sql_insert:
                            driver: postgres
                            dsn: ${DESTINATION_0_CONNECTION_DSN}
                            schema: public
                            table: users
                            columns:
                                - id
                                - name
                            on_conflict_do_nothing: false
                            truncate_on_retry: false
                            args_mapping: root = [this."id", this."name"]
                            batching:
                                count: 100
                                byte_size: 0
                                period: 5s
                                check: ""
                                processors: []
                        processors:
                            - mutation: |-
                                root."id" = "DEFAULT"
                                root."name" = generate_ssn()
                            - catch:
                                - error:
                                    error_msg: ${! error()}
                    max_retries: 10
                    backoff: {}
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	// create a new streambuilder instance so we can access the SetYaml method
	benthosenv := service.NewEnvironment()
	err = neosync_benthos_sql.RegisterPooledSqlInsertOutput(benthosenv, nil, false)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlRawInput(benthosenv, nil, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorProcessor(benthosenv, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorOutput(benthosenv, nil)
	require.NoError(t, err)
	newSB := benthosenv.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	require.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_PrimaryKey_Transformer_Pg_Pg(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

	redisConfig := &shared.RedisConfig{
		Url:  "redis://localhost:6379",
		Kind: "simple",
	}

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
							Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
								ConnectionId: "123",
							},
						},
					},
				},
				Mappings: []*mgmtv1alpha1.JobMapping{
					{
						Schema: "public",
						Table:  "users",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UUID,
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
										IncludeHyphens: true,
									},
								},
							},
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "orders",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "orders",
						Column: "buyer_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
					},
				},
			},
		}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "123",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "123",
			Name: "prod",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-prod-url",
						},
					},
				},
			},
		},
	}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "456",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "456",
			Name: "stage",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-stage-url",
						},
					},
				},
			},
		},
	}), nil)
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb, Driver: sql_manager.PostgresDriver}, nil)
	mockSqlDb.On("Close").Return(nil)
	mockSqlDb.On("GetSchemaColumnMap", mock.Anything).Return(map[string]map[string]*sql_manager.ColumnInfo{
		"public.users":  {"id": &sql_manager.ColumnInfo{}, "name": &sql_manager.ColumnInfo{}},
		"public.orders": {"id": &sql_manager.ColumnInfo{}, "buyer_id": &sql_manager.ColumnInfo{}},
	}, nil)
	mockSqlDb.On("GetPrimaryKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]string{
		"public.users":  {"id"},
		"public.orders": {"id"},
	}, nil)
	mockSqlDb.On("GetForeignKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]*sql_manager.ForeignConstraint{
		"public.orders": {{Columns: []string{"buyer_id"}, NotNullable: []bool{true}, ForeignKey: &sql_manager.ForeignKey{Table: "public.users", Columns: []string{"id"}}}},
	}, nil)
	bbuilder := newBenthosBuilder(mockSqlManager, mockJobClient, mockConnectionClient, mockTransformerClient, mockJobId, mockRunId, redisConfig, false)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)

	require.NoError(t, err)
	require.NotEmpty(t, resp.BenthosConfigs)
	require.Len(t, resp.BenthosConfigs, 2)
	bc := getBenthosConfigByName(resp.BenthosConfigs, "public.users.insert")
	require.Equal(t, bc.Name, "public.users.insert")
	require.Len(t, bc.RedisConfig, 1)
	require.Equal(t, bc.RedisConfig[0].Table, "public.users")
	require.Equal(t, bc.RedisConfig[0].Column, "id")
	out, err := yaml.Marshal(bc.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        query: SELECT "id", "name" FROM "public"."users";
pipeline:
    threads: -1
    processors:
        - mapping: meta neosync_public_users_id = this."id"
        - mutation: root."id" = generate_uuid(include_hyphens:true)
        - catch:
            - error:
                error_msg: ${! error()}
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - redis_hash_output:
                url: redis://localhost:6379
                key: 03d74b3f8146d46519bb4cb56b2f7b327e603540eb4d96a0feb9acba88a4d79d
                walk_metadata: false
                walk_json_object: false
                fields_mapping: 'root = {meta("neosync_public_users_id"): json("id")}'
                kind: simple
            - fallback:
                - pooled_sql_insert:
                    driver: postgres
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: users
                    columns:
                        - id
                        - name
                    on_conflict_do_nothing: false
                    truncate_on_retry: false
                    args_mapping: root = [this."id", this."name"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	bc = getBenthosConfigByName(resp.BenthosConfigs, "public.orders.insert")
	require.Equal(t, bc.Name, "public.orders.insert")
	require.Empty(t, bc.RedisConfig)
	out, err = yaml.Marshal(bc.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        query: SELECT "id", "buyer_id" FROM "public"."orders";
pipeline:
    threads: -1
    processors:
        - branch:
            processors:
                - redis:
                    url: redis://localhost:6379
                    command: hget
                    args_mapping: root = ["03d74b3f8146d46519bb4cb56b2f7b327e603540eb4d96a0feb9acba88a4d79d", json("buyer_id")]
                    kind: simple
            request_map: root = if this."buyer_id" == null { deleted() } else { this }
            result_map: root."buyer_id" = this
        - catch:
            - error:
                error_msg: ${! error()}
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - pooled_sql_insert:
                    driver: postgres
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: orders
                    columns:
                        - id
                        - buyer_id
                    on_conflict_do_nothing: false
                    truncate_on_retry: false
                    args_mapping: root = [this."id", this."buyer_id"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	benthosenv := service.NewEnvironment()
	err = neosync_benthos_sql.RegisterPooledSqlInsertOutput(benthosenv, nil, false)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlRawInput(benthosenv, nil, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorProcessor(benthosenv, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorOutput(benthosenv, nil)
	require.NoError(t, err)

	newSB := benthosenv.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	require.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_PrimaryKey_Passthrough_Pg_Pg(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
							Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
								ConnectionId: "123",
							},
						},
					},
				},
				Mappings: []*mgmtv1alpha1.JobMapping{
					{
						Schema: "public",
						Table:  "users",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "orders",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "orders",
						Column: "buyer_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
					},
				},
			},
		}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "123",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "123",
			Name: "prod",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-prod-url",
						},
					},
				},
			},
		},
	}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "456",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "456",
			Name: "stage",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-stage-url",
						},
					},
				},
			},
		},
	}), nil)
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb, Driver: sql_manager.PostgresDriver}, nil)
	mockSqlDb.On("Close").Return(nil)
	mockSqlDb.On("GetSchemaColumnMap", mock.Anything).Return(map[string]map[string]*sql_manager.ColumnInfo{
		"public.users":  {"id": &sql_manager.ColumnInfo{}, "name": &sql_manager.ColumnInfo{}},
		"public.orders": {"id": &sql_manager.ColumnInfo{}, "buyer_id": &sql_manager.ColumnInfo{}},
	}, nil)
	mockSqlDb.On("GetPrimaryKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]string{
		"public.users":  {"id"},
		"public.orders": {"id"},
	}, nil)
	mockSqlDb.On("GetForeignKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]*sql_manager.ForeignConstraint{
		"public.orders": {{Columns: []string{"buyer_id"}, NotNullable: []bool{true}, ForeignKey: &sql_manager.ForeignKey{Table: "public.users", Columns: []string{"id"}}}},
	}, nil)

	bbuilder := newBenthosBuilder(mockSqlManager, mockJobClient, mockConnectionClient, mockTransformerClient, mockJobId, mockRunId, nil, false)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)

	require.Nil(t, err)
	require.NotEmpty(t, resp.BenthosConfigs)
	require.Len(t, resp.BenthosConfigs, 2)
	bc := getBenthosConfigByName(resp.BenthosConfigs, "public.users.insert")
	require.Equal(t, bc.Name, "public.users.insert")
	require.Empty(t, bc.RedisConfig)
	out, err := yaml.Marshal(bc.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        query: SELECT "id", "name" FROM "public"."users";
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - pooled_sql_insert:
                    driver: postgres
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: users
                    columns:
                        - id
                        - name
                    on_conflict_do_nothing: false
                    truncate_on_retry: false
                    args_mapping: root = [this."id", this."name"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	bc = getBenthosConfigByName(resp.BenthosConfigs, "public.orders.insert")
	require.Equal(t, bc.Name, "public.orders.insert")
	require.Empty(t, bc.RedisConfig)
	out, err = yaml.Marshal(bc.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        query: SELECT "id", "buyer_id" FROM "public"."orders";
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - pooled_sql_insert:
                    driver: postgres
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: orders
                    columns:
                        - id
                        - buyer_id
                    on_conflict_do_nothing: false
                    truncate_on_retry: false
                    args_mapping: root = [this."id", this."buyer_id"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	benthosenv := service.NewEnvironment()
	err = neosync_benthos_sql.RegisterPooledSqlInsertOutput(benthosenv, nil, false)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlRawInput(benthosenv, nil, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorOutput(benthosenv, nil)
	require.NoError(t, err)
	newSB := benthosenv.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	require.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_CircularDependency_PrimaryKey_Transformer_Pg_Pg(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

	redisConfig := &shared.RedisConfig{
		Url:  "redis://localhost:6379",
		Kind: "simple",
	}

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
							Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
								ConnectionId: "123",
							},
						},
					},
				},
				Mappings: []*mgmtv1alpha1.JobMapping{
					{
						Schema: "public",
						Table:  "jobs",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UUID,
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
										IncludeHyphens: true,
									},
								},
							},
						},
					},
					{
						Schema: "public",
						Table:  "jobs",
						Column: "parent_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
					},
				},
			},
		}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "123",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "123",
			Name: "prod",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-prod-url",
						},
					},
				},
			},
		},
	}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "456",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "456",
			Name: "stage",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-stage-url",
						},
					},
				},
			},
		},
	}), nil)

	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb, Driver: sql_manager.PostgresDriver}, nil)
	mockSqlDb.On("Close").Return(nil)
	mockSqlDb.On("GetSchemaColumnMap", mock.Anything).Return(map[string]map[string]*sql_manager.ColumnInfo{
		"public.jobs": {"id": &sql_manager.ColumnInfo{}, "parent_id": &sql_manager.ColumnInfo{}},
	}, nil)
	mockSqlDb.On("GetPrimaryKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]string{
		"public.jobs": {"id"},
	}, nil)
	mockSqlDb.On("GetForeignKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]*sql_manager.ForeignConstraint{
		"public.jobs": {{Columns: []string{"parent_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.jobs", Columns: []string{"id"}}}},
	}, nil)

	bbuilder := newBenthosBuilder(mockSqlManager, mockJobClient, mockConnectionClient, mockTransformerClient, mockJobId, mockRunId, redisConfig, false)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)

	require.NoError(t, err)
	require.NotEmpty(t, resp.BenthosConfigs)
	require.Len(t, resp.BenthosConfigs, 2)
	bc := getBenthosConfigByName(resp.BenthosConfigs, "public.jobs.insert")
	require.Equal(t, bc.Name, "public.jobs.insert")
	require.Len(t, bc.RedisConfig, 1)
	require.Equal(t, bc.RedisConfig[0].Table, "public.jobs")
	require.Equal(t, bc.RedisConfig[0].Column, "id")
	out, err := yaml.Marshal(bc.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        query: SELECT "id", "parent_id" FROM "public"."jobs";
pipeline:
    threads: -1
    processors:
        - mapping: meta neosync_public_jobs_id = this."id"
        - mutation: root."id" = generate_uuid(include_hyphens:true)
        - catch:
            - error:
                error_msg: ${! error()}
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - redis_hash_output:
                url: redis://localhost:6379
                key: 60c88959dfc6d1e114ca65511290b36786e08961ee1acc3c2b605e1468b561cb
                walk_metadata: false
                walk_json_object: false
                fields_mapping: 'root = {meta("neosync_public_jobs_id"): json("id")}'
                kind: simple
            - fallback:
                - pooled_sql_insert:
                    driver: postgres
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: jobs
                    columns:
                        - id
                    on_conflict_do_nothing: false
                    truncate_on_retry: false
                    args_mapping: root = [this."id"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	bc = getBenthosConfigByName(resp.BenthosConfigs, "public.jobs.update")
	require.Equal(t, bc.Name, "public.jobs.update")
	require.Empty(t, bc.RedisConfig)
	out, err = yaml.Marshal(bc.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        query: SELECT "id", "parent_id" FROM "public"."jobs";
pipeline:
    threads: -1
    processors:
        - branch:
            processors:
                - redis:
                    url: redis://localhost:6379
                    command: hget
                    args_mapping: root = ["60c88959dfc6d1e114ca65511290b36786e08961ee1acc3c2b605e1468b561cb", json("parent_id")]
                    kind: simple
            request_map: root = if this."parent_id" == null { deleted() } else { this }
            result_map: root."parent_id" = this
        - branch:
            processors:
                - redis:
                    url: redis://localhost:6379
                    command: hget
                    args_mapping: root = ["60c88959dfc6d1e114ca65511290b36786e08961ee1acc3c2b605e1468b561cb", json("id")]
                    kind: simple
            request_map: root = if this."id" == null { deleted() } else { this }
            result_map: root."id" = this
        - catch:
            - error:
                error_msg: ${! error()}
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - pooled_sql_update:
                    driver: postgres
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: jobs
                    columns:
                        - parent_id
                    where_columns:
                        - id
                    args_mapping: root = [this."parent_id", this."id"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []

`),
	)

	benthosenv := service.NewEnvironment()
	err = neosync_benthos_sql.RegisterPooledSqlInsertOutput(benthosenv, nil, false)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlUpdateOutput(benthosenv, nil)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlRawInput(benthosenv, nil, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorProcessor(benthosenv, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorOutput(benthosenv, nil)
	require.NoError(t, err)
	newSB := benthosenv.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	require.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Pg_Pg_With_Constraints(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
							Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
								ConnectionId: "123",
							},
						},
					},
				},
				Mappings: []*mgmtv1alpha1.JobMapping{
					{
						Schema: "public",
						Table:  "users",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "user_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
					},
				},
			},
		}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "123",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "123",
			Name: "prod",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-prod-url",
						},
					},
				},
			},
		},
	}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "456",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "456",
			Name: "stage",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-stage-url",
						},
					},
				},
			},
		},
	}), nil)

	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb, Driver: sql_manager.PostgresDriver}, nil)
	mockSqlDb.On("Close").Return(nil)
	mockSqlDb.On("GetSchemaColumnMap", mock.Anything).Return(map[string]map[string]*sql_manager.ColumnInfo{
		"public.users":                     {"id": &sql_manager.ColumnInfo{}, "name": &sql_manager.ColumnInfo{}},
		"public.user_account_associations": {"id": &sql_manager.ColumnInfo{}, "user_id": &sql_manager.ColumnInfo{}},
	}, nil)
	mockSqlDb.On("GetPrimaryKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]string{
		"public.users":                     {"id"},
		"public.user_account_associations": {"id"},
	}, nil)
	mockSqlDb.On("GetForeignKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]*sql_manager.ForeignConstraint{
		"public.user_account_associations": {{Columns: []string{"user_id"}, NotNullable: []bool{true}, ForeignKey: &sql_manager.ForeignKey{Table: "public.users", Columns: []string{"id"}}}},
	}, nil)

	bbuilder := newBenthosBuilder(mockSqlManager, mockJobClient, mockConnectionClient, mockTransformerClient, mockJobId, mockRunId, nil, false)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	require.Nil(t, err)
	require.NotEmpty(t, resp.BenthosConfigs)
	require.Len(t, resp.BenthosConfigs, 2)

	bc := getBenthosConfigByName(resp.BenthosConfigs, "public.users.insert")
	require.NotNil(t, bc)
	require.Equal(t, bc.Name, "public.users.insert")
	require.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	require.NoError(t, err)

	bc2 := getBenthosConfigByName(resp.BenthosConfigs, "public.user_account_associations.insert")
	require.Equal(t, bc2.Name, "public.user_account_associations.insert")
	require.Equal(t, bc2.DependsOn, []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}})
	out2, err := yaml.Marshal(bc2.Config)
	require.NoError(t, err)

	benthosenv := service.NewEnvironment()
	err = neosync_benthos_sql.RegisterPooledSqlInsertOutput(benthosenv, nil, false)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlRawInput(benthosenv, nil, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorOutput(benthosenv, nil)
	require.NoError(t, err)
	newSB := benthosenv.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	require.NoError(t, err)

	err = newSB.SetYAML(string(out2))
	require.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Pg_Pg_With_Circular_Dependency(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
							Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
								ConnectionId: "123",
							},
						},
					},
				},
				Mappings: []*mgmtv1alpha1.JobMapping{
					{
						Schema: "public",
						Table:  "users",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "user_assoc_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "user_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
					},
				},
			},
		}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "123",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "123",
			Name: "prod",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-prod-url",
						},
					},
				},
			},
		},
	}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "456",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "456",
			Name: "stage",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-stage-url",
						},
					},
				},
			},
		},
	}), nil)
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb, Driver: sql_manager.PostgresDriver}, nil)
	mockSqlDb.On("Close").Return(nil)
	mockSqlDb.On("GetSchemaColumnMap", mock.Anything).Return(map[string]map[string]*sql_manager.ColumnInfo{
		"public.users":                     {"id": &sql_manager.ColumnInfo{}, "user_assoc_id": &sql_manager.ColumnInfo{}, "name": &sql_manager.ColumnInfo{}},
		"public.user_account_associations": {"id": &sql_manager.ColumnInfo{}, "user_id": &sql_manager.ColumnInfo{}},
	}, nil)
	mockSqlDb.On("GetPrimaryKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]string{
		"public.users":                     {"id"},
		"public.user_account_associations": {"id"},
	}, nil)
	mockSqlDb.On("GetForeignKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]*sql_manager.ForeignConstraint{
		"public.user_account_associations": {{Columns: []string{"user_id"}, NotNullable: []bool{true}, ForeignKey: &sql_manager.ForeignKey{Table: "public.users", Columns: []string{"id"}}}},
		"public.users":                     {{Columns: []string{"user_assoc_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.user_account_associations", Columns: []string{"id"}}}},
	}, nil)

	bbuilder := newBenthosBuilder(mockSqlManager, mockJobClient, mockConnectionClient, mockTransformerClient, mockJobId, mockRunId, nil, false)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	require.Nil(t, err)
	require.NotEmpty(t, resp.BenthosConfigs)
	require.Len(t, resp.BenthosConfigs, 3)

	insertConfig := getBenthosConfigByName(resp.BenthosConfigs, "public.users.insert")
	require.NotNil(t, insertConfig)
	require.Equal(t, insertConfig.Name, "public.users.insert")
	require.Empty(t, insertConfig.DependsOn)
	out, err := yaml.Marshal(insertConfig.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        query: SELECT "id", "name", "user_assoc_id" FROM "public"."users";
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - pooled_sql_insert:
                    driver: postgres
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: users
                    columns:
                        - id
                        - name
                    on_conflict_do_nothing: false
                    truncate_on_retry: false
                    args_mapping: root = [this."id", this."name"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	updateConfig := getBenthosConfigByName(resp.BenthosConfigs, "public.users.update")
	require.NotNil(t, updateConfig)
	require.Equal(t, updateConfig.Name, "public.users.update")
	require.Equal(t, updateConfig.DependsOn, []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.user_account_associations", Columns: []string{"id"}}})
	out1, err := yaml.Marshal(updateConfig.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out1)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        query: SELECT "id", "name", "user_assoc_id" FROM "public"."users";
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - pooled_sql_update:
                    driver: postgres
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: users
                    columns:
                        - user_assoc_id
                    where_columns:
                        - id
                    args_mapping: root = [this."user_assoc_id", this."id"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	bc2 := getBenthosConfigByName(resp.BenthosConfigs, "public.user_account_associations.insert")
	require.Equal(t, bc2.Name, "public.user_account_associations.insert")
	require.Equal(t, bc2.DependsOn, []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}})
	out2, err := yaml.Marshal(bc2.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out2)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        query: SELECT "id", "user_id" FROM "public"."user_account_associations";
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - pooled_sql_insert:
                    driver: postgres
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: user_account_associations
                    columns:
                        - id
                        - user_id
                    on_conflict_do_nothing: false
                    truncate_on_retry: false
                    args_mapping: root = [this."id", this."user_id"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	benthosenv := service.NewEnvironment()
	err = neosync_benthos_sql.RegisterPooledSqlInsertOutput(benthosenv, nil, false)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlUpdateOutput(benthosenv, nil)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlRawInput(benthosenv, nil, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorOutput(benthosenv, nil)
	require.NoError(t, err)
	newSB := benthosenv.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	require.NoError(t, err)

	err = newSB.SetYAML(string(out2))
	require.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Pg_Pg_With_Circular_Dependency_S3(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
							Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
								ConnectionId: "123",
							},
						},
					},
				},
				Mappings: []*mgmtv1alpha1.JobMapping{
					{
						Schema: "public",
						Table:  "users",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "user_assoc_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "user_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
					},
					{
						ConnectionId: "789",
					},
				},
			},
		}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "123",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "123",
			Name: "prod",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-prod-url",
						},
					},
				},
			},
		},
	}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "456",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "456",
			Name: "stage",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: "fake-stage-url",
						},
					},
				},
			},
		},
	}), nil)
	region := "us-west-2"
	accessId := "access-key"
	secret := "secret"
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "789",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "789",
			Name: "s3",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_AwsS3Config{
					AwsS3Config: &mgmtv1alpha1.AwsS3ConnectionConfig{
						Bucket: "s3-bucket",
						Region: &region,
						Credentials: &mgmtv1alpha1.AwsS3Credentials{
							AccessKeyId:     &accessId,
							SecretAccessKey: &secret,
						},
					},
				},
			},
		},
	}), nil)
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb, Driver: sql_manager.PostgresDriver}, nil)
	mockSqlDb.On("Close").Return(nil)
	mockSqlDb.On("GetSchemaColumnMap", mock.Anything).Return(map[string]map[string]*sql_manager.ColumnInfo{
		"public.users":                     {"id": &sql_manager.ColumnInfo{}, "user_assoc_id": &sql_manager.ColumnInfo{}, "name": &sql_manager.ColumnInfo{}},
		"public.user_account_associations": {"id": &sql_manager.ColumnInfo{}, "user_id": &sql_manager.ColumnInfo{}},
	}, nil)

	mockSqlDb.On("GetPrimaryKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]string{
		"public.users":                     {"id"},
		"public.user_account_associations": {"id"},
	}, nil)
	mockSqlDb.On("GetForeignKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]*sql_manager.ForeignConstraint{
		"public.user_account_associations": {{Columns: []string{"user_id"}, NotNullable: []bool{true}, ForeignKey: &sql_manager.ForeignKey{Table: "public.users", Columns: []string{"id"}}}},
		"public.users":                     {{Columns: []string{"user_assoc_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.user_account_associations", Columns: []string{"id"}}}},
	}, nil)

	bbuilder := newBenthosBuilder(mockSqlManager, mockJobClient, mockConnectionClient, mockTransformerClient, mockJobId, mockRunId, nil, false)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	require.Nil(t, err)
	require.NotEmpty(t, resp.BenthosConfigs)
	require.Len(t, resp.BenthosConfigs, 3)

	insertConfig := getBenthosConfigByName(resp.BenthosConfigs, "public.users.insert")
	require.NotNil(t, insertConfig)
	require.Equal(t, insertConfig.Name, "public.users.insert")
	require.Empty(t, insertConfig.DependsOn)
	out, err := yaml.Marshal(insertConfig.Config)

	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        query: SELECT "id", "name", "user_assoc_id" FROM "public"."users";
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - pooled_sql_insert:
                    driver: postgres
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: users
                    columns:
                        - id
                        - name
                    on_conflict_do_nothing: false
                    truncate_on_retry: false
                    args_mapping: root = [this."id", this."name"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
            - fallback:
                - aws_s3:
                    bucket: s3-bucket
                    max_in_flight: 64
                    path: /workflows/123/activities/public.users/data/${!count("files")}.txt.gz
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors:
                            - archive:
                                format: lines
                            - compress:
                                algorithm: gzip
                    region: us-west-2
                    credentials:
                        id: access-key
                        secret: secret
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	updateConfig := getBenthosConfigByName(resp.BenthosConfigs, "public.users.update")
	require.NotNil(t, updateConfig)
	require.Equal(t, updateConfig.Name, "public.users.update")
	require.Equal(t, updateConfig.DependsOn, []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.user_account_associations", Columns: []string{"id"}}})
	out1, err := yaml.Marshal(updateConfig.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out1)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        query: SELECT "id", "name", "user_assoc_id" FROM "public"."users";
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - pooled_sql_update:
                    driver: postgres
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: users
                    columns:
                        - user_assoc_id
                    where_columns:
                        - id
                    args_mapping: root = [this."user_assoc_id", this."id"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	bc2 := getBenthosConfigByName(resp.BenthosConfigs, "public.user_account_associations.insert")
	require.Equal(t, bc2.Name, "public.user_account_associations.insert")
	require.Equal(t, bc2.DependsOn, []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}})
	out2, err := yaml.Marshal(bc2.Config)

	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out2)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        query: SELECT "id", "user_id" FROM "public"."user_account_associations";
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - pooled_sql_insert:
                    driver: postgres
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: user_account_associations
                    columns:
                        - id
                        - user_id
                    on_conflict_do_nothing: false
                    truncate_on_retry: false
                    args_mapping: root = [this."id", this."user_id"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
            - fallback:
                - aws_s3:
                    bucket: s3-bucket
                    max_in_flight: 64
                    path: /workflows/123/activities/public.user_account_associations/data/${!count("files")}.txt.gz
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors:
                            - archive:
                                format: lines
                            - compress:
                                algorithm: gzip
                    region: us-west-2
                    credentials:
                        id: access-key
                        secret: secret
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	benthosenv := service.NewEnvironment()
	err = neosync_benthos_sql.RegisterPooledSqlInsertOutput(benthosenv, nil, false)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlRawInput(benthosenv, nil, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorOutput(benthosenv, nil)
	require.NoError(t, err)
	newSB := benthosenv.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	require.NoError(t, err)

	err = newSB.SetYAML(string(out2))
	require.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Mysql_Mysql(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformersClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Mysql{
							Mysql: &mgmtv1alpha1.MysqlSourceConnectionOptions{
								ConnectionId: "123",
							},
						},
					},
				},
				Mappings: []*mgmtv1alpha1.JobMapping{
					{
						Schema: "public",
						Table:  "users",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "user_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
					},
				},
			},
		}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "123",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "123",
			Name: "prod",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
							Url: "fake-prod-url",
						},
					},
				},
			},
		},
	}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "456",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "456",
			Name: "stage",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
							Url: "fake-stage-url",
						},
					},
				},
			},
		},
	}), nil)
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb, Driver: sql_manager.MysqlDriver}, nil)
	mockSqlDb.On("Close").Return(nil)
	mockSqlDb.On("GetSchemaColumnMap", mock.Anything).Return(map[string]map[string]*sql_manager.ColumnInfo{
		"public.users":                     {"id": &sql_manager.ColumnInfo{}, "name": &sql_manager.ColumnInfo{}},
		"public.user_account_associations": {"id": &sql_manager.ColumnInfo{}, "user_id": &sql_manager.ColumnInfo{}},
	}, nil)
	mockSqlDb.On("GetPrimaryKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]string{
		"public.users":                     {"id"},
		"public.user_account_associations": {"id"},
	}, nil)
	mockSqlDb.On("GetForeignKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]*sql_manager.ForeignConstraint{
		"public.user_account_associations": {{Columns: []string{"user_id"}, NotNullable: []bool{true}, ForeignKey: &sql_manager.ForeignKey{Table: "public.users", Columns: []string{"id"}}}},
	}, nil)

	bbuilder := newBenthosBuilder(mockSqlManager, mockJobClient, mockConnectionClient, mockTransformersClient, mockJobId, mockRunId, nil, false)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	require.Nil(t, err)
	require.NotEmpty(t, resp.BenthosConfigs)
	require.Len(t, resp.BenthosConfigs, 2)

	bc := getBenthosConfigByName(resp.BenthosConfigs, "public.users.insert")
	require.Equal(t, bc.Name, "public.users.insert")
	require.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: mysql
        dsn: ${SOURCE_CONNECTION_DSN}
        `+"query: SELECT `id`, `name` FROM `public`.`users`;"+`
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - pooled_sql_insert:
                    driver: mysql
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: users
                    columns:
                        - id
                        - name
                    on_conflict_do_nothing: false
                    truncate_on_retry: false
                    args_mapping: root = [this."id", this."name"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	bc2 := getBenthosConfigByName(resp.BenthosConfigs, "public.user_account_associations.insert")
	require.Equal(t, bc2.Name, "public.user_account_associations.insert")
	require.Equal(t, bc2.DependsOn, []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}})
	out2, err := yaml.Marshal(bc2.Config)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(string(out2)),
		strings.TrimSpace(`
input:
    label: ""
    pooled_sql_raw:
        driver: mysql
        dsn: ${SOURCE_CONNECTION_DSN}
        `+"query: SELECT `id`, `user_id` FROM `public`.`user_account_associations`;"+`
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - fallback:
                - pooled_sql_insert:
                    driver: mysql
                    dsn: ${DESTINATION_0_CONNECTION_DSN}
                    schema: public
                    table: user_account_associations
                    columns:
                        - id
                        - user_id
                    on_conflict_do_nothing: false
                    truncate_on_retry: false
                    args_mapping: root = [this."id", this."user_id"]
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
                - error:
                    error_msg: ${! meta("fallback_error")}
                    batching:
                        count: 100
                        byte_size: 0
                        period: 5s
                        check: ""
                        processors: []
`),
	)

	benthosenv := service.NewEnvironment()
	err = neosync_benthos_sql.RegisterPooledSqlInsertOutput(benthosenv, nil, false)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlRawInput(benthosenv, nil, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorOutput(benthosenv, nil)
	require.NoError(t, err)
	newSB := benthosenv.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	require.NoError(t, err)

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out2))
	require.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Mysql_Mysql_With_Circular_Dependency(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Mysql{
							Mysql: &mgmtv1alpha1.MysqlSourceConnectionOptions{
								ConnectionId: "123",
							},
						},
					},
				},
				Mappings: []*mgmtv1alpha1.JobMapping{
					{
						Schema: "public",
						Table:  "users",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "user_assoc_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "user_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						},
					},
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
					},
				},
			},
		}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "123",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "123",
			Name: "prod",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
							Url: "fake-prod-url",
						},
					},
				},
			},
		},
	}), nil)
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: "456",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   "456",
			Name: "stage",
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
							Url: "fake-stage-url",
						},
					},
				},
			},
		},
	}), nil)

	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb, Driver: sql_manager.MysqlDriver}, nil)
	mockSqlDb.On("Close").Return(nil)
	mockSqlDb.On("GetSchemaColumnMap", mock.Anything).Return(map[string]map[string]*sql_manager.ColumnInfo{
		"public.users":                     {"id": &sql_manager.ColumnInfo{}, "user_assoc_id": &sql_manager.ColumnInfo{}, "name": &sql_manager.ColumnInfo{}},
		"public.user_account_associations": {"id": &sql_manager.ColumnInfo{}, "user_id": &sql_manager.ColumnInfo{}},
	}, nil)
	mockSqlDb.On("GetPrimaryKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]string{
		"public.users":                     {"id"},
		"public.user_account_associations": {"id"},
	}, nil)
	mockSqlDb.On("GetForeignKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]*sql_manager.ForeignConstraint{
		"public.user_account_associations": {{Columns: []string{"user_id"}, NotNullable: []bool{true}, ForeignKey: &sql_manager.ForeignKey{Table: "public.users", Columns: []string{"id"}}}},
		"public.users":                     {{Columns: []string{"user_assoc_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.user_account_associations", Columns: []string{"id"}}}},
	}, nil)

	bbuilder := newBenthosBuilder(mockSqlManager, mockJobClient, mockConnectionClient, mockTransformerClient, mockJobId, mockRunId, nil, false)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	require.Nil(t, err)
	require.NotEmpty(t, resp.BenthosConfigs)
	require.Len(t, resp.BenthosConfigs, 3)

	insertConfig := getBenthosConfigByName(resp.BenthosConfigs, "public.users.insert")
	require.NotNil(t, insertConfig)
	require.Equal(t, insertConfig.Name, "public.users.insert")
	require.Empty(t, insertConfig.DependsOn)
	out, err := yaml.Marshal(insertConfig.Config)
	require.NoError(t, err)

	updateConfig := getBenthosConfigByName(resp.BenthosConfigs, "public.users.update")
	require.NotNil(t, updateConfig)
	require.Equal(t, updateConfig.Name, "public.users.update")
	require.Equal(t, updateConfig.DependsOn, []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}, {Table: "public.user_account_associations", Columns: []string{"id"}}})
	out1, err := yaml.Marshal(updateConfig.Config)
	require.NoError(t, err)

	bc2 := getBenthosConfigByName(resp.BenthosConfigs, "public.user_account_associations.insert")
	require.Equal(t, bc2.Name, "public.user_account_associations.insert")
	require.Equal(t, bc2.DependsOn, []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}})
	out2, err := yaml.Marshal(bc2.Config)
	require.NoError(t, err)

	benthosenv := service.NewEnvironment()
	err = neosync_benthos_sql.RegisterPooledSqlInsertOutput(benthosenv, nil, false)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlUpdateOutput(benthosenv, nil)
	require.NoError(t, err)
	err = neosync_benthos_sql.RegisterPooledSqlRawInput(benthosenv, nil, nil)
	require.NoError(t, err)
	err = neosync_benthos_error.RegisterErrorOutput(benthosenv, nil)
	require.NoError(t, err)
	newSB := benthosenv.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	require.NoError(t, err)

	err = newSB.SetYAML(string(out1))
	require.NoError(t, err)

	err = newSB.SetYAML(string(out2))
	require.NoError(t, err)
}

func getBenthosConfigByName(resps []*BenthosConfigResponse, name string) *BenthosConfigResponse {
	for _, cfg := range resps {
		if cfg.Name == name {
			return cfg
		}
	}
	return nil
}

var dsn = "dsn"
var driver = "driver"

func Test_ProcessorConfigEmpty(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	tableMappings := map[string]*tableMapping{
		"public.users": {Schema: "public",
			Table: "users",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "users",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
					},
				},
				{
					Schema: "public",
					Table:  "users",
					Column: "name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED,
					},
				},
			},
		}}

	groupedSchemas := map[string]map[string]*sql_manager.ColumnInfo{
		"public.users": {
			"id": &sql_manager.ColumnInfo{
				OrdinalPosition:        1,
				ColumnDefault:          "324",
				IsNullable:             false,
				DataType:               "",
				CharacterMaximumLength: nil,
				NumericPrecision:       nil,
				NumericScale:           nil,
			},
		},
	}
	queryMap := map[string]map[tabledependency.RunType]string{
		"public.users": {tabledependency.RunTypeInsert: ""},
	}
	runconfigs := []*tabledependency.RunConfig{
		{Table: "public.users", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, Columns: []string{"id", "name"}, DependsOn: []*tabledependency.DependsOn{}},
	}

	res, err := buildBenthosSqlSourceConfigResponses(
		context.Background(),
		mockTransformerClient,
		tableMappings,
		runconfigs,
		dsn,
		driver,
		queryMap,
		groupedSchemas,
		map[string][]*sql_manager.ForeignConstraint{},
		map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{},
		mockJobId,
		mockRunId,
		nil,
		nil,
	)
	require.Nil(t, err)
	require.Empty(t, res[0].Config.StreamConfig.Pipeline.Processors)
}

func Test_ProcessorConfigEmptyJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	tableMappings := map[string]*tableMapping{
		"public.users": {Schema: "public",
			Table: "users",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "users",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
						Config: &mgmtv1alpha1.TransformerConfig{},
					},
				},
				{
					Schema: "public",
					Table:  "users",
					Column: "name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
								TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: ""},
							},
						},
					},
				},
			},
		}}

	groupedSchemas := map[string]map[string]*sql_manager.ColumnInfo{
		"public.users": {
			"id": &sql_manager.ColumnInfo{
				OrdinalPosition:        1,
				ColumnDefault:          "324",
				IsNullable:             false,
				DataType:               "",
				CharacterMaximumLength: nil,
				NumericPrecision:       nil,
				NumericScale:           nil,
			},
		},
	}

	runconfigs := []*tabledependency.RunConfig{
		{Table: "public.users", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, Columns: []string{"id", "name"}, DependsOn: []*tabledependency.DependsOn{}},
	}

	queryMap := map[string]map[tabledependency.RunType]string{
		"public.users": {tabledependency.RunTypeInsert: ""},
	}

	res, err := buildBenthosSqlSourceConfigResponses(
		context.Background(),
		mockTransformerClient,
		tableMappings,
		runconfigs,
		dsn,
		driver,
		queryMap,
		groupedSchemas,
		map[string][]*sql_manager.ForeignConstraint{},
		map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{},
		mockJobId,
		mockRunId,
		nil,
		nil,
	)
	require.Nil(t, err)
	require.Empty(t, res[0].Config.StreamConfig.Pipeline.Processors)
}

func Test_ProcessorConfigMultiJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	tableMappings := map[string]*tableMapping{
		"public.users": {Schema: "public",
			Table: "users",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "users",
					Column: "name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
								TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: `var payload = value + " hello";return payload;`},
							},
						},
					},
				},
				{
					Schema: "public",
					Table:  "users",
					Column: "first_name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
								TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: `var payload = value + " firstname";return payload;`},
							},
						},
					},
				},
			},
		}}

	groupedSchemas := map[string]map[string]*sql_manager.ColumnInfo{
		"public.users": {
			"id": &sql_manager.ColumnInfo{
				OrdinalPosition:        1,
				ColumnDefault:          "324",
				IsNullable:             false,
				DataType:               "",
				CharacterMaximumLength: nil,
				NumericPrecision:       nil,
				NumericScale:           nil,
			},
		},
	}
	queryMap := map[string]map[tabledependency.RunType]string{
		"public.users": {tabledependency.RunTypeInsert: ""},
	}

	runconfigs := []*tabledependency.RunConfig{
		{Table: "public.users", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, Columns: []string{"id", "name"}, DependsOn: []*tabledependency.DependsOn{}},
	}

	res, err := buildBenthosSqlSourceConfigResponses(
		context.Background(),
		mockTransformerClient,
		tableMappings,
		runconfigs,
		dsn,
		driver,
		queryMap,
		groupedSchemas,
		map[string][]*sql_manager.ForeignConstraint{},
		map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{},
		mockJobId,
		mockRunId,
		nil,
		nil,
	)
	require.Nil(t, err)

	out, err := yaml.Marshal(res[0].Config.Pipeline.Processors)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(`
- javascript:
    code: |4-
        (() => {

        function fn_name(value, input){
          var payload = value + " hello";return payload;
        };


        function fn_first_name(value, input){
          var payload = value + " firstname";return payload;
        };

        const input = benthos.v0_msg_as_structured();
        const output = { ...input };
        output["name"] = fn_name(input["name"], input);
        output["first_name"] = fn_first_name(input["first_name"], input);
        benthos.v0_msg_set_structured(output);
        })();
- catch:
    - error:
        error_msg: ${! error()}
      `), strings.TrimSpace(string(out)))
}

func Test_ProcessorConfigMutationAndJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	tableMappings := map[string]*tableMapping{
		"public.users": {Schema: "public",
			Table: "users",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "users",
					Column: "email",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_EMAIL,
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
								GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{},
							},
						},
					},
				},
				{
					Schema: "public",
					Table:  "users",
					Column: "first_name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
								TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: `var payload = value + " firstname";return payload;`},
							},
						},
					},
				},
			},
		}}

	var email int32 = int32(40)

	groupedSchemas := map[string]map[string]*sql_manager.ColumnInfo{
		"public.users": {
			"email": &sql_manager.ColumnInfo{
				OrdinalPosition:        2,
				ColumnDefault:          "",
				IsNullable:             true,
				DataType:               "timestamptz",
				CharacterMaximumLength: &email,
				NumericPrecision:       nil,
				NumericScale:           nil,
			},
		},
	}

	queryMap := map[string]map[tabledependency.RunType]string{
		"public.users": {tabledependency.RunTypeInsert: ""},
	}
	runconfigs := []*tabledependency.RunConfig{
		{Table: "public.users", RunType: tabledependency.RunTypeInsert, PrimaryKeys: []string{"id"}, Columns: []string{"id", "name"}, DependsOn: []*tabledependency.DependsOn{}},
	}

	res, err := buildBenthosSqlSourceConfigResponses(
		context.Background(),
		mockTransformerClient,
		tableMappings,
		runconfigs,
		dsn,
		driver,
		queryMap,
		groupedSchemas,
		map[string][]*sql_manager.ForeignConstraint{},
		map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{},
		mockJobId,
		mockRunId,
		nil,
		nil,
	)

	require.Nil(t, err)

	require.Len(t, res[0].Config.Pipeline.Processors, 3)

	out, err := yaml.Marshal(res[0].Config.Pipeline.Processors)
	require.NoError(t, err)
	require.Equal(
		t,
		strings.TrimSpace(`
- mutation: root."email" = generate_email(max_length:40,email_type:"uuidv4")
- javascript:
    code: |4-
        (() => {

        function fn_first_name(value, input){
          var payload = value + " firstname";return payload;
        };

        const input = benthos.v0_msg_as_structured();
        const output = { ...input };
        output["first_name"] = fn_first_name(input["first_name"], input);
        benthos.v0_msg_set_structured(output);
        })();
- catch:
    - error:
        error_msg: ${! error()}
      `), strings.TrimSpace(string(out)))
}

func TestAreMappingsSubsetOfSchemas(t *testing.T) {
	ok := areMappingsSubsetOfSchemas(
		map[string]map[string]*sql_manager.ColumnInfo{
			"public.users": {
				"id":         &sql_manager.ColumnInfo{},
				"created_by": &sql_manager.ColumnInfo{},
				"updated_by": &sql_manager.ColumnInfo{},
			},
			"neosync_api.accounts": {
				"id": &sql_manager.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	require.True(t, ok, "job mappings are a subset of the present database schemas")

	ok = areMappingsSubsetOfSchemas(
		map[string]map[string]*sql_manager.ColumnInfo{
			"public.users": {
				"id": &sql_manager.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id2"},
		},
	)
	require.False(t, ok, "job mappings contain mapping that is not in the source schema")

	ok = areMappingsSubsetOfSchemas(
		map[string]map[string]*sql_manager.ColumnInfo{
			"public.users": {
				"id": &sql_manager.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	require.False(t, ok, "job mappings contain more mappings than are present in the source schema")
}

func TestShouldHaltOnSchemaAddition(t *testing.T) {
	ok := shouldHaltOnSchemaAddition(
		map[string]map[string]*sql_manager.ColumnInfo{
			"public.users": {
				"id":         &sql_manager.ColumnInfo{},
				"created_by": &sql_manager.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	require.False(t, ok, "job mappings are valid set of database schemas")

	ok = shouldHaltOnSchemaAddition(
		map[string]map[string]*sql_manager.ColumnInfo{
			"public.users": {
				"id":         &sql_manager.ColumnInfo{},
				"created_by": &sql_manager.ColumnInfo{},
			},
			"neosync_api.accounts": {
				"id": &sql_manager.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	require.True(t, ok, "job mappings are missing database schema mappings")

	ok = shouldHaltOnSchemaAddition(
		map[string]map[string]*sql_manager.ColumnInfo{
			"public.users": {
				"id":         &sql_manager.ColumnInfo{},
				"created_by": &sql_manager.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
		},
	)
	require.True(t, ok, "job mappings are missing table column")

	ok = shouldHaltOnSchemaAddition(
		map[string]map[string]*sql_manager.ColumnInfo{
			"public.users": {
				"id":         &sql_manager.ColumnInfo{},
				"created_by": &sql_manager.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "updated_by"},
		},
	)
	require.True(t, ok, "job mappings have same column count, but missing specific column")
}

func Test_buildProcessorConfigsMutation(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	ctx := context.Background()

	output, err := buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{}, map[string]*sql_manager.ColumnInfo{}, map[string][]*referenceKey{}, []string{}, mockJobId, mockRunId, nil)
	require.Nil(t, err)
	require.Empty(t, output)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{}, map[string]*sql_manager.ColumnInfo{}, map[string][]*referenceKey{}, []string{}, mockJobId, mockRunId, nil)
	require.Nil(t, err)
	require.Empty(t, output)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id"},
	}, map[string]*sql_manager.ColumnInfo{}, map[string][]*referenceKey{}, []string{}, mockJobId, mockRunId, nil)
	require.Nil(t, err)
	require.Empty(t, output)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{}},
	}, map[string]*sql_manager.ColumnInfo{}, map[string][]*referenceKey{}, []string{}, mockJobId, mockRunId, nil)
	require.Nil(t, err)
	require.Empty(t, output)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH}},
	}, map[string]*sql_manager.ColumnInfo{}, map[string][]*referenceKey{}, []string{}, mockJobId, mockRunId, nil)
	require.Nil(t, err)
	require.Empty(t, output)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL, Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		}}},
		{Schema: "public", Table: "users", Column: "name", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL, Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		}}},
	}, map[string]*sql_manager.ColumnInfo{}, map[string][]*referenceKey{}, []string{}, mockJobId, mockRunId, nil)

	require.Nil(t, err)

	require.Equal(t, *output[0].Mutation, "root.\"id\" = null\nroot.\"name\" = null")

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
		Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_EMAIL,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
				TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
					PreserveDomain:  true,
					PreserveLength:  false,
					ExcludedDomains: []string{},
				},
			},
		},
	}

	var email int32 = int32(40)

	groupedSchemas := map[string]*sql_manager.ColumnInfo{

		"email": {
			OrdinalPosition:        2,
			ColumnDefault:          "",
			IsNullable:             true,
			DataType:               "timestamptz",
			CharacterMaximumLength: &email,
			NumericPrecision:       nil,
			NumericScale:           nil,
		},
	}

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "email", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}}}, groupedSchemas, map[string][]*referenceKey{}, []string{}, mockJobId, mockRunId, nil)

	require.Nil(t, err)
	require.Equal(t, *output[0].Mutation, `root."email" = transform_email(email:this."email",preserve_domain:true,preserve_length:false,excluded_domains:[],max_length:40,email_type:"uuidv4")`)
}

const transformJsCodeFnStr = `var payload = value+=" hello";return payload;`
const generateJSCodeFnStr = `var payload = "hello";return payload;`

func Test_buildProcessorConfigsJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	ctx := context.Background()

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
		Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: transformJsCodeFnStr,
				},
			},
		},
	}

	res, err := buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "address", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}}}, map[string]*sql_manager.ColumnInfo{}, map[string][]*referenceKey{}, []string{}, mockJobId, mockRunId, nil)

	require.NoError(t, err)
	require.Equal(t, `
(() => {

function fn_address(value, input){
  var payload = value+=" hello";return payload;
};

const input = benthos.v0_msg_as_structured();
const output = { ...input };
output["address"] = fn_address(input["address"], input);
benthos.v0_msg_set_structured(output);
})();`,
		res[0].Javascript.Code,
	)
}

func Test_buildProcessorConfigsGenerateJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	ctx := context.Background()
	genCode := `var payload = "test";return payload;`

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
		Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
				GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
					Code: genCode,
				},
			},
		},
	}

	res, err := buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "test", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}}}, map[string]*sql_manager.ColumnInfo{}, map[string][]*referenceKey{}, []string{}, mockJobId, mockRunId, nil)

	require.NoError(t, err)
	require.Equal(t, `
(() => {

function fn_test(){
  var payload = "test";return payload;
};

const input = benthos.v0_msg_as_structured();
const output = { ...input };
output["test"] = fn_test();
benthos.v0_msg_set_structured(output);
})();`,
		res[0].Javascript.Code,
	)
}

const nameCol = "name"

func Test_buildProcessorConfigsJavascriptMultiLineScript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	ctx := context.Background()

	code :=
		`var payload = value+=" hello";
  payload.replace("hello","newHello");
  return payload;`

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
		Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: code,
				},
			},
		},
	}

	res, err := buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: nameCol, Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}}}, map[string]*sql_manager.ColumnInfo{}, map[string][]*referenceKey{}, []string{}, mockJobId, mockRunId, nil)

	require.NoError(t, err)
	require.Equal(t, `
(() => {

function fn_name(value, input){
  var payload = value+=" hello";
  payload.replace("hello","newHello");
  return payload;
};

const input = benthos.v0_msg_as_structured();
const output = { ...input };
output["name"] = fn_name(input["name"], input);
benthos.v0_msg_set_structured(output);
})();`,
		res[0].Javascript.Code,
	)
}

func Test_buildProcessorConfigsJavascriptMultiple(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	ctx := context.Background()

	code2 := `var payload = value*2;return payload;`
	col2 := "age"

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
		Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: transformJsCodeFnStr,
				},
			},
		},
	}

	jsT2 := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
		Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: code2,
				},
			},
		},
	}

	res, err := buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: nameCol, Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}},
		{Schema: "public", Table: "users", Column: col2, Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT2.Source, Config: jsT2.Config}}}, map[string]*sql_manager.ColumnInfo{}, map[string][]*referenceKey{}, []string{}, mockJobId, mockRunId, nil)

	require.NoError(t, err)
	require.Equal(t, `
(() => {

function fn_name(value, input){
  var payload = value+=" hello";return payload;
};


function fn_age(value, input){
  var payload = value*2;return payload;
};

const input = benthos.v0_msg_as_structured();
const output = { ...input };
output["name"] = fn_name(input["name"], input);
output["age"] = fn_age(input["age"], input);
benthos.v0_msg_set_structured(output);
})();`,
		res[0].Javascript.Code,
	)
}

func Test_buildProcessorConfigsTransformAndGenerateJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	ctx := context.Background()

	col2 := "test"
	genCode := `var payload = "test";return payload;`

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
		Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: transformJsCodeFnStr,
				},
			},
		},
	}

	jsT2 := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
		Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
				GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
					Code: genCode,
				},
			},
		},
	}

	res, err := buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: nameCol, Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}},
		{Schema: "public", Table: "users", Column: col2, Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT2.Source, Config: jsT2.Config}}}, map[string]*sql_manager.ColumnInfo{}, map[string][]*referenceKey{}, []string{}, mockJobId, mockRunId, nil)

	require.NoError(t, err)
	require.Equal(t, `
(() => {

function fn_name(value, input){
  var payload = value+=" hello";return payload;
};


function fn_test(){
  var payload = "test";return payload;
};

const input = benthos.v0_msg_as_structured();
const output = { ...input };
output["name"] = fn_name(input["name"], input);
output["test"] = fn_test();
benthos.v0_msg_set_structured(output);
})();`,
		res[0].Javascript.Code,
	)
}

func Test_ShouldProcessColumnTrue(t *testing.T) {
	val := &mgmtv1alpha1.JobMappingTransformer{
		Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_EMAIL,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		},
	}

	res := shouldProcessColumn(val)
	require.Equal(t, true, res)
}

func Test_ShouldProcessColumnFalse(t *testing.T) {
	val := &mgmtv1alpha1.JobMappingTransformer{
		Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
				PassthroughConfig: &mgmtv1alpha1.Passthrough{},
			},
		},
	}

	res := shouldProcessColumn(val)
	require.Equal(t, false, res)
}

func Test_ConstructJsFunctionTransformJs(t *testing.T) {
	s := mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT

	res := constructJsFunction(transformJsCodeFnStr, "col", s)
	require.Equal(t, `
function fn_col(value, input){
  var payload = value+=" hello";return payload;
};
`, res)
}

func Test_ConstructJsFunctionGenerateJS(t *testing.T) {
	s := mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT

	res := constructJsFunction(generateJSCodeFnStr, "col", s)
	require.Equal(t, `
function fn_col(){
  var payload = "hello";return payload;
};
`, res)
}

func Test_ConstructBenthosJsProcessorTransformJS(t *testing.T) {
	jsFunctions := []string{}
	benthosOutputs := []string{}
	s := mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT

	benthosOutput := constructBenthosJavascriptObject(nameCol, s)
	jsFunction := constructJsFunction(transformJsCodeFnStr, nameCol, s)
	benthosOutputs = append(benthosOutputs, benthosOutput)

	jsFunctions = append(jsFunctions, jsFunction)

	res := constructBenthosJsProcessor(jsFunctions, benthosOutputs)

	require.Equal(t, `
(() => {

function fn_name(value, input){
  var payload = value+=" hello";return payload;
};

const input = benthos.v0_msg_as_structured();
const output = { ...input };
output["name"] = fn_name(input["name"], input);
benthos.v0_msg_set_structured(output);
})();`, res)
}

func Test_ConstructBenthosJsProcessorGenerateJS(t *testing.T) {
	jsFunctions := []string{}
	benthosOutputs := []string{}
	s := mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT

	benthosOutput := constructBenthosJavascriptObject(nameCol, s)
	jsFunction := constructJsFunction(generateJSCodeFnStr, nameCol, s)
	benthosOutputs = append(benthosOutputs, benthosOutput)

	jsFunctions = append(jsFunctions, jsFunction)

	res := constructBenthosJsProcessor(jsFunctions, benthosOutputs)

	require.Equal(t, `
(() => {

function fn_name(){
  var payload = "hello";return payload;
};

const input = benthos.v0_msg_as_structured();
const output = { ...input };
output["name"] = fn_name();
benthos.v0_msg_set_structured(output);
})();`, res)
}

func Test_ConstructBenthosOutputTranformJs(t *testing.T) {
	s := mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT
	res := constructBenthosJavascriptObject("col", s)
	require.Equal(t, `output["col"] = fn_col(input["col"], input);`, res)
}

func Test_ConstructBenthosOutputGenerateJs(t *testing.T) {
	s := mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT
	res := constructBenthosJavascriptObject("col", s)
	require.Equal(t, `output["col"] = fn_col();`, res)
}

func Test_buildProcessorConfigsJavascriptEmpty(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	ctx := context.Background()

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
		Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: ``,
				},
			},
		},
	}

	resp, err := buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}}}, map[string]*sql_manager.ColumnInfo{}, map[string][]*referenceKey{}, []string{}, mockJobId, mockRunId, nil)

	require.NoError(t, err)
	require.Empty(t, resp)
}

func Test_convertUserDefinedFunctionConfig(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	ctx := context.Background()

	mockTransformerClient.On(
		"GetUserDefinedTransformerById",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetUserDefinedTransformerByIdRequest{
			TransformerId: "123",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetUserDefinedTransformerByIdResponse{
		Transformer: &mgmtv1alpha1.UserDefinedTransformer{
			Id:          "123",
			Name:        "stage",
			Description: "description",
			DataType:    mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_EMAIL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
					TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
						PreserveDomain:  true,
						PreserveLength:  false,
						ExcludedDomains: []string{},
					},
				},
			},
		},
	}), nil)

	jmt := &mgmtv1alpha1.JobMappingTransformer{
		Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_EMAIL,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig{
				UserDefinedTransformerConfig: &mgmtv1alpha1.UserDefinedTransformerConfig{
					Id: "123",
				},
			},
		},
	}

	expected := &mgmtv1alpha1.JobMappingTransformer{
		Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_EMAIL,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
				TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
					PreserveDomain:  true,
					PreserveLength:  false,
					ExcludedDomains: []string{},
				},
			},
		},
	}

	resp, err := convertUserDefinedFunctionConfig(ctx, mockTransformerClient, jmt)
	require.NoError(t, err)
	require.Equal(t, resp, expected)
}

func MockJobMappingTransformer(source int32, transformerId string) db_queries.NeosyncApiTransformer {
	return db_queries.NeosyncApiTransformer{
		Source:            source,
		TransformerConfig: &pg_models.TransformerConfigs{},
	}
}

func Test_buildPlainInsertArgs(t *testing.T) {
	require.Empty(t, buildPlainInsertArgs(nil))
	require.Empty(t, buildPlainInsertArgs([]string{}))
	require.Equal(t, buildPlainInsertArgs([]string{"foo", "bar", "baz"}), `root = [this."foo", this."bar", this."baz"]`)
}

func Test_buildPlainColumns(t *testing.T) {
	require.Empty(t, buildPlainColumns(nil))
	require.Empty(t, buildPlainColumns([]*mgmtv1alpha1.JobMapping{}))
	require.Equal(
		t,
		buildPlainColumns([]*mgmtv1alpha1.JobMapping{
			{Column: "foo"},
			{Column: "bar"},
			{Column: "baz"},
		}),
		[]string{"foo", "bar", "baz"},
	)
}

func Test_splitTableKey(t *testing.T) {
	schema, table := splitTableKey("foo")
	require.Equal(t, schema, "public")
	require.Equal(t, table, "foo")

	schema, table = splitTableKey("neosync.foo")
	require.Equal(t, schema, "neosync")
	require.Equal(t, table, "foo")
}

func Test_buildBenthosS3Credentials(t *testing.T) {
	require.Nil(t, buildBenthosS3Credentials(nil))

	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{}),
		&neosync_benthos.AwsCredentials{},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{Profile: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Profile: "foo"},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{AccessKeyId: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Id: "foo"},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{SecretAccessKey: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Secret: "foo"},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{SessionToken: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Token: "foo"},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{FromEc2Role: shared.Ptr(true)}),
		&neosync_benthos.AwsCredentials{FromEc2Role: true},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{RoleArn: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Role: "foo"},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{RoleExternalId: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{RoleExternalId: "foo"},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{
			Profile:         shared.Ptr("profile"),
			AccessKeyId:     shared.Ptr("access-key"),
			SecretAccessKey: shared.Ptr("secret"),
			SessionToken:    shared.Ptr("session"),
			FromEc2Role:     shared.Ptr(false),
			RoleArn:         shared.Ptr("role"),
			RoleExternalId:  shared.Ptr("foo"),
		}),
		&neosync_benthos.AwsCredentials{
			Profile:        "profile",
			Id:             "access-key",
			Secret:         "secret",
			Token:          "session",
			FromEc2Role:    false,
			Role:           "role",
			RoleExternalId: "foo",
		},
	)
}

func Test_computeMutationFunction_null(t *testing.T) {
	val, err := computeMutationFunction(
		&mgmtv1alpha1.JobMapping{
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL,
			},
		}, &sql_manager.ColumnInfo{})
	require.NoError(t, err)
	require.Equal(t, val, "null")
}

func Test_computeMutationFunction_Validate_Bloblang_Output(t *testing.T) {
	uuidEmailType := mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UUID_V4
	transformers := []*mgmtv1alpha1.SystemTransformer{
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_EMAIL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
					GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{
						EmailType: &uuidEmailType,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_EMAIL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
					TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
						PreserveDomain:  false,
						PreserveLength:  false,
						ExcludedDomains: []string{},
						EmailType:       &uuidEmailType,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
					GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CARD_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
					GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{
						ValidLuhn: true,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CITY,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
					GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_E164_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateE164PhoneNumberConfig{
					GenerateE164PhoneNumberConfig: &mgmtv1alpha1.GenerateE164PhoneNumber{
						Min: 9,
						Max: 15,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FIRST_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
					GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FLOAT64,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{
					GenerateFloat64Config: &mgmtv1alpha1.GenerateFloat64{
						RandomizeSign: true,
						Min:           1.00,
						Max:           100.00,
						Precision:     6,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_ADDRESS,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig{
					GenerateFullAddressConfig: &mgmtv1alpha1.GenerateFullAddress{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_GENDER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{
					GenerateGenderConfig: &mgmtv1alpha1.GenerateGender{
						Abbreviate: false,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig{
					GenerateInt64PhoneNumberConfig: &mgmtv1alpha1.GenerateInt64PhoneNumber{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
						RandomizeSign: true,
						Min:           1,
						Max:           40,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_LAST_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{
					GenerateLastNameConfig: &mgmtv1alpha1.GenerateLastName{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SHA256HASH,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{
					GenerateSha256HashConfig: &mgmtv1alpha1.GenerateSha256Hash{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SSN,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
					GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STATE,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{
					GenerateStateConfig: &mgmtv1alpha1.GenerateState{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STREET_ADDRESS,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig{
					GenerateStreetAddressConfig: &mgmtv1alpha1.GenerateStreetAddress{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STRING_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig{
					GenerateStringPhoneNumberConfig: &mgmtv1alpha1.GenerateStringPhoneNumber{
						Min: 9,
						Max: 14,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_RANDOM_STRING,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
					GenerateStringConfig: &mgmtv1alpha1.GenerateString{
						Min: 2,
						Max: 7,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UNIXTIMESTAMP,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{
					GenerateUnixtimestampConfig: &mgmtv1alpha1.GenerateUnixTimestamp{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_USERNAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig{
					GenerateUsernameConfig: &mgmtv1alpha1.GenerateUsername{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UTCTIMESTAMP,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig{
					GenerateUtctimestampConfig: &mgmtv1alpha1.GenerateUtcTimestamp{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UUID,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
					GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
						IncludeHyphens: true,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_ZIPCODE,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig{
					GenerateZipcodeConfig: &mgmtv1alpha1.GenerateZipcode{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_E164_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig{
					TransformE164PhoneNumberConfig: &mgmtv1alpha1.TransformE164PhoneNumber{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FIRST_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
					TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FLOAT64,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFloat64Config{
					TransformFloat64Config: &mgmtv1alpha1.TransformFloat64{
						RandomizationRangeMin: 20.00,
						RandomizationRangeMax: 50.00,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FULL_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
					TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig{
					TransformInt64PhoneNumberConfig: &mgmtv1alpha1.TransformInt64PhoneNumber{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
					TransformInt64Config: &mgmtv1alpha1.TransformInt64{
						RandomizationRangeMin: 20,
						RandomizationRangeMax: 50,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_LAST_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{
					TransformLastNameConfig: &mgmtv1alpha1.TransformLastName{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig{
					TransformPhoneNumberConfig: &mgmtv1alpha1.TransformPhoneNumber{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_STRING,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
					TransformStringConfig: &mgmtv1alpha1.TransformString{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CATEGORICAL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig{
					GenerateCategoricalConfig: &mgmtv1alpha1.GenerateCategorical{
						Categories: "value1,value2",
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_CHARACTER_SCRAMBLE,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
					TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{
						UserProvidedRegex: nil,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{
					GenerateDefaultConfig: &mgmtv1alpha1.GenerateDefault{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
					Nullconfig: &mgmtv1alpha1.Null{},
				},
			},
		},
	}

	emailColInfo := &sql_manager.ColumnInfo{
		OrdinalPosition:        2,
		ColumnDefault:          "",
		IsNullable:             true,
		DataType:               "timestamptz",
		CharacterMaximumLength: shared.Ptr(int32(40)),
		NumericPrecision:       nil,
		NumericScale:           nil,
	}

	for _, transformer := range transformers {
		t.Run(fmt.Sprintf("%s_%s_lint", t.Name(), transformer.Source), func(t *testing.T) {
			val, err := computeMutationFunction(
				&mgmtv1alpha1.JobMapping{
					Column: "email",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: transformer.Source,
						Config: transformer.Config,
					},
				}, emailColInfo)
			require.NoError(t, err)
			ex, err := bloblang.Parse(val)
			require.NoError(t, err, fmt.Sprintf("transformer lint failed, check that the transformer string is being constructed correctly. Failing source: %s", transformer.Source))
			_, err = ex.Query(nil)
			require.NoError(t, err)
		})
	}
}

func Test_computeMutationFunction_handles_Db_Maxlen(t *testing.T) {
	type testcase struct {
		jm       *mgmtv1alpha1.JobMapping
		ci       *sql_manager.ColumnInfo
		expected string
	}
	jm := &mgmtv1alpha1.JobMapping{
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_RANDOM_STRING,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
					GenerateStringConfig: &mgmtv1alpha1.GenerateString{
						Min: 2,
						Max: 7,
					},
				},
			},
		},
	}
	testcases := []testcase{
		{
			jm:       jm,
			ci:       &sql_manager.ColumnInfo{},
			expected: "generate_string(min:2,max:7)",
		},
		{
			jm: jm,
			ci: &sql_manager.ColumnInfo{
				CharacterMaximumLength: nil,
			},
			expected: "generate_string(min:2,max:7)",
		},
		{
			jm: jm,
			ci: &sql_manager.ColumnInfo{
				CharacterMaximumLength: shared.Ptr(int32(-1)),
			},
			expected: "generate_string(min:2,max:7)",
		},
		{
			jm: jm,
			ci: &sql_manager.ColumnInfo{
				CharacterMaximumLength: shared.Ptr(int32(0)),
			},
			expected: "generate_string(min:2,max:7)",
		},
		{
			jm: jm,
			ci: &sql_manager.ColumnInfo{
				CharacterMaximumLength: shared.Ptr(int32(10)),
			},
			expected: "generate_string(min:2,max:7)",
		},
		{
			jm: jm,
			ci: &sql_manager.ColumnInfo{
				CharacterMaximumLength: shared.Ptr(int32(3)),
			},
			expected: "generate_string(min:2,max:3)",
		},
		{
			jm: jm,
			ci: &sql_manager.ColumnInfo{
				CharacterMaximumLength: shared.Ptr(int32(1)),
			},
			expected: "generate_string(min:1,max:1)",
		},
	}

	for _, tc := range testcases {
		t.Run(t.Name(), func(t *testing.T) {
			out, err := computeMutationFunction(tc.jm, tc.ci)
			require.NoError(t, err)
			require.NotNil(t, out)
			require.Equal(t, tc.expected, out, "computed bloblang string was not expected")
			ex, err := bloblang.Parse(out)
			require.NoError(t, err)
			_, err = ex.Query(nil)
			require.NoError(t, err)
		})
	}
}

func Test_buildBranchCacheConfigs_null(t *testing.T) {
	cols := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "public",
			Table:  "users",
			Column: "user_id",
		},
	}

	constraints := map[string][]*referenceKey{
		"name": {
			{
				Table:  "public.orders",
				Column: "buyer_id",
			},
		},
	}

	resp, err := buildBranchCacheConfigs(cols, constraints, mockJobId, mockRunId, nil)
	require.NoError(t, err)
	require.Len(t, resp, 0)
}

func Test_buildBranchCacheConfigs_missing_redis(t *testing.T) {
	cols := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "public",
			Table:  "users",
			Column: "user_id",
		},
	}

	constraints := map[string][]*referenceKey{
		"user_id": {
			{
				Table:  "public.orders",
				Column: "buyer_id",
			},
		},
	}

	_, err := buildBranchCacheConfigs(cols, constraints, mockJobId, mockRunId, nil)
	require.Error(t, err)
}

func Test_buildBranchCacheConfigs_success(t *testing.T) {
	cols := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "public",
			Table:  "users",
			Column: "user_id",
		},
		{
			Schema: "public",
			Table:  "users",
			Column: "name",
		},
	}

	constraints := map[string][]*referenceKey{
		"user_id": {
			{
				Table:  "public.orders",
				Column: "buyer_id",
			},
		},
	}
	redisConfig := &shared.RedisConfig{
		Url:  "redis://localhost:6379",
		Kind: "simple",
	}

	resp, err := buildBranchCacheConfigs(cols, constraints, mockJobId, mockRunId, redisConfig)

	require.NoError(t, err)
	require.Len(t, resp, 1)
	require.Equal(t, *resp[0].RequestMap, `root = if this."user_id" == null { deleted() } else { this }`)
	require.Equal(t, *resp[0].ResultMap, `root."user_id" = this`)
}

func Test_buildBranchCacheConfigs_self_referencing(t *testing.T) {
	cols := []*mgmtv1alpha1.JobMapping{
		{
			Schema: "public",
			Table:  "users",
			Column: "user_id",
		},
	}

	constraints := map[string][]*referenceKey{
		"user_id": {
			{
				Table:  "public.users",
				Column: "other_id",
			},
		},
	}
	redisConfig := &shared.RedisConfig{
		Url:  "redis://localhost:6379",
		Kind: "simple",
	}

	resp, err := buildBranchCacheConfigs(cols, constraints, mockJobId, mockRunId, redisConfig)
	require.NoError(t, err)
	require.Len(t, resp, 0)
}

func Test_ConverStringSliceToStringEmptySlice(t *testing.T) {
	slc := []string{}

	res, err := convertStringSliceToString(slc)
	require.NoError(t, err)
	require.Equal(t, "[]", res)
}

func Test_ConverStringSliceToStringNotEmptySlice(t *testing.T) {
	slc := []string{"gmail.com", "yahoo.com"}

	res, err := convertStringSliceToString(slc)
	require.NoError(t, err)
	require.Equal(t, `["gmail.com","yahoo.com"]`, res)
}

func Test_getPrimaryKeyDependencyMap(t *testing.T) {
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"hr.countries": {
			{
				Columns:     []string{"region_id"},
				NotNullable: []bool{true},
				ForeignKey: &sql_manager.ForeignKey{
					Table:   "hr.regions",
					Columns: []string{"region_id"},
				},
			},
		},
		"hr.departments": {
			{
				Columns:     []string{"location_id"},
				NotNullable: []bool{false},
				ForeignKey: &sql_manager.ForeignKey{
					Table:   "hr.locations",
					Columns: []string{"location_id"},
				},
			},
		},
		"hr.dependents": {
			{
				Columns:     []string{"employee_id"},
				NotNullable: []bool{true},
				ForeignKey: &sql_manager.ForeignKey{
					Table:   "hr.employees",
					Columns: []string{"employee_id"},
				},
			},
		},
		"hr.employees": {
			{
				Columns:     []string{"job_id"},
				NotNullable: []bool{true},
				ForeignKey: &sql_manager.ForeignKey{
					Table:   "hr.jobs",
					Columns: []string{"job_id"},
				},
			},
			{
				Columns:     []string{"department_id"},
				NotNullable: []bool{false},
				ForeignKey: &sql_manager.ForeignKey{
					Table:   "hr.departments",
					Columns: []string{"department_id"},
				},
			},
			{
				Columns:     []string{"manager_id"},
				NotNullable: []bool{false},
				ForeignKey: &sql_manager.ForeignKey{
					Table:   "hr.employees",
					Columns: []string{"employee_id"},
				},
			},
		},
		"hr.locations": {
			{
				Columns:     []string{"country_id"},
				NotNullable: []bool{true},
				ForeignKey: &sql_manager.ForeignKey{
					Table:   "hr.countries",
					Columns: []string{"country_id"},
				},
			},
		},
	}

	expected := map[string]map[string][]*referenceKey{
		"hr.regions": {
			"region_id": {
				{
					Table:  "hr.countries",
					Column: "region_id",
				},
			},
		},
		"hr.locations": {
			"location_id": {
				{
					Table:  "hr.departments",
					Column: "location_id",
				},
			},
		},
		"hr.employees": {
			"employee_id": {
				{
					Table:  "hr.dependents",
					Column: "employee_id",
				},
				{
					Table:  "hr.employees",
					Column: "manager_id",
				},
			},
		},
		"hr.jobs": {
			"job_id": {
				{
					Table:  "hr.employees",
					Column: "job_id",
				},
			},
		},
		"hr.departments": {
			"department_id": {
				{
					Table:  "hr.employees",
					Column: "department_id",
				},
			},
		},
		"hr.countries": {
			"country_id": {
				{
					Table:  "hr.locations",
					Column: "country_id",
				},
			},
		},
	}

	actual := getPrimaryKeyDependencyMap(tableDependencies)
	for table, depsMap := range expected {
		actualDepsMap := actual[table]
		require.NotNil(t, actualDepsMap)
		for col, deps := range depsMap {
			actualDeps := actualDepsMap[col]
			require.ElementsMatch(t, deps, actualDeps)
		}
	}
}

func Test_getPrimaryKeyDependencyMap_compositekeys(t *testing.T) {
	tableDependencies := map[string][]*sql_manager.ForeignConstraint{
		"employees": {
			{
				Columns:     []string{"department_id"},
				NotNullable: []bool{false},
				ForeignKey: &sql_manager.ForeignKey{
					Table:   "department",
					Columns: []string{"department_id"},
				},
			},
		},
		"projects": {
			{
				Columns:     []string{"responsible_employee_id", "responsible_department_id"},
				NotNullable: []bool{true},
				ForeignKey: &sql_manager.ForeignKey{
					Table:   "employees",
					Columns: []string{"employee_id", "department_id"},
				},
			},
		},
	}

	expected := map[string]map[string][]*referenceKey{
		"department": {
			"department_id": {
				{
					Table:  "employees",
					Column: "department_id",
				},
			},
		},
		"employees": {
			"employee_id": {{
				Table:  "projects",
				Column: "responsible_employee_id",
			}},
			"department_id": {{
				Table:  "projects",
				Column: "responsible_department_id",
			}},
		},
	}

	actual := getPrimaryKeyDependencyMap(tableDependencies)
	require.Equal(t, expected, actual)
}
