package runsqlinittablestmts_activity

import (
	"context"
	"log/slog"
	"slices"
	"testing"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_InitStatementBuilder_Pg_Generate(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgPoolContainerMock := sqlconnect.NewMockPgPoolContainer(t)

	dbtx := pg_queries.NewMockDBTX(t)
	pgcache := map[string]pg_queries.DBTX{
		"fake-prod-url": pg_queries.NewMockDBTX(t),
		"postgresql://postgres:foofar@localhost:5435/nucleus": dbtx,
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	connectionId := "456"

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
								FkSourceConnectionId: &connectionId,
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
							Source: "generate_full_name",
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
						Options: &mgmtv1alpha1.JobDestinationOptions{
							Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
								PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
									TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
										TruncateBeforeInsert: true,
										Cascade:              true,
									},
									InitTableSchema: true,
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
							Url: "postgresql://postgres:foofar@localhost:5435/nucleus",
						},
					},
				},
			},
		},
	}), nil)
	var cmdtag pgconn.CommandTag
	dbtx.On("Exec", mock.Anything, `TRUNCATE TABLE "public"."users" CASCADE;`).Return(cmdtag, nil)
	pgPoolContainerMock.On("Open", mock.Anything).Return(dbtx, nil)
	pgPoolContainerMock.On("Close")
	mockSqlConnector.On("NewPgPoolFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(pgPoolContainerMock, nil)

	bbuilder := newInitStatementBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockSqlConnector)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
}

func Test_InitStatementBuilder_Pg_Generate_NoInitStatement(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)

	pgcache := map[string]pg_queries.DBTX{
		"123": pg_queries.NewMockDBTX(t),
		"456": pg_queries.NewMockDBTX(t),
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	connectionId := "456"

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
								FkSourceConnectionId: &connectionId,
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
							Source: "generate_full_name",
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

	bbuilder := newInitStatementBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockSqlConnector)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
}

func Test_InitStatementBuilder_Pg_TruncateCascade(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgPoolContainerMock := sqlconnect.NewMockPgPoolContainer(t)

	dbtx := pg_queries.NewMockDBTX(t)
	pgcache := map[string]pg_queries.DBTX{
		"fake-prod-url": pg_queries.NewMockDBTX(t),
		"postgresql://postgres:foofar@localhost:5435/nucleus": dbtx,
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	connectionId := "456"

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
							Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
								Schemas: []*mgmtv1alpha1.PostgresSourceSchemaOption{
									{
										Schema: "public",
										Tables: []*mgmtv1alpha1.PostgresSourceTableOption{
											{
												Table: "users",
											},
											{
												Table: "accounts",
											},
										},
									},
								},
								ConnectionId: connectionId,
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
							Source: "generate_full_name",
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
									GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
								},
							},
						},
					},
					{
						Schema: "public",
						Table:  "accounts",
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
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
						Options: &mgmtv1alpha1.JobDestinationOptions{
							Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
								PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
									TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
										TruncateBeforeInsert: true,
										Cascade:              true,
									},
									InitTableSchema: false,
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
							Url: "postgresql://postgres:foofar@localhost:5435/nucleus",
						},
					},
				},
			},
		},
	}), nil)
	pgquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetForeignKeyConstraintsRow{}, nil)

	var cmdtag pgconn.CommandTag
	allowedQueries := []string{
		"TRUNCATE TABLE \"public\".\"users\" CASCADE;\nTRUNCATE TABLE \"public\".\"accounts\" CASCADE;",
		"TRUNCATE TABLE \"public\".\"accounts\" CASCADE;\nTRUNCATE TABLE \"public\".\"users\" CASCADE;",
	}
	dbtx.On("Exec", mock.Anything, mock.MatchedBy(func(query string) bool { return slices.Contains(allowedQueries, query) })).Return(cmdtag, nil)

	pgPoolContainerMock.On("Open", mock.Anything).Return(dbtx, nil)
	pgPoolContainerMock.On("Close")
	mockSqlConnector.On("NewPgPoolFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(pgPoolContainerMock, nil)

	bbuilder := newInitStatementBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockSqlConnector)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
}

