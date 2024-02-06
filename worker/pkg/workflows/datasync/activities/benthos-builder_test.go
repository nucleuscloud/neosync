package datasync_activities

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"connectrpc.com/connect"
	sb "github.com/benthosdev/benthos/v4/public/service"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v3"
)

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Generate_Pg(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{
		"123": pg_queries.NewMockDBTX(t),
		"456": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

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
							Source: "generate_uuid",
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
							Source: "generate_ssn",
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

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)
	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.BenthosConfigs)
	assert.Len(t, resp.BenthosConfigs, 1)
	bc := resp.BenthosConfigs[0]
	assert.Equal(t, bc.Name, "public.users")
	assert.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    generate:
        mapping: |-
            root.id = generate_uuid(include_hyphens:true)
            root.name = generate_ssn()
        interval: ""
        count: 10
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: postgres
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.users (id, name) VALUES ($1, $2);
                args_mapping: root = [this.id, this.name]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	// create a new streambuilder instance so we can access the SetYaml method
	newSB := sb.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	assert.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Generate_Pg_Default(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{
		"123": pg_queries.NewMockDBTX(t),
		"456": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

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
							Source: "generate_default",
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
							Source: "generate_ssn",
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

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)
	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.BenthosConfigs)
	assert.Len(t, resp.BenthosConfigs, 1)
	bc := resp.BenthosConfigs[0]
	assert.Equal(t, bc.Name, "public.users")
	assert.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    generate:
        mapping: root.name = generate_ssn()
        interval: ""
        count: 10
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: postgres
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.users (id, name) VALUES (DEFAULT, $1);
                args_mapping: root = [this.name]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	// create a new streambuilder instance so we can access the SetYaml method
	newSB := sb.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	assert.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Pg_Pg(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{
		"123": pg_queries.NewMockDBTX(t),
		"456": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

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
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
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

	pgquerier.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*pg_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
		}, nil)
	pgquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetForeignKeyConstraintsRow{}, nil)
	pgquerier.On("GetPrimaryKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetPrimaryKeyConstraintsRow{{
			SchemaName:     "public",
			TableName:      "users",
			ConstraintName: "name",
			ColumnName:     "id",
		}}, nil)
	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.BenthosConfigs)
	assert.Len(t, resp.BenthosConfigs, 1)
	bc := resp.BenthosConfigs[0]
	assert.Equal(t, bc.Name, "public.users")
	assert.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.users
        columns:
            - id
            - name
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: postgres
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.users (id, name) VALUES ($1, $2);
                args_mapping: root = [this.id, this.name]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	newSB := sb.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	assert.NoError(t, err)

}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Pg_Pg_Default(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{
		"123": pg_queries.NewMockDBTX(t),
		"456": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

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
							Source: "generate_default",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
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

	pgquerier.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*pg_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
		}, nil)
	pgquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetForeignKeyConstraintsRow{}, nil)
	pgquerier.On("GetPrimaryKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetPrimaryKeyConstraintsRow{{
			SchemaName:     "public",
			TableName:      "users",
			ConstraintName: "name",
			ColumnName:     "id",
		}}, nil)
	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.BenthosConfigs)
	assert.Len(t, resp.BenthosConfigs, 1)
	bc := resp.BenthosConfigs[0]
	assert.Equal(t, bc.Name, "public.users")
	assert.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.users
        columns:
            - id
            - name
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: postgres
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.users (id, name) VALUES (DEFAULT, $1);
                args_mapping: root = [this.name]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	// create a new streambuilder instance so we can access the SetYaml method
	newSB := sb.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	assert.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Pg_Pg_With_Constraints(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{
		"123": pg_queries.NewMockDBTX(t),
		"456": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

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
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "user_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
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

	pgquerier.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*pg_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
			{
				TableSchema: "public",
				TableName:   "user_account_associations",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "user_account_associations",
				ColumnName:  "user_id",
			},
		}, nil)
	pgquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetForeignKeyConstraintsRow{
			{
				ConstraintName:    "fk_user_account_associations_user_id_users_id",
				SchemaName:        "public",
				TableName:         "user_account_associations",
				ColumnName:        "user_id",
				ForeignSchemaName: "public",
				ForeignTableName:  "users",
				ForeignColumnName: "id",
			},
		}, nil)
	pgquerier.On("GetPrimaryKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetPrimaryKeyConstraintsRow{{
			SchemaName:     "public",
			TableName:      "users",
			ConstraintName: "name",
			ColumnName:     "id",
		}, {
			SchemaName:     "public",
			TableName:      "user_account_associations",
			ConstraintName: "acc_assoc_constraint",
			ColumnName:     "id",
		}}, nil)
	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.BenthosConfigs)
	assert.Len(t, resp.BenthosConfigs, 2)

	bc := getBenthosConfigByName(resp.BenthosConfigs, "public.users")
	assert.NotNil(t, bc)
	assert.Equal(t, bc.Name, "public.users")
	assert.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.users
        columns:
            - id
            - name
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: postgres
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.users (id, name) VALUES ($1, $2);
                args_mapping: root = [this.id, this.name]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	bc2 := getBenthosConfigByName(resp.BenthosConfigs, "public.user_account_associations")
	assert.Equal(t, bc2.Name, "public.user_account_associations")
	assert.Equal(t, bc2.DependsOn, []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}})
	out2, err := yaml.Marshal(bc2.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out2)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.user_account_associations
        columns:
            - id
            - user_id
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: postgres
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.user_account_associations (id, user_id) VALUES ($1, $2);
                args_mapping: root = [this.id, this.user_id]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	newSB := sb.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	assert.NoError(t, err)

	err = newSB.SetYAML(string(out2))
	assert.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Pg_Pg_With_Circular_Dependency(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{
		"123": pg_queries.NewMockDBTX(t),
		"456": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

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
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "user_assoc_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "user_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
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

	pgquerier.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*pg_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "user_assoc_id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
			{
				TableSchema: "public",
				TableName:   "user_account_associations",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "user_account_associations",
				ColumnName:  "user_id",
			},
		}, nil)
	pgquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetForeignKeyConstraintsRow{
			{
				ConstraintName:    "fk_user_account_associations_user_id_users_id",
				SchemaName:        "public",
				TableName:         "user_account_associations",
				ColumnName:        "user_id",
				ForeignSchemaName: "public",
				ForeignTableName:  "users",
				ForeignColumnName: "id",
				IsNullable:        "NO",
			},
			{
				ConstraintName:    "fk_users_user_assoc_id_user_account_associations_id",
				SchemaName:        "public",
				TableName:         "users",
				ColumnName:        "user_assoc_id",
				ForeignSchemaName: "public",
				ForeignTableName:  "user_account_associations",
				ForeignColumnName: "id",
				IsNullable:        "YES",
			},
		}, nil)
	pgquerier.On("GetPrimaryKeyConstraints", mock.Anything, mock.Anything, mock.Anything).Return([]*pg_queries.GetPrimaryKeyConstraintsRow{
		{
			ConstraintName: "pkey-user-id",
			SchemaName:     "public",
			TableName:      "users",
			ColumnName:     "id",
		},
		{
			ConstraintName: "pkey-user-assoc-id",
			SchemaName:     "public",
			TableName:      "users_account_associations",
			ColumnName:     "id",
		},
	}, nil)
	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.BenthosConfigs)
	assert.Len(t, resp.BenthosConfigs, 3)

	insertConfig := getBenthosConfigByName(resp.BenthosConfigs, "public.users")
	assert.NotNil(t, insertConfig)
	assert.Equal(t, insertConfig.Name, "public.users")
	assert.Empty(t, insertConfig.DependsOn)
	out, err := yaml.Marshal(insertConfig.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.users
        columns:
            - id
            - name
            - user_assoc_id
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: postgres
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.users (id, name) VALUES ($1, $2);
                args_mapping: root = [this.id, this.name]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	updateConfig := getBenthosConfigByName(resp.BenthosConfigs, "public.users.update")
	assert.NotNil(t, updateConfig)
	assert.Equal(t, updateConfig.Name, "public.users.update")
	assert.Equal(t, updateConfig.DependsOn, []*tabledependency.DependsOn{{Table: "public.user_account_associations", Columns: []string{"id"}}})
	out1, err := yaml.Marshal(updateConfig.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out1)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.users
        columns:
            - id
            - name
            - user_assoc_id
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: postgres
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: UPDATE public.users SET user_assoc_id = $1 WHERE id = $2;
                args_mapping: root = [this.user_assoc_id, this.id]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	bc2 := getBenthosConfigByName(resp.BenthosConfigs, "public.user_account_associations")
	assert.Equal(t, bc2.Name, "public.user_account_associations")
	assert.Equal(t, bc2.DependsOn, []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}})
	out2, err := yaml.Marshal(bc2.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out2)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.user_account_associations
        columns:
            - id
            - user_id
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: postgres
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.user_account_associations (id, user_id) VALUES ($1, $2);
                args_mapping: root = [this.id, this.user_id]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	newSB := sb.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	assert.NoError(t, err)

	err = newSB.SetYAML(string(out2))
	assert.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Pg_Pg_With_Circular_Dependency_S3(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{
		"123": pg_queries.NewMockDBTX(t),
		"456": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

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
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "user_assoc_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "user_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
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

	pgquerier.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*pg_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "user_assoc_id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
			{
				TableSchema: "public",
				TableName:   "user_account_associations",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "user_account_associations",
				ColumnName:  "user_id",
			},
		}, nil)
	pgquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetForeignKeyConstraintsRow{
			{
				ConstraintName:    "fk_user_account_associations_user_id_users_id",
				SchemaName:        "public",
				TableName:         "user_account_associations",
				ColumnName:        "user_id",
				ForeignSchemaName: "public",
				ForeignTableName:  "users",
				ForeignColumnName: "id",
				IsNullable:        "NO",
			},
			{
				ConstraintName:    "fk_users_user_assoc_id_user_account_associations_id",
				SchemaName:        "public",
				TableName:         "users",
				ColumnName:        "user_assoc_id",
				ForeignSchemaName: "public",
				ForeignTableName:  "user_account_associations",
				ForeignColumnName: "id",
				IsNullable:        "YES",
			},
		}, nil)
	pgquerier.On("GetPrimaryKeyConstraints", mock.Anything, mock.Anything, mock.Anything).Return([]*pg_queries.GetPrimaryKeyConstraintsRow{
		{
			ConstraintName: "pkey-user-id",
			SchemaName:     "public",
			TableName:      "users",
			ColumnName:     "id",
		},
		{
			ConstraintName: "pkey-user-assoc-id",
			SchemaName:     "public",
			TableName:      "users_account_associations",
			ColumnName:     "id",
		},
	}, nil)
	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.BenthosConfigs)
	assert.Len(t, resp.BenthosConfigs, 3)

	insertConfig := getBenthosConfigByName(resp.BenthosConfigs, "public.users")
	assert.NotNil(t, insertConfig)
	assert.Equal(t, insertConfig.Name, "public.users")
	assert.Empty(t, insertConfig.DependsOn)
	out, err := yaml.Marshal(insertConfig.Config)

	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.users
        columns:
            - id
            - name
            - user_assoc_id
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: postgres
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.users (id, name) VALUES ($1, $2);
                args_mapping: root = [this.id, this.name]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
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
`),
	)

	updateConfig := getBenthosConfigByName(resp.BenthosConfigs, "public.users.update")
	assert.NotNil(t, updateConfig)
	assert.Equal(t, updateConfig.Name, "public.users.update")
	assert.Equal(t, updateConfig.DependsOn, []*tabledependency.DependsOn{{Table: "public.user_account_associations", Columns: []string{"id"}}})
	out1, err := yaml.Marshal(updateConfig.Config)

	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out1)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.users
        columns:
            - id
            - name
            - user_assoc_id
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: postgres
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: UPDATE public.users SET user_assoc_id = $1 WHERE id = $2;
                args_mapping: root = [this.user_assoc_id, this.id]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	bc2 := getBenthosConfigByName(resp.BenthosConfigs, "public.user_account_associations")
	assert.Equal(t, bc2.Name, "public.user_account_associations")
	assert.Equal(t, bc2.DependsOn, []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}})
	out2, err := yaml.Marshal(bc2.Config)

	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out2)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: postgres
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.user_account_associations
        columns:
            - id
            - user_id
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: postgres
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.user_account_associations (id, user_id) VALUES ($1, $2);
                args_mapping: root = [this.id, this.user_id]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
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
`),
	)

	newSB := sb.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	assert.NoError(t, err)

	err = newSB.SetYAML(string(out2))
	assert.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Generate_Mysql(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{
		"123": mysql_queries.NewMockDBTX(t),
		"456": mysql_queries.NewMockDBTX(t),
	}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

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
							Source: "generate_uuid",
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
							Source: "generate_ssn",
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

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.BenthosConfigs)
	assert.Len(t, resp.BenthosConfigs, 1)
	bc := resp.BenthosConfigs[0]
	assert.Equal(t, bc.Name, "public.users")
	assert.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    generate:
        mapping: |-
            root.id = generate_uuid(include_hyphens:true)
            root.name = generate_ssn()
        interval: ""
        count: 10
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: mysql
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.users (id, name) VALUES (?, ?);
                args_mapping: root = [this.id, this.name]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	newSB := sb.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	assert.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Mysql_Mysql(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{
		"123": mysql_queries.NewMockDBTX(t),
		"456": mysql_queries.NewMockDBTX(t),
	}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

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
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
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

	mysqlquerier.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
		}, nil)
	mysqlquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetForeignKeyConstraintsRow{}, nil)
	mysqlquerier.On("GetPrimaryKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetPrimaryKeyConstraintsRow{{
			SchemaName:     "public",
			TableName:      "users",
			ConstraintName: "pk-id",
			ColumnName:     "id",
		}}, nil)
	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.BenthosConfigs)
	assert.Len(t, resp.BenthosConfigs, 1)
	bc := resp.BenthosConfigs[0]
	assert.Equal(t, bc.Name, "public.users")
	assert.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: mysql
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.users
        columns:
            - id
            - name
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: mysql
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.users (id, name) VALUES (?, ?);
                args_mapping: root = [this.id, this.name]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	newSB := sb.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	assert.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Mysql_Mysql_With_Constraints(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformersClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{
		"123": mysql_queries.NewMockDBTX(t),
		"456": mysql_queries.NewMockDBTX(t),
	}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

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
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "user_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
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

	mysqlquerier.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
			{
				TableSchema: "public",
				TableName:   "user_account_associations",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "user_account_associations",
				ColumnName:  "user_id",
			},
		}, nil)
	mysqlquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetForeignKeyConstraintsRow{
			{
				ConstraintName:    "fk_user_account_associations_user_id_users_id",
				SchemaName:        "public",
				TableName:         "user_account_associations",
				ColumnName:        "user_id",
				ForeignSchemaName: "public",
				ForeignTableName:  "users",
				ForeignColumnName: "id",
			},
		}, nil)
	mysqlquerier.On("GetPrimaryKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetPrimaryKeyConstraintsRow{{
			SchemaName:     "public",
			TableName:      "users",
			ConstraintName: "pk-users-id",
			ColumnName:     "id",
		}, {
			SchemaName:     "public",
			TableName:      "user_account_associations",
			ConstraintName: "pk-users-assoc-id",
			ColumnName:     "id",
		}}, nil)
	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformersClient, mockSqlConnector)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.BenthosConfigs)
	assert.Len(t, resp.BenthosConfigs, 2)

	bc := getBenthosConfigByName(resp.BenthosConfigs, "public.users")
	assert.Equal(t, bc.Name, "public.users")
	assert.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: mysql
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.users
        columns:
            - id
            - name
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: mysql
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.users (id, name) VALUES (?, ?);
                args_mapping: root = [this.id, this.name]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	bc2 := getBenthosConfigByName(resp.BenthosConfigs, "public.user_account_associations")
	assert.Equal(t, bc2.Name, "public.user_account_associations")
	assert.Equal(t, bc2.DependsOn, []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}})
	out2, err := yaml.Marshal(bc2.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out2)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: mysql
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.user_account_associations
        columns:
            - id
            - user_id
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: mysql
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.user_account_associations (id, user_id) VALUES (?, ?);
                args_mapping: root = [this.id, this.user_id]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	newSB := sb.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	assert.NoError(t, err)

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out2))
	assert.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Mysql_Mysql_With_Circular_Dependency(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{
		"123": mysql_queries.NewMockDBTX(t),
		"456": mysql_queries.NewMockDBTX(t),
	}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

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
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "user_assoc_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "user_account_associations",
						Column: "user_id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
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

	mysqlquerier.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "user_assoc_id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
			{
				TableSchema: "public",
				TableName:   "user_account_associations",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "user_account_associations",
				ColumnName:  "user_id",
			},
		}, nil)
	mysqlquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetForeignKeyConstraintsRow{
			{
				ConstraintName:    "fk_user_account_associations_user_id_users_id",
				SchemaName:        "public",
				TableName:         "user_account_associations",
				ColumnName:        "user_id",
				ForeignSchemaName: "public",
				ForeignTableName:  "users",
				ForeignColumnName: "id",
				IsNullable:        "NO",
			},
			{
				ConstraintName:    "fk_users_user_assoc_id_user_account_associations_id",
				SchemaName:        "public",
				TableName:         "users",
				ColumnName:        "user_assoc_id",
				ForeignSchemaName: "public",
				ForeignTableName:  "user_account_associations",
				ForeignColumnName: "id",
				IsNullable:        "YES",
			},
		}, nil)
	mysqlquerier.On("GetPrimaryKeyConstraints", mock.Anything, mock.Anything, mock.Anything).Return([]*mysql_queries.GetPrimaryKeyConstraintsRow{
		{
			ConstraintName: "pkey-user-id",
			SchemaName:     "public",
			TableName:      "users",
			ColumnName:     "id",
		},
		{
			ConstraintName: "pkey-user-assoc-id",
			SchemaName:     "public",
			TableName:      "users_account_associations",
			ColumnName:     "id",
		},
	}, nil)
	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.BenthosConfigs)
	assert.Len(t, resp.BenthosConfigs, 3)

	insertConfig := getBenthosConfigByName(resp.BenthosConfigs, "public.users")
	assert.NotNil(t, insertConfig)
	assert.Equal(t, insertConfig.Name, "public.users")
	assert.Empty(t, insertConfig.DependsOn)
	out, err := yaml.Marshal(insertConfig.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: mysql
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.users
        columns:
            - id
            - name
            - user_assoc_id
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: mysql
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.users (id, name) VALUES (?, ?);
                args_mapping: root = [this.id, this.name]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	updateConfig := getBenthosConfigByName(resp.BenthosConfigs, "public.users.update")
	assert.NotNil(t, updateConfig)
	assert.Equal(t, updateConfig.Name, "public.users.update")
	assert.Equal(t, updateConfig.DependsOn, []*tabledependency.DependsOn{{Table: "public.user_account_associations", Columns: []string{"id"}}})
	out1, err := yaml.Marshal(updateConfig.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out1)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: mysql
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.users
        columns:
            - id
            - name
            - user_assoc_id
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: mysql
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: UPDATE public.users SET user_assoc_id = ? WHERE id = ?;
                args_mapping: root = [this.user_assoc_id, this.id]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	bc2 := getBenthosConfigByName(resp.BenthosConfigs, "public.user_account_associations")
	assert.Equal(t, bc2.Name, "public.user_account_associations")
	assert.Equal(t, bc2.DependsOn, []*tabledependency.DependsOn{{Table: "public.users", Columns: []string{"id"}}})
	out2, err := yaml.Marshal(bc2.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out2)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: mysql
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.user_account_associations
        columns:
            - id
            - user_id
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: mysql
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.user_account_associations (id, user_id) VALUES (?, ?);
                args_mapping: root = [this.id, this.user_id]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	newSB := sb.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	assert.NoError(t, err)

	err = newSB.SetYAML(string(out2))
	assert.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Generate_Mysql_Default(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{
		"123": mysql_queries.NewMockDBTX(t),
		"456": mysql_queries.NewMockDBTX(t),
	}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

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
							Source: "generate_default",
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
							Source: "generate_ssn",
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

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.BenthosConfigs)
	assert.Len(t, resp.BenthosConfigs, 1)
	bc := resp.BenthosConfigs[0]
	assert.Equal(t, bc.Name, "public.users")
	assert.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    generate:
        mapping: root.name = generate_ssn()
        interval: ""
        count: 10
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: mysql
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.users (id, name) VALUES (DEFAULT, ?);
                args_mapping: root = [this.name]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)

	// create a new streambuilder instance so we can access the SetYaml method
	newSB := sb.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	assert.NoError(t, err)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Mysql_Default(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{
		"123": mysql_queries.NewMockDBTX(t),
		"456": mysql_queries.NewMockDBTX(t),
	}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

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
							Source: "generate_default",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Source: "passthrough",
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

	mysqlquerier.On("GetDatabaseSchema", mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetDatabaseSchemaRow{
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "id",
			},
			{
				TableSchema: "public",
				TableName:   "users",
				ColumnName:  "name",
			},
		}, nil)
	mysqlquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetForeignKeyConstraintsRow{}, nil)
	mysqlquerier.On("GetPrimaryKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetPrimaryKeyConstraintsRow{{
			SchemaName:     "public",
			TableName:      "users",
			ConstraintName: "pk-id",
			ColumnName:     "id",
		}}, nil)
	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.BenthosConfigs)
	assert.Len(t, resp.BenthosConfigs, 1)
	bc := resp.BenthosConfigs[0]
	assert.Equal(t, bc.Name, "public.users")
	assert.Empty(t, bc.DependsOn)
	out, err := yaml.Marshal(bc.Config)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(string(out)),
		strings.TrimSpace(`
input:
    label: ""
    sql_select:
        driver: mysql
        dsn: ${SOURCE_CONNECTION_DSN}
        table: public.users
        columns:
            - id
            - name
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_raw:
                driver: mysql
                dsn: ${DESTINATION_0_CONNECTION_DSN}
                query: INSERT INTO public.users (id, name) VALUES (DEFAULT, ?);
                args_mapping: root = [this.name]
                init_statement: ""
                batching:
                    count: 100
                    byte_size: 0
                    period: 5s
                    check: ""
                    processors: []
`),
	)
	// create a new streambuilder instance so we can access the SetYaml method
	newSB := sb.NewStreamBuilder()

	// SetYAML parses a full Benthos config and uses it to configure the builder.
	err = newSB.SetYAML(string(out))
	assert.NoError(t, err)
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
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{
		"123": pg_queries.NewMockDBTX(t),
		"456": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	tableMappings := []*TableMapping{
		{Schema: "public",
			Table: "users",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "users",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: "generate_default",
					},
				},
				{
					Schema: "public",
					Table:  "users",
					Column: "name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: "",
					},
				},
			},
		}}

	sourceTableOpts := map[string]*sqlSourceTableOptions{"where": {WhereClause: &dsn}}

	groupedSchemas := map[string]map[string]*dbschemas_utils.ColumnInfo{
		"public.users": {
			"id": &dbschemas_utils.ColumnInfo{
				OrdinalPosition:        1,
				ColumnDefault:          "324",
				IsNullable:             "false",
				DataType:               "",
				CharacterMaximumLength: nil,
				NumericPrecision:       nil,
				NumericScale:           nil,
			},
		},
	}

	res, err := bbuilder.buildBenthosSqlSourceConfigResponses(context.Background(), tableMappings, dsn, driver, sourceTableOpts, groupedSchemas)
	assert.Nil(t, err)
	assert.Empty(t, res[0].Config.StreamConfig.Pipeline.Processors)

}
func Test_ProcessorConfigEmptyJavascript(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := map[string]pg_queries.DBTX{
		"123": pg_queries.NewMockDBTX(t),
		"456": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	tableMappings := []*TableMapping{
		{Schema: "public",
			Table: "users",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "users",
					Column: "id",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: "generate_default",
					},
				},
				{
					Schema: "public",
					Table:  "users",
					Column: "name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: "transform_javascript",
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
								TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: ""},
							},
						},
					},
				},
			},
		}}

	sourceTableOpts := map[string]*sqlSourceTableOptions{"where": {WhereClause: &dsn}}

	groupedSchemas := map[string]map[string]*dbschemas_utils.ColumnInfo{
		"public.users": {
			"id": &dbschemas_utils.ColumnInfo{
				OrdinalPosition:        1,
				ColumnDefault:          "324",
				IsNullable:             "false",
				DataType:               "",
				CharacterMaximumLength: nil,
				NumericPrecision:       nil,
				NumericScale:           nil,
			},
		},
	}

	res, err := bbuilder.buildBenthosSqlSourceConfigResponses(context.Background(), tableMappings, dsn, driver, sourceTableOpts, groupedSchemas)
	assert.Nil(t, err)
	assert.Empty(t, res[0].Config.StreamConfig.Pipeline.Processors)

}

