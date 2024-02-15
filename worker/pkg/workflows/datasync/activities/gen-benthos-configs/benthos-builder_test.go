package genbenthosconfigs_activity

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/benthosdev/benthos/v4/public/bloblang"
	sb "github.com/benthosdev/benthos/v4/public/service"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	pg_models "github.com/nucleuscloud/neosync/backend/sql/postgresql/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v3"

	_ "github.com/benthosdev/benthos/v4/public/components/aws"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	_ "github.com/benthosdev/benthos/v4/public/components/javascript"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/benthosdev/benthos/v4/public/components/pure/extended"
	_ "github.com/benthosdev/benthos/v4/public/components/sql"
	_ "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"

	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
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
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	tableMappings := []*tableMapping{
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

	res, err := buildBenthosSqlSourceConfigResponses(context.Background(), mockTransformerClient, tableMappings, dsn, driver, sourceTableOpts, groupedSchemas)
	assert.Nil(t, err)
	assert.Empty(t, res[0].Config.StreamConfig.Pipeline.Processors)
}
func Test_ProcessorConfigEmptyJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	tableMappings := []*tableMapping{
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

	res, err := buildBenthosSqlSourceConfigResponses(context.Background(), mockTransformerClient, tableMappings, dsn, driver, sourceTableOpts, groupedSchemas)
	assert.Nil(t, err)
	assert.Empty(t, res[0].Config.StreamConfig.Pipeline.Processors)
}

func Test_ProcessorConfigMultiJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	tableMappings := []*tableMapping{
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

	res, err := buildBenthosSqlSourceConfigResponses(context.Background(), mockTransformerClient, tableMappings, dsn, driver, sourceTableOpts, groupedSchemas)
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
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	tableMappings := []*tableMapping{
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

	res, err := buildBenthosSqlSourceConfigResponses(context.Background(), mockTransformerClient, tableMappings, dsn, driver, sourceTableOpts, groupedSchemas)

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

func TestAreMappingsSubsetOfSchemas(t *testing.T) {
	ok := areMappingsSubsetOfSchemas(
		map[string]map[string]*dbschemas_utils.ColumnInfo{
			"public.users": {
				"id":         &dbschemas_utils.ColumnInfo{},
				"created_by": &dbschemas_utils.ColumnInfo{},
				"updated_by": &dbschemas_utils.ColumnInfo{},
			},
			"neosync_api.accounts": {
				"id": &dbschemas_utils.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	assert.True(t, ok, "job mappings are a subset of the present database schemas")

	ok = areMappingsSubsetOfSchemas(
		map[string]map[string]*dbschemas_utils.ColumnInfo{
			"public.users": {
				"id": &dbschemas_utils.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id2"},
		},
	)
	assert.False(t, ok, "job mappings contain mapping that is not in the source schema")

	ok = areMappingsSubsetOfSchemas(
		map[string]map[string]*dbschemas_utils.ColumnInfo{
			"public.users": {
				"id": &dbschemas_utils.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	assert.False(t, ok, "job mappings contain more mappings than are present in the source schema")
}

func TestShouldHaltOnSchemaAddition(t *testing.T) {
	ok := shouldHaltOnSchemaAddition(
		map[string]map[string]*dbschemas_utils.ColumnInfo{
			"public.users": {
				"id":         &dbschemas_utils.ColumnInfo{},
				"created_by": &dbschemas_utils.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	assert.False(t, ok, "job mappings are valid set of database schemas")

	ok = shouldHaltOnSchemaAddition(
		map[string]map[string]*dbschemas_utils.ColumnInfo{
			"public.users": {
				"id":         &dbschemas_utils.ColumnInfo{},
				"created_by": &dbschemas_utils.ColumnInfo{},
			},
			"neosync_api.accounts": {
				"id": &dbschemas_utils.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "created_by"},
		},
	)
	assert.True(t, ok, "job mappings are missing database schema mappings")

	ok = shouldHaltOnSchemaAddition(
		map[string]map[string]*dbschemas_utils.ColumnInfo{
			"public.users": {
				"id":         &dbschemas_utils.ColumnInfo{},
				"created_by": &dbschemas_utils.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
		},
	)
	assert.True(t, ok, "job mappings are missing table column")

	ok = shouldHaltOnSchemaAddition(
		map[string]map[string]*dbschemas_utils.ColumnInfo{
			"public.users": {
				"id":         &dbschemas_utils.ColumnInfo{},
				"created_by": &dbschemas_utils.ColumnInfo{},
			},
		},
		[]*mgmtv1alpha1.JobMapping{
			{Schema: "public", Table: "users", Column: "id"},
			{Schema: "public", Table: "users", Column: "updated_by"},
		},
	)
	assert.True(t, ok, "job mappings have same column count, but missing specific column")
}

func Test_buildProcessorConfigsMutation(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	ctx := context.Background()

	output, err := buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{}, map[string]*dbschemas_utils.ColumnInfo{})
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{}, map[string]*dbschemas_utils.ColumnInfo{})
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id"},
	}, map[string]*dbschemas_utils.ColumnInfo{})
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{}},
	}, map[string]*dbschemas_utils.ColumnInfo{})
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: "passthrough"}},
	}, map[string]*dbschemas_utils.ColumnInfo{})
	assert.Nil(t, err)
	assert.Empty(t, output)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: "null", Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		}}},
		{Schema: "public", Table: "users", Column: "name", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: "null", Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		}}},
	}, map[string]*dbschemas_utils.ColumnInfo{})

	assert.Nil(t, err)

	assert.Equal(t, *output[0].Mutation, "root.id = null\nroot.name = null")

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    "string",
		Source:      "transform_email",
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

	groupedSchemas := map[string]*dbschemas_utils.ColumnInfo{

		"email": {
			OrdinalPosition:        2,
			ColumnDefault:          "",
			IsNullable:             "true",
			DataType:               "timestamptz",
			CharacterMaximumLength: &email,
			NumericPrecision:       nil,
			NumericScale:           nil,
		},
	}

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "email", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}}}, groupedSchemas)

	assert.Nil(t, err)
	assert.Equal(t, *output[0].Mutation, `root.email = transform_email(email:this.email,preserve_domain:true,preserve_length:false,excluded_domains:[],max_length:40)`)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: "i_do_not_exist", Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		}}},
	}, map[string]*dbschemas_utils.ColumnInfo{})
	assert.Error(t, err)
	assert.Empty(t, output)
}