func Test_InitStatementBuilder_Pg_Truncate(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgPoolContainerMock := sqlconnect.NewMockPgPoolContainer(t)

	dbtx := pg_queries.NewMockDBTX(t)
	pgcache := map[string]pg_queries.DBTX{
		"fake-prod-url": pg_queries.NewMockDBTX(t),
		"postgresql://postgres:foofar@localhost:5435/nucleus": dbtx,
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	connectionId := "456"

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
							Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
								Schemas: []*mgmtv1alpha1.PostgresSourceSchemaOption{
									{
										Schema: "public",
										Tables: []*mgmtv1alpha1.PostgresSourceTableOption{
											{
												Table: "users",
											},
											{
												Table: "accounts",
											},
										},
									},
								},
								ConnectionId: connectionId,
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
							Source: "generate_full_name",
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
									GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
								},
							},
						},
					},
					{
						Schema: "public",
						Table:  "accounts",
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
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
						Options: &mgmtv1alpha1.JobDestinationOptions{
							Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
								PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
									TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
										TruncateBeforeInsert: true,
										Cascade:              false,
									},
									InitTableSchema: false,
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
							Url: "postgresql://postgres:foofar@localhost:5435/nucleus",
						},
					},
				},
			},
		},
	}), nil)
	pgquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetForeignKeyConstraintsRow{
			{
				ConstraintName:    "fk_user_account_id",
				SchemaName:        "public",
				TableName:         "users",
				ColumnName:        "account_id",
				ForeignSchemaName: "public",
				ForeignTableName:  "accounts",
				ForeignColumnName: "id",
			},
		}, nil)

	var cmdtag pgconn.CommandTag
	dbtx.On("Exec", mock.Anything, "TRUNCATE TABLE \"public\".\"accounts\", \"public\".\"users\";").Return(cmdtag, nil)
	pgPoolContainerMock.On("Open", mock.Anything).Return(dbtx, nil)
	pgPoolContainerMock.On("Close")
	mockSqlConnector.On("NewPgPoolFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(pgPoolContainerMock, nil)

	bbuilder := newInitStatementBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockSqlConnector)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
}

