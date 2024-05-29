package runsqlinittablestmts_activity

import (
	"context"
	"log/slog"
	"slices"
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_InitStatementBuilder_Pg_Generate_InitSchema(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)
	connectionId := "456"
	fkconnectionId := "789"

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
								FkSourceConnectionId: &fkconnectionId,
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
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME,
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
						ConnectionId: connectionId,
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
			Id: fkconnectionId,
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   connectionId,
			Name: "prod",
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
	mockConnectionClient.On(
		"GetConnection",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
			Id: connectionId,
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
		Connection: &mgmtv1alpha1.Connection{
			Id:   connectionId,
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
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb}, nil)
	mockSqlDb.On("GetTableInitStatements", mock.Anything, []*sql_manager.SchemaTable{{Schema: "public", Table: "users"}}).Return([]*sql_manager.TableInitStatement{
		{
			CreateTableStatement: "test-create-statement",
			AlterTableStatements: []*sql_manager.AlterTableStatement{
				{
					Statement:      "test-pk-statement",
					ConstraintType: sql_manager.PrimaryConstraintType,
				},
				{
					Statement:      "test-fk-statement",
					ConstraintType: sql_manager.ForeignConstraintType,
				},
			},
			IndexStatements: []string{"test-idx-statement"},
		},
	}, nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"TRUNCATE \"public\".\"users\" CASCADE;"}, &sql_manager.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-create-statement"}, &sql_manager.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-pk-statement"}, &sql_manager.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-fk-statement"}, &sql_manager.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-idx-statement"}, &sql_manager.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("Close").Return(nil)

	bbuilder := newInitStatementBuilder(mockSqlManager, mockJobClient, mockConnectionClient)
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
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)
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
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME,
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
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb}, nil)
	mockSqlDb.On("Close").Return(nil)

	bbuilder := newInitStatementBuilder(mockSqlManager, mockJobClient, mockConnectionClient)
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
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)
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
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME,
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
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb}, nil)
	stmts := []string{"TRUNCATE \"public\".\"users\" CASCADE;", "TRUNCATE \"public\".\"accounts\" CASCADE;"}
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, mock.MatchedBy(func(query []string) bool { return compareSlices(query, stmts) }), &sql_manager.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("Close").Return(nil)

	bbuilder := newInitStatementBuilder(mockSqlManager, mockJobClient, mockConnectionClient)
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
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)
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
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME,
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
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb}, nil)
	mockSqlDb.On("GetForeignKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]*sql_manager.ForeignConstraint{
		"public.users": {{
			Columns:     []string{"account_id"},
			NotNullable: []bool{true},
			ForeignKey:  &sql_manager.ForeignKey{Table: "public.accounts", Columns: []string{"id"}},
		}},
	}, nil)
	mockSqlDb.On("Exec", mock.Anything, "TRUNCATE TABLE \"public\".\"accounts\", \"public\".\"users\";").Return(nil)
	mockSqlDb.On("Close").Return(nil)

	bbuilder := newInitStatementBuilder(mockSqlManager, mockJobClient, mockConnectionClient)
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
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)
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
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME,
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

	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb}, nil)
	mockSqlDb.On("GetTableInitStatements", mock.Anything, mock.Anything).Return([]*sql_manager.TableInitStatement{
		{
			CreateTableStatement: "test-create-statement",
			AlterTableStatements: []*sql_manager.AlterTableStatement{
				{
					Statement:      "test-pk-statement",
					ConstraintType: sql_manager.PrimaryConstraintType,
				},
				{
					Statement:      "test-fk-statement",
					ConstraintType: sql_manager.ForeignConstraintType,
				},
			},
			IndexStatements: []string{"test-idx-statement"},
		},
	}, nil)

	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-create-statement"}, &sql_manager.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-pk-statement"}, &sql_manager.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-fk-statement"}, &sql_manager.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-idx-statement"}, &sql_manager.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("Close").Return(nil)

	bbuilder := newInitStatementBuilder(mockSqlManager, mockJobClient, mockConnectionClient)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
}

func Test_InitStatementBuilder_Mysql_Generate(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)
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
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME,
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
							Config: &mgmtv1alpha1.JobDestinationOptions_MysqlOptions{
								MysqlOptions: &mgmtv1alpha1.MysqlDestinationConnectionOptions{
									TruncateTable: &mgmtv1alpha1.MysqlTruncateTableConfig{
										TruncateBeforeInsert: false,
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
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb}, nil)
	mockSqlDb.On("Close").Return(nil)

	bbuilder := newInitStatementBuilder(mockSqlManager, mockJobClient, mockConnectionClient)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
}