func Test_ProcessorConfigMultiJavascript(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{
		"123": pg_queries.NewMockDBTX(t),
		"456": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	tableMappings := []*TableMapping{
		{Schema: "public",
			Table: "users",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "users",
					Column: "name",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: "transform_javascript",
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
						Source: "transform_javascript",
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
								TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: `var payload = value + " firstname";return payload;`},
							},
						},
					},
				},
			},
		}}

	sourceTableOpts := map[string]*sqlSourceTableOptions{"test": {WhereClause: &dsn}}

	groupedSchemas := map[string]map[string]*dbschemas_utils.ColumnInfo{
		"public.users": {
			"id": &dbschemas_utils.ColumnInfo{
				OrdinalPosition:        1,
				ColumnDefault:          "324",
				IsNullable:             "false",
				DataType:               "",
				CharacterMaximumLength: nil,
				NumericPrecision:       nil,
				NumericScale:           nil,
			},
		},
	}

	res, err := bbuilder.buildBenthosSqlSourceConfigResponses(context.Background(), tableMappings, dsn, driver, sourceTableOpts, groupedSchemas)
	assert.Nil(t, err)

	out, err := yaml.Marshal(res[0].Config.Pipeline.Processors)
	assert.NoError(t, err)
	assert.Equal(
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
      `), strings.TrimSpace(string(out)))
}

func Test_ProcessorConfigMutationAndJavascript(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgcache := map[string]pg_queries.DBTX{
		"123": pg_queries.NewMockDBTX(t),
		"456": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockTransformerClient, mockSqlConnector)

	tableMappings := []*TableMapping{
		{Schema: "public",
			Table: "users",
			Mappings: []*mgmtv1alpha1.JobMapping{
				{
					Schema: "public",
					Table:  "users",
					Column: "email",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: "generate_email",
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
						Source: "transform_javascript",
						Config: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
								TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: `var payload = value + " firstname";return payload;`},
							},
						},
					},
				},
			},
		}}

	sourceTableOpts := map[string]*sqlSourceTableOptions{"test": {WhereClause: &dsn}}

	var email int32 = int32(40)

	groupedSchemas := map[string]map[string]*dbschemas_utils.ColumnInfo{
		"public.users": {
			"email": &dbschemas_utils.ColumnInfo{
				OrdinalPosition:        2,
				ColumnDefault:          "",
				IsNullable:             "true",
				DataType:               "timestamptz",
				CharacterMaximumLength: &email,
				NumericPrecision:       nil,
				NumericScale:           nil,
			},
		},
	}

	res, err := bbuilder.buildBenthosSqlSourceConfigResponses(context.Background(), tableMappings, dsn, driver, sourceTableOpts, groupedSchemas)

	assert.Nil(t, err)

	assert.Len(t, res[0].Config.Pipeline.Processors, 2)

	out, err := yaml.Marshal(res[0].Config.Pipeline.Processors)
	assert.NoError(t, err)
	assert.Equal(
		t,
		strings.TrimSpace(`
- mutation: root.email = generate_email(max_length:40)
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
	    `), strings.TrimSpace(string(out)))

}