const defaultJavascriptCodeFnStr = `var payload = value+=" hello";return payload;`

func Test_buildProcessorConfigsJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	ctx := context.Background()

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    "string",
		Source:      "transform_javascript",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: defaultJavascriptCodeFnStr,
				},
			},
		},
	}

	res, err := buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "address", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}}}, map[string]*dbschemas_utils.ColumnInfo{})

	assert.NoError(t, err)
	assert.Equal(t, `
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
		DataType:    "string",
		Source:      "transform_javascript",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: code,
				},
			},
		},
	}

	res, err := buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: nameCol, Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}}}, map[string]*dbschemas_utils.ColumnInfo{})

	assert.NoError(t, err)
	assert.Equal(t, `
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
		DataType:    "string",
		Source:      "transform_javascript",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: defaultJavascriptCodeFnStr,
				},
			},
		},
	}

	jsT2 := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    "string",
		Source:      "transform_javascript",
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
		{Schema: "public", Table: "users", Column: col2, Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT2.Source, Config: jsT2.Config}}}, map[string]*dbschemas_utils.ColumnInfo{})

	assert.NoError(t, err)
	assert.Equal(t, `
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

func Test_ShouldProcessColumnTrue(t *testing.T) {
	val := &mgmtv1alpha1.JobMappingTransformer{
		Source: "generate_email",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		},
	}

	res := shouldProcessColumn(val)
	assert.Equal(t, true, res)
}

func Test_ShouldProcessColumnFalse(t *testing.T) {
	val := &mgmtv1alpha1.JobMappingTransformer{
		Source: "passthrough",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
				PassthroughConfig: &mgmtv1alpha1.Passthrough{},
			},
		},
	}

	res := shouldProcessColumn(val)
	assert.Equal(t, false, res)
}

func Test_ConstructJsFunction(t *testing.T) {
	res := constructJsFunction(defaultJavascriptCodeFnStr, "col")
	assert.Equal(t, `
function fn_col(value, input){
  var payload = value+=" hello";return payload;
};
`, res)
}

func Test_ConstructBenthosJsProcessor(t *testing.T) {
	jsFunctions := []string{}
	benthosOutputs := []string{}

	benthosOutput := constructBenthosOutput(nameCol)
	jsFunction := constructJsFunction(defaultJavascriptCodeFnStr, nameCol)
	benthosOutputs = append(benthosOutputs, benthosOutput)

	jsFunctions = append(jsFunctions, jsFunction)

	res := constructBenthosJsProcessor(jsFunctions, benthosOutputs)

	assert.Equal(t, `
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

func Test_ConstructBenthosOutput(t *testing.T) {
	res := constructBenthosOutput("col")
	assert.Equal(t, `output["col"] = fn_col(input["col"], input);`, res)
}

func Test_buildProcessorConfigsJavascriptEmpty(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	ctx := context.Background()

	jsT := mgmtv1alpha1.SystemTransformer{
		Name:        "stage",
		Description: "description",
		DataType:    "string",
		Source:      "transform_javascript",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: ``,
				},
			},
		},
	}

	resp, err := buildProcessorConfigs(ctx, mockTransformerClient, []*mgmtv1alpha1.JobMapping{
		{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Source: jsT.Source, Config: jsT.Config}}}, map[string]*dbschemas_utils.ColumnInfo{})

	assert.NoError(t, err)
	assert.Empty(t, resp)
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
			DataType:    "string",
			Source:      "transform_email",
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
		Source: "transform_email",
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig{
				UserDefinedTransformerConfig: &mgmtv1alpha1.UserDefinedTransformerConfig{
					Id: "123",
				},
			},
		},
	}

	expected := &mgmtv1alpha1.JobMappingTransformer{
		Source: "transform_email",
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
	assert.NoError(t, err)
	assert.Equal(t, resp, expected)
}