func Test_InitStatementBuilder_Mysql_TruncateCreate(t *testing.T) {
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlDb := sql_manager.NewMockSqlDatabase(t)
	mockSqlManager := sql_manager.NewMockSqlManagerClient(t)

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
							Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME,
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

	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sql_manager.SqlConnection{Db: mockSqlDb}, nil)
	mockSqlDb.On("GetForeignKeyConstraintsMap", mock.Anything, []string{"public"}).Return(map[string][]*sql_manager.ForeignConstraint{
		"public.users": {{
			Columns:     []string{"account_id"},
			NotNullable: []bool{true},
			ForeignKey:  &sql_manager.ForeignKey{Table: "public.accounts", Columns: []string{"id"}},
		}},
	}, nil)
	accountCreateStmt := "CREATE TABLE IF NOT EXISTS \"public\".\"accounts\" (\"id\" uuid NOT NULL DEFAULT gen_random_uuid(), CONSTRAINT accounts_pkey PRIMARY KEY (id));"
	usersCreateStmt := "CREATE TABLE IF NOT EXISTS \"public\".\"users\" (\"id\" uuid NOT NULL DEFAULT gen_random_uuid(), \"account_id\" uuid NULL, CONSTRAINT users_pkey PRIMARY KEY (id), CONSTRAINT accounts_pkey PRIMARY KEY (id));"
	mockSqlDb.On("GetCreateTableStatement", mock.Anything, "public", "accounts").Return(accountCreateStmt, nil)
	mockSqlDb.On("GetCreateTableStatement", mock.Anything, "public", "users").Return(usersCreateStmt, nil)

	createStmts := []string{accountCreateStmt, usersCreateStmt}
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, mock.MatchedBy(func(query []string) bool { return compareSlices(query, createStmts) }), &sql_manager.BatchExecOpts{}).Return(nil)
	disableFkChecks := sql_manager.DisableForeignKeyChecks
	truncateStmts := []string{"TRUNCATE \"public\".\"users\";", "TRUNCATE \"public\".\"accounts\";"}
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, mock.MatchedBy(func(query []string) bool { return compareSlices(query, truncateStmts) }), &sql_manager.BatchExecOpts{Prefix: &disableFkChecks}).Return(nil)
	mockSqlDb.On("Close").Return(nil)

	bbuilder := newInitStatementBuilder(mockSqlManager, mockJobClient, mockConnectionClient)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123", WorkflowId: "123"},
		slog.Default(),
	)
	assert.Nil(t, err)
}

func Test_getFilteredForeignToPrimaryTableMap(t *testing.T) {
	tables := map[string]struct{}{
		"public.regions":     {},
		"public.jobs":        {},
		"public.countries":   {},
		"public.locations":   {},
		"public.dependents":  {},
		"public.departments": {},
		"public.employees":   {},
	}
	dependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.countries": {
			{Columns: []string{"region_id"}, NotNullable: []bool{true}, ForeignKey: &sql_manager.ForeignKey{Table: "public.regions", Columns: []string{"region_id"}}},
		},
		"public.departments": {
			{Columns: []string{"location_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.locations", Columns: []string{"location_id"}}},
		},
		"public.dependents": {
			{Columns: []string{"dependent_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.employees", Columns: []string{"employees_id"}}},
		},
		"public.locations": {
			{Columns: []string{"country_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.countries", Columns: []string{"country_id"}}},
		},
		"public.employees": {
			{Columns: []string{"department_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.departments", Columns: []string{"department_id"}}},
			{Columns: []string{"job_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.jobs", Columns: []string{"job_id"}}},
			{Columns: []string{"manager_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.employees", Columns: []string{"employee_id"}}},
		},
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
	actual := getFilteredForeignToPrimaryTableMap(dependencies, tables)
	assert.Len(t, actual, len(expected))
	for table, deps := range actual {
		assert.Len(t, deps, len(expected[table]))
		assert.ElementsMatch(t, expected[table], deps)
	}
}

func Test_getFilteredForeignToPrimaryTableMap_filtered(t *testing.T) {
	tables := map[string]struct{}{
		"public.countries": {},
	}
	dependencies := map[string][]*sql_manager.ForeignConstraint{
		"public.countries": {
			{Columns: []string{"region_id"}, NotNullable: []bool{true}, ForeignKey: &sql_manager.ForeignKey{Table: "public.regions", Columns: []string{"region_id"}}}},

		"public.departments": {
			{Columns: []string{"location_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.locations", Columns: []string{"location_id"}}},
		},
		"public.dependents": {
			{Columns: []string{"dependent_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.employees", Columns: []string{"employees_id"}}},
		},
		"public.locations": {
			{Columns: []string{"country_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.countries", Columns: []string{"country_id"}}},
		},
		"public.employees": {
			{Columns: []string{"department_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.departments", Columns: []string{"department_id"}}},
			{Columns: []string{"job_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.jobs", Columns: []string{"job_id"}}},
			{Columns: []string{"manager_id"}, NotNullable: []bool{false}, ForeignKey: &sql_manager.ForeignKey{Table: "public.employees", Columns: []string{"employee_id"}}},
		},
	}

	expected := map[string][]string{
		"public.countries": {},
	}
	actual := getFilteredForeignToPrimaryTableMap(dependencies, tables)
	assert.Len(t, actual, len(expected))
	for table, deps := range actual {
		assert.Len(t, deps, len(expected[table]))
		assert.ElementsMatch(t, expected[table], deps)
	}
}

func compareSlices(slice1, slice2 []string) bool {
	for _, ele := range slice1 {
		if !slices.Contains(slice2, ele) {
			return false
		}
	}
	return true
}