func Test_InitStatementBuilder_Pg_InitSchema(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	pgPoolContainerMock := sqlconnect.NewMockPgPoolContainer(t)

	dbtx := pg_queries.NewMockDBTX(t)
	pgcache := map[string]pg_queries.DBTX{
		"fake-prod-url": pg_queries.NewMockDBTX(t),
		"postgresql://postgres:foofar@localhost:5435/nucleus": dbtx,
	}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	connectionId := "456"

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
							Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
								Schemas: []*mgmtv1alpha1.PostgresSourceSchemaOption{
									{
										Schema: "public",
										Tables: []*mgmtv1alpha1.PostgresSourceTableOption{
											{
												Table: "users",
											},
											{
												Table: "accounts",
											},
										},
									},
								},
								ConnectionId: connectionId,
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
							Source: "generate_full_name",
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
									GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
								},
							},
						},
					},
					{
						Schema: "public",
						Table:  "accounts",
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
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
						Options: &mgmtv1alpha1.JobDestinationOptions{
							Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
								PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
									TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
										TruncateBeforeInsert: false,
										Cascade:              false,
									},
									InitTableSchema: true,
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
							Url: "postgresql://postgres:foofar@localhost:5435/nucleus",
						},
					},
				},
			},
		},
	}), nil)
	pgquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*pg_queries.GetForeignKeyConstraintsRow{
			{
				ConstraintName:    "fk_user_account_id",
				SchemaName:        "public",
				TableName:         "users",
				ColumnName:        "account_id",
				ForeignSchemaName: "public",
				ForeignTableName:  "accounts",
				ForeignColumnName: "id",
			},
		}, nil)
	pgquerier.On("GetTableConstraints", mock.Anything, mock.Anything, &pg_queries.GetTableConstraintsParams{
		Schema: "public",
		Table:  "users",
	}).Return([]*pg_queries.GetTableConstraintsRow{
		{
			ConstraintName:       "users_pkey",
			ConstraintDefinition: "PRIMARY KEY (id)",
		},
		{
			ConstraintName:       "accounts_pkey",
			ConstraintDefinition: "PRIMARY KEY (id)",
		},
	}, nil)
	pgquerier.On("GetDatabaseTableSchema", mock.Anything, mock.Anything, &pg_queries.GetDatabaseTableSchemaParams{
		Schema: "public",
		Table:  "users",
	}).Return([]*pg_queries.GetDatabaseTableSchemaRow{
		{
			ColumnName:      "id",
			DataType:        "uuid",
			OrdinalPosition: 1,
			IsNullable:      "NO",
			ColumnDefault:   "gen_random_uuid()",
		},
		{
			ColumnName:             "account_id",
			DataType:               "uuid",
			OrdinalPosition:        1,
			IsNullable:             "No",
			CharacterMaximumLength: 40,
		},
	}, nil)
	pgquerier.On("GetDatabaseTableSchema", mock.Anything, mock.Anything, &pg_queries.GetDatabaseTableSchemaParams{
		Schema: "public",
		Table:  "accounts",
	}).Return([]*pg_queries.GetDatabaseTableSchemaRow{
		{
			ColumnName:      "id",
			DataType:        "uuid",
			OrdinalPosition: 1,
			IsNullable:      "NO",
			ColumnDefault:   "gen_random_uuid()",
		},
	}, nil)

	pgquerier.On("GetTableConstraints", mock.Anything, mock.Anything, &pg_queries.GetTableConstraintsParams{
		Schema: "public",
		Table:  "users",
	}).Return([]*pg_queries.GetTableConstraintsRow{
		{
			ConstraintName:       "users_pkey",
			ConstraintDefinition: "PRIMARY KEY (id)",
		},
	}, nil)
	pgquerier.On("GetTableConstraints", mock.Anything, mock.Anything, &pg_queries.GetTableConstraintsParams{
		Schema: "public",
		Table:  "accounts",
	}).Return([]*pg_queries.GetTableConstraintsRow{
		{
			ConstraintName:       "accounts_pkey",
			ConstraintDefinition: "PRIMARY KEY (id)",
		},
	}, nil)

	var cmdtag pgconn.CommandTag
	dbtx.On("Exec", mock.Anything, "CREATE TABLE IF NOT EXISTS \"public\".\"accounts\" (\"id\" uuid NOT NULL DEFAULT gen_random_uuid(), CONSTRAINT accounts_pkey PRIMARY KEY (id));\nCREATE TABLE IF NOT EXISTS \"public\".\"users\" (\"id\" uuid NOT NULL DEFAULT gen_random_uuid(), \"account_id\" uuid NULL, CONSTRAINT users_pkey PRIMARY KEY (id), CONSTRAINT accounts_pkey PRIMARY KEY (id));").Return(cmdtag, nil)
	pgPoolContainerMock.On("Open", mock.Anything).Return(dbtx, nil)
	pgPoolContainerMock.On("Close")
	mockSqlConnector.On("NewPgPoolFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(pgPoolContainerMock, nil)

	bbuilder := newInitStatementBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockSqlConnector)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
}