func MockJobMappingTransformer(source, transformerId string) db_queries.NeosyncApiTransformer {
	return db_queries.NeosyncApiTransformer{
		Source:            source,
		TransformerConfig: &pg_models.TransformerConfigs{},
	}
}

func Test_buildPlainInsertArgs(t *testing.T) {
	assert.Empty(t, buildPlainInsertArgs(nil))
	assert.Empty(t, buildPlainInsertArgs([]string{}))
	assert.Equal(t, buildPlainInsertArgs([]string{"foo", "bar", "baz"}), "root = [this.foo, this.bar, this.baz]")
}

func Test_buildPlainColumns(t *testing.T) {
	assert.Empty(t, buildPlainColumns(nil))
	assert.Empty(t, buildPlainColumns([]*mgmtv1alpha1.JobMapping{}))
	assert.Equal(
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
	assert.Equal(t, schema, "public")
	assert.Equal(t, table, "foo")

	schema, table = splitTableKey("neosync.foo")
	assert.Equal(t, schema, "neosync")
	assert.Equal(t, table, "foo")
}

func Test_buildBenthosS3Credentials(t *testing.T) {
	assert.Nil(t, buildBenthosS3Credentials(nil))

	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{}),
		&neosync_benthos.AwsCredentials{},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{Profile: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Profile: "foo"},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{AccessKeyId: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Id: "foo"},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{SecretAccessKey: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Secret: "foo"},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{SessionToken: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Token: "foo"},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{FromEc2Role: shared.Ptr(true)}),
		&neosync_benthos.AwsCredentials{FromEc2Role: true},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{RoleArn: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Role: "foo"},
	)
	assert.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{RoleExternalId: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{RoleExternalId: "foo"},
	)
	assert.Equal(
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
				Source: "null",
			},
		}, &dbschemas_utils.ColumnInfo{})
	assert.NoError(t, err)
	assert.Equal(t, val, "null")
}

