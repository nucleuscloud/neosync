package datasync_activities

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	mysql_queries "github.com/nucleuscloud/neosync/worker/gen/go/db/mysql"
	pg_queries "github.com/nucleuscloud/neosync/worker/gen/go/db/postgresql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/log"
	"gopkg.in/yaml.v3"
)

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Generate_Pg(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)

	pgcache := map[string]pg_queries.DBTX{
		"fake-prod-url":  pg_queries.NewMockDBTX(t),
		"fake-stage-url": pg_queries.NewMockDBTX(t),
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
						Transformer: &mgmtv1alpha1.Transformer{
							Value: "generate_uuid",
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
						Transformer: &mgmtv1alpha1.Transformer{
							Value: "generate_full_name",
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
									GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
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
	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient)
	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		log.NewStructuredLogger(slog.Default()),
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
            root.name = generate_full_name()
        interval: ""
        count: 10
        batch_size: 0
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_insert:
                driver: postgres
                dsn: fake-stage-url
                table: public.users
                columns:
                    - id
                    - name
                args_mapping: root = [this.id, this.name]
                init_statement: ""
                conn_max_idle: 2
                conn_max_open: 2
                batching:
                    count: 32767
                    byte_size: 0
                    period: 1s
                    check: ""
                    processors: []
`),
	)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Pg_Pg(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)

	pgcache := map[string]pg_queries.DBTX{
		"fake-prod-url":  pg_queries.NewMockDBTX(t),
		"fake-stage-url": pg_queries.NewMockDBTX(t),
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
						Transformer: &mgmtv1alpha1.Transformer{
							Value: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.Transformer{
							Value: "passthrough",
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
	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		log.NewStructuredLogger(slog.Default()),
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
        dsn: fake-prod-url
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
            - sql_insert:
                driver: postgres
                dsn: fake-stage-url
                table: public.users
                columns:
                    - id
                    - name
                args_mapping: root = [this.id, this.name]
                init_statement: ""
                conn_max_idle: 2
                conn_max_open: 2
                batching:
                    count: 32767
                    byte_size: 0
                    period: 1s
                    check: ""
                    processors: []
`),
	)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Generate_Mysql(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)

	pgcache := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{
		"fake-prod-url":  mysql_queries.NewMockDBTX(t),
		"fake-stage-url": mysql_queries.NewMockDBTX(t),
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
						Transformer: &mgmtv1alpha1.Transformer{
							Value: "generate_uuid",
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
						Transformer: &mgmtv1alpha1.Transformer{
							Value: "generate_full_name",
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
									GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
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

	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		log.NewStructuredLogger(slog.Default()),
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
            root.name = generate_full_name()
        interval: ""
        count: 10
        batch_size: 0
buffer: null
pipeline:
    threads: -1
    processors: []
output:
    label: ""
    broker:
        pattern: fan_out
        outputs:
            - sql_insert:
                driver: mysql
                dsn: fake-stage-url
                table: public.users
                columns:
                    - id
                    - name
                args_mapping: root = [this.id, this.name]
                init_statement: ""
                conn_max_idle: 2
                conn_max_open: 2
                batching:
                    count: 32767
                    byte_size: 0
                    period: 1s
                    check: ""
                    processors: []
`),
	)
}

func Test_BenthosBuilder_GenerateBenthosConfigs_Basic_Mysql_Mysql(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)

	pgcache := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{
		"fake-prod-url":  mysql_queries.NewMockDBTX(t),
		"fake-stage-url": mysql_queries.NewMockDBTX(t),
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
						Transformer: &mgmtv1alpha1.Transformer{
							Value: "passthrough",
						},
					},
					{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.Transformer{
							Value: "passthrough",
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
	bbuilder := newBenthosBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient)

	resp, err := bbuilder.GenerateBenthosConfigs(
		context.Background(),
		&GenerateBenthosConfigsRequest{JobId: "123", WorkflowId: "123"},
		log.NewStructuredLogger(slog.Default()),
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
        dsn: fake-prod-url
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
            - sql_insert:
                driver: mysql
                dsn: fake-stage-url
                table: public.users
                columns:
                    - id
                    - name
                args_mapping: root = [this.id, this.name]
                init_statement: ""
                conn_max_idle: 2
                conn_max_open: 2
                batching:
                    count: 32767
                    byte_size: 0
                    period: 1s
                    check: ""
                    processors: []
`),
	)
}