func Test_InitStatementBuilder_Mysql_Truncate(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	sqlDbContainerMock := sqlconnect.NewMockSqlDbContainer(t)

	sqlDbMock, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(false))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	pgcache := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{
		"fake-prod-url": sqlDbMock,
		"postgresql://postgres:foofar@localhost:5435/nucleus": mysql_queries.NewMockDBTX(t),
	}
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	connectionId := "456"

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Mysql{
							Mysql: &mgmtv1alpha1.MysqlSourceConnectionOptions{
								Schemas: []*mgmtv1alpha1.MysqlSourceSchemaOption{
									{
										Schema: "public",
										Tables: []*mgmtv1alpha1.MysqlSourceTableOption{
											{
												Table: "users",
											},
										},
									},
									{
										Schema: "public",
										Tables: []*mgmtv1alpha1.MysqlSourceTableOption{
											{
												Table: "accounts",
											},
										},
									},
								},
								ConnectionId: connectionId,
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
							Source: "generate_full_name",
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
									GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
								},
							},
						},
					},
					{
						Schema: "public",
						Table:  "accounts",
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
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
						Options: &mgmtv1alpha1.JobDestinationOptions{
							Config: &mgmtv1alpha1.JobDestinationOptions_MysqlOptions{
								MysqlOptions: &mgmtv1alpha1.MysqlDestinationConnectionOptions{
									TruncateTable: &mgmtv1alpha1.MysqlTruncateTableConfig{
										TruncateBeforeInsert: true,
									},
									InitTableSchema: false,
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
	mysqlquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetForeignKeyConstraintsRow{
			{
				ConstraintName:    "fk_user_account_id",
				SchemaName:        "public",
				TableName:         "users",
				ColumnName:        "account_id",
				ForeignSchemaName: "public",
				ForeignTableName:  "accounts",
				ForeignColumnName: "id",
			},
		}, nil)
	regexPattern := "TRUNCATE TABLE `?public`?\\.\\`(users|accounts)\\`"
	sqlMock.ExpectExec(regexPattern).WillReturnResult(sqlmock.NewResult(1, 2))
	sqlDbContainerMock.On("Open", mock.Anything).Return(sqlDbMock, nil)
	sqlDbContainerMock.On("Close").Return(nil)
	mockSqlConnector.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(sqlDbContainerMock, nil)

	bbuilder := newInitStatementBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockSqlConnector)
	_, err = bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
}

func Test_InitStatementBuilder_Mysql_InitSchema(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlConnector := sqlconnect.NewMockSqlConnector(t)
	sqlDbContainerMock := sqlconnect.NewMockSqlDbContainer(t)

	sqlDbMock, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(false))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	pgcache := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.NewMockQuerier(t)
	mysqlcache := map[string]mysql_queries.DBTX{
		"fake-prod-url": sqlDbMock,
		"postgresql://postgres:foofar@localhost:5435/nucleus": mysql_queries.NewMockDBTX(t),
	}
	mysqlquerier := mysql_queries.NewMockQuerier(t)
	connectionId := "456"

	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
			Job: &mgmtv1alpha1.Job{
				Source: &mgmtv1alpha1.JobSource{
					Options: &mgmtv1alpha1.JobSourceOptions{
						Config: &mgmtv1alpha1.JobSourceOptions_Mysql{
							Mysql: &mgmtv1alpha1.MysqlSourceConnectionOptions{
								Schemas: []*mgmtv1alpha1.MysqlSourceSchemaOption{
									{
										Schema: "public",
										Tables: []*mgmtv1alpha1.MysqlSourceTableOption{
											{
												Table: "users",
											},
										},
									},
									{
										Schema: "public",
										Tables: []*mgmtv1alpha1.MysqlSourceTableOption{
											{
												Table: "accounts",
											},
										},
									},
								},
								ConnectionId: connectionId,
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
							Source: "generate_full_name",
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
									GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
								},
							},
						},
					},
					{
						Schema: "public",
						Table:  "accounts",
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
				},
				Destinations: []*mgmtv1alpha1.JobDestination{
					{
						ConnectionId: "456",
						Options: &mgmtv1alpha1.JobDestinationOptions{
							Config: &mgmtv1alpha1.JobDestinationOptions_MysqlOptions{
								MysqlOptions: &mgmtv1alpha1.MysqlDestinationConnectionOptions{
									TruncateTable: &mgmtv1alpha1.MysqlTruncateTableConfig{
										TruncateBeforeInsert: false,
									},
									InitTableSchema: true,
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
	mysqlquerier.On("GetForeignKeyConstraints", mock.Anything, mock.Anything, mock.Anything).
		Return([]*mysql_queries.GetForeignKeyConstraintsRow{
			{
				ConstraintName:    "fk_user_account_id",
				SchemaName:        "public",
				TableName:         "users",
				ColumnName:        "account_id",
				ForeignSchemaName: "public",
				ForeignTableName:  "accounts",
				ForeignColumnName: "id",
			},
		}, nil)

	createAccountRows := sqlmock.NewRows([]string{"Table", "Create Table"}).
		AddRow("accounts", "CREATE TABLE public.accounts")
	sqlMock.ExpectQuery("SHOW CREATE TABLE public.accounts;").WillReturnRows(createAccountRows)
	createUsersRows := sqlmock.NewRows([]string{"Table", "Create Table"}).
		AddRow("users", "CREATE TABLE public.users")
	sqlMock.ExpectQuery("SHOW CREATE TABLE public.users;").WillReturnRows(createUsersRows)
	sqlMock.ExpectExec("CREATE TABLE IF NOT EXISTS  public.accounts; CREATE TABLE IF NOT EXISTS  public.users;").WillReturnResult(sqlmock.NewResult(1, 2))
	sqlDbContainerMock.On("Open", mock.Anything).Return(sqlDbMock, nil)
	sqlDbContainerMock.On("Close").Return(nil)
	mockSqlConnector.On("NewDbFromConnectionConfig", mock.Anything, mock.Anything, mock.Anything).Return(sqlDbContainerMock, nil)

	bbuilder := newInitStatementBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient, mockSqlConnector)
	_, err = bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
	if err := sqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_getForeignToPrimaryTableMap(t *testing.T) {
	tables := map[string]struct{}{
		"public.regions":     {},
		"public.jobs":        {},
		"public.countries":   {},
		"public.locations":   {},
		"public.dependents":  {},
		"public.departments": {},
		"public.employees":   {},
	}
	dependencies := map[string]*dbschemas_utils.TableConstraints{
		"public.countries": {Constraints: []*dbschemas_utils.ForeignConstraint{
			{Column: "region_id", IsNullable: false, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.regions", Column: "region_id"}},
		}},
		"public.departments": {Constraints: []*dbschemas_utils.ForeignConstraint{
			{Column: "location_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.locations", Column: "location_id"}},
		}},
		"public.dependents": {Constraints: []*dbschemas_utils.ForeignConstraint{
			{Column: "dependent_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.employees", Column: "employees_id"}},
		}},
		"public.locations": {Constraints: []*dbschemas_utils.ForeignConstraint{
			{Column: "country_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.countries", Column: "country_id"}},
		}},
		"public.employees": {Constraints: []*dbschemas_utils.ForeignConstraint{
			{Column: "department_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.departments", Column: "department_id"}},
			{Column: "job_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.jobs", Column: "job_id"}},
			{Column: "manager_id", IsNullable: true, ForeignKey: &dbschemas_utils.ForeignKey{Table: "public.employees", Column: "employee_id"}},
		}},
	}

	expected := map[string][]string{
		"public.regions":     {},
		"public.jobs":        {},
		"public.countries":   {"public.regions"},
		"public.departments": {"public.locations"},
		"public.dependents":  {"public.employees"},
		"public.employees":   {"public.departments", "public.jobs", "public.employees"},
		"public.locations":   {"public.countries"},
	}
	actual := getForeignToPrimaryTableMap(dependencies, tables)
	assert.Len(t, actual, len(expected))
	for table, deps := range actual {
		assert.Len(t, deps, len(expected[table]))
		assert.ElementsMatch(t, expected[table], deps)
	}
}