func Test_computeMutationFunction_Validate_Bloblang_Output(t *testing.T) {
	transformers := []*mgmtv1alpha1.SystemTransformer{
		{
			Source: "generate_email",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
					GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{},
				},
			},
		},
		{
			Source: "transform_email",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
					TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
						PreserveDomain:  false,
						PreserveLength:  false,
						ExcludedDomains: []string{},
					},
				},
			},
		},
		{
			Source: "generate_bool",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
					GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
				},
			},
		},
		{
			Source: "generate_card_number",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
					GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{
						ValidLuhn: true,
					},
				},
			},
		},
		{
			Source: "generate_city",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
					GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
				},
			},
		},
		{
			Source: "generate_e164_phone_number",
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
			Source: "generate_first_name",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
					GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
				},
			},
		},
		{
			Source: "generate_float64",
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
			Source: "generate_full_address",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig{
					GenerateFullAddressConfig: &mgmtv1alpha1.GenerateFullAddress{},
				},
			},
		},
		{
			Source: "generate_full_name",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		},
		{
			Source: "generate_gender",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{
					GenerateGenderConfig: &mgmtv1alpha1.GenerateGender{
						Abbreviate: false,
					},
				},
			},
		},
		{
			Source: "generate_int64_phone_number",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig{
					GenerateInt64PhoneNumberConfig: &mgmtv1alpha1.GenerateInt64PhoneNumber{},
				},
			},
		},
		{
			Source: "generate_int64",
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
			Source: "generate_last_name",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{
					GenerateLastNameConfig: &mgmtv1alpha1.GenerateLastName{},
				},
			},
		},
		{
			Source: "generate_sha256hash",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{
					GenerateSha256HashConfig: &mgmtv1alpha1.GenerateSha256Hash{},
				},
			},
		},
		{
			Source: "generate_ssn",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
					GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
				},
			},
		},
		{
			Source: "generate_state",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{
					GenerateStateConfig: &mgmtv1alpha1.GenerateState{},
				},
			},
		},
		{
			Source: "generate_street_address",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig{
					GenerateStreetAddressConfig: &mgmtv1alpha1.GenerateStreetAddress{},
				},
			},
		},
		{
			Source: "generate_string_phone_number",
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
			Source: "generate_string",
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
			Source: "generate_unixtimestamp",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{
					GenerateUnixtimestampConfig: &mgmtv1alpha1.GenerateUnixTimestamp{},
				},
			},
		},
		{
			Source: "generate_username",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig{
					GenerateUsernameConfig: &mgmtv1alpha1.GenerateUsername{},
				},
			},
		},
		{
			Source: "generate_utctimestamp",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig{
					GenerateUtctimestampConfig: &mgmtv1alpha1.GenerateUtcTimestamp{},
				},
			},
		},
		{
			Source: "generate_uuid",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
					GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
						IncludeHyphens: true,
					},
				},
			},
		},
		{
			Source: "generate_zipcode",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig{
					GenerateZipcodeConfig: &mgmtv1alpha1.GenerateZipcode{},
				},
			},
		},
		{
			Source: "transform_e164_phone_number",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig{
					TransformE164PhoneNumberConfig: &mgmtv1alpha1.TransformE164PhoneNumber{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: "transform_first_name",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
					TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: "transform_float64",
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
			Source: "transform_full_name",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
					TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: "transform_int64_phone_number",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig{
					TransformInt64PhoneNumberConfig: &mgmtv1alpha1.TransformInt64PhoneNumber{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: "transform_int64",
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
			Source: "transform_last_name",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{
					TransformLastNameConfig: &mgmtv1alpha1.TransformLastName{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: "transform_phone_number",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig{
					TransformPhoneNumberConfig: &mgmtv1alpha1.TransformPhoneNumber{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: "transform_string",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
					TransformStringConfig: &mgmtv1alpha1.TransformString{
						PreserveLength: false,
					},
				},
			},
		},
		{
			Source: "generate_categorical",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig{
					GenerateCategoricalConfig: &mgmtv1alpha1.GenerateCategorical{
						Categories: "value1,value2",
					},
				},
			},
		},
		{
			Source: "transform_character_scramble",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
					TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{
						Regex: "",
					},
				},
			},
		},
		{
			Source: "generate_default",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{
					GenerateDefaultConfig: &mgmtv1alpha1.GenerateDefault{},
				},
			},
		},
		{
			Source: "null",
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
					Nullconfig: &mgmtv1alpha1.Null{},
				},
			},
		},
	}

	emailColInfo := &dbschemas_utils.ColumnInfo{
		OrdinalPosition:        2,
		ColumnDefault:          "",
		IsNullable:             "true",
		DataType:               "timestamptz",
		CharacterMaximumLength: shared.Ptr(int32(40)),
		NumericPrecision:       nil,
		NumericScale:           nil,
	}

	for _, transformer := range transformers {
		t.Run(fmt.Sprintf("%s_lint", transformer.Source), func(t *testing.T) {
			val, err := computeMutationFunction(
				&mgmtv1alpha1.JobMapping{
					Column: "email",
					Transformer: &mgmtv1alpha1.JobMappingTransformer{
						Source: transformer.Source,
						Config: transformer.Config,
					},
				}, emailColInfo)

			assert.NoError(t, err)
			_, err = bloblang.Parse(val)
			assert.NoError(t, err, fmt.Sprintf("transformer lint failed, check that the transformer string is being constructed correctly. Failing source: %s", transformer.Source))
		})
	}
}
