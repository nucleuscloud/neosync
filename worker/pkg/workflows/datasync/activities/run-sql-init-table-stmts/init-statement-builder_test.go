package runsqlinittablestmts_activity

import (
	"context"
	"slices"
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	workflowId = "workflow-id"
)

func Test_InitStatementBuilder_Pg_Generate_InitSchema(t *testing.T) {
	t.Parallel()
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlDb := sqlmanager.NewMockSqlDatabase(t)
	mockSqlManager := sqlmanager.NewMockSqlManagerClient(t)
	connectionId := "456"
	fkconnectionId := "789"

	mockJobClient.On("SetRunContext", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.SetRunContextResponse{}), nil)
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
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
										IncludeHyphens: gotypeutil.ToPtr(true),
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
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sqlmanager.SqlConnection{Db: mockSqlDb}, nil)
	mockSqlDb.On("GetSequencesByTables", mock.Anything, mock.Anything, mock.Anything).Return([]*sqlmanager_shared.DataType{}, nil)
	mockSqlDb.On("GetSchemaInitStatements", mock.Anything, []*sqlmanager_shared.SchemaTable{{Schema: "public", Table: "users"}}).Return([]*sqlmanager_shared.InitSchemaStatements{
		{Label: "data types", Statements: []string{}},
		{Label: "create table", Statements: []string{"test-create-statement"}},
		{Label: "non-fk alter table", Statements: []string{"test-pk-statement"}},
		{Label: "fk alter table", Statements: []string{"test-fk-statement"}},
		{Label: "table index", Statements: []string{"test-idx-statement"}},
		{Label: "table triggers", Statements: []string{"test-trigger-statement"}},
	}, nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"TRUNCATE \"public\".\"users\" RESTART IDENTITY CASCADE;"}, &sqlmanager_shared.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-trigger-statement"}, &sqlmanager_shared.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-create-statement"}, &sqlmanager_shared.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-pk-statement"}, &sqlmanager_shared.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-fk-statement"}, &sqlmanager_shared.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-idx-statement"}, &sqlmanager_shared.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("Close").Return(nil)

	bbuilder := newInitStatementBuilder(mockSqlManager, mockJobClient, mockConnectionClient, nil, true, workflowId)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123"},
		testutil.GetTestLogger(t),
	)
	assert.Nil(t, err)
}

func Test_InitStatementBuilder_Pg_Generate_NoInitStatement(t *testing.T) {
	t.Parallel()
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlDb := sqlmanager.NewMockSqlDatabase(t)
	mockSqlManager := sqlmanager.NewMockSqlManagerClient(t)
	connectionId := "456"

	mockJobClient.On("SetRunContext", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.SetRunContextResponse{}), nil)
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
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
										IncludeHyphens: gotypeutil.ToPtr(true),
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
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sqlmanager.SqlConnection{Db: mockSqlDb}, nil)
	mockSqlDb.On("Close").Return(nil)

	bbuilder := newInitStatementBuilder(mockSqlManager, mockJobClient, mockConnectionClient, nil, true, workflowId)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123"},
		testutil.GetTestLogger(t),
	)
	assert.Nil(t, err)
}

func Test_InitStatementBuilder_Pg_TruncateCascade(t *testing.T) {
	t.Parallel()
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlDb := sqlmanager.NewMockSqlDatabase(t)
	mockSqlManager := sqlmanager.NewMockSqlManagerClient(t)
	connectionId := "456"

	mockJobClient.On("SetRunContext", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.SetRunContextResponse{}), nil)
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
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
										IncludeHyphens: gotypeutil.ToPtr(true),
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
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
										IncludeHyphens: gotypeutil.ToPtr(true),
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
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sqlmanager.SqlConnection{Db: mockSqlDb}, nil)
	mockSqlDb.On("GetSequencesByTables", mock.Anything, mock.Anything, mock.Anything).Return([]*sqlmanager_shared.DataType{}, nil)
	stmts := []string{"TRUNCATE \"public\".\"users\" RESTART IDENTITY CASCADE;", "TRUNCATE \"public\".\"accounts\" RESTART IDENTITY CASCADE;"}
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, mock.MatchedBy(func(query []string) bool { return compareSlices(query, stmts) }), &sqlmanager_shared.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("Close").Return(nil)

	bbuilder := newInitStatementBuilder(mockSqlManager, mockJobClient, mockConnectionClient, nil, true, workflowId)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123"},
		testutil.GetTestLogger(t),
	)
	assert.Nil(t, err)
}

func Test_InitStatementBuilder_Pg_Truncate(t *testing.T) {
	t.Parallel()
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlDb := sqlmanager.NewMockSqlDatabase(t)
	mockSqlManager := sqlmanager.NewMockSqlManagerClient(t)
	connectionId := "456"

	mockJobClient.On("SetRunContext", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.SetRunContextResponse{}), nil)
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
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
										IncludeHyphens: gotypeutil.ToPtr(true),
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
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
										IncludeHyphens: gotypeutil.ToPtr(true),
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
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sqlmanager.SqlConnection{Db: mockSqlDb}, nil)
	mockSqlDb.On("GetTableConstraintsBySchema", mock.Anything, []string{"public"}).Return(&sqlmanager_shared.TableConstraints{
		ForeignKeyConstraints: map[string][]*sqlmanager_shared.ForeignConstraint{
			"public.users": {{
				Columns:     []string{"account_id"},
				NotNullable: []bool{true},
				ForeignKey:  &sqlmanager_shared.ForeignKey{Table: "public.accounts", Columns: []string{"id"}},
			}},
		},
	}, nil)
	mockSqlDb.On("GetSequencesByTables", mock.Anything, mock.Anything, mock.Anything).Return([]*sqlmanager_shared.DataType{}, nil)
	mockSqlDb.On("Exec", mock.Anything, "TRUNCATE \"public\".\"accounts\", \"public\".\"users\" RESTART IDENTITY;").Return(nil)
	mockSqlDb.On("Close").Return(nil)

	bbuilder := newInitStatementBuilder(mockSqlManager, mockJobClient, mockConnectionClient, nil, true, workflowId)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123"},
		testutil.GetTestLogger(t),
	)
	assert.Nil(t, err)
}

func Test_InitStatementBuilder_Pg_InitSchema(t *testing.T) {
	t.Parallel()
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlDb := sqlmanager.NewMockSqlDatabase(t)
	mockSqlManager := sqlmanager.NewMockSqlManagerClient(t)
	connectionId := "456"

	mockJobClient.On("SetRunContext", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.SetRunContextResponse{}), nil)
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
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
										IncludeHyphens: gotypeutil.ToPtr(true),
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
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
										IncludeHyphens: gotypeutil.ToPtr(true),
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

	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sqlmanager.SqlConnection{Db: mockSqlDb}, nil)
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sqlmanager.SqlConnection{Db: mockSqlDb}, nil)
	mockSqlDb.On("GetSchemaInitStatements", mock.Anything, mock.Anything).Return([]*sqlmanager_shared.InitSchemaStatements{
		{Label: "data types", Statements: []string{}},
		{Label: "create table", Statements: []string{"test-create-statement"}},
		{Label: "non-fk alter table", Statements: []string{"test-pk-statement"}},
		{Label: "fk alter table", Statements: []string{"test-fk-statement"}},
		{Label: "table index", Statements: []string{"test-idx-statement"}},
		{Label: "table triggers", Statements: []string{"test-trigger-statement"}},
	}, nil)

	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-create-statement"}, &sqlmanager_shared.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-trigger-statement"}, &sqlmanager_shared.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-pk-statement"}, &sqlmanager_shared.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-fk-statement"}, &sqlmanager_shared.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("BatchExec", mock.Anything, mock.Anything, []string{"test-idx-statement"}, &sqlmanager_shared.BatchExecOpts{}).Return(nil)
	mockSqlDb.On("Close").Return(nil)

	bbuilder := newInitStatementBuilder(mockSqlManager, mockJobClient, mockConnectionClient, nil, true, workflowId)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123"},
		testutil.GetTestLogger(t),
	)
	assert.Nil(t, err)
}

func Test_InitStatementBuilder_Mysql_Generate(t *testing.T) {
	t.Parallel()
	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)
	mockSqlDb := sqlmanager.NewMockSqlDatabase(t)
	mockSqlManager := sqlmanager.NewMockSqlManagerClient(t)
	connectionId := "456"

	mockJobClient.On("SetRunContext", mock.Anything, mock.Anything).Return(connect.NewResponse(&mgmtv1alpha1.SetRunContextResponse{}), nil)
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
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
										IncludeHyphens: gotypeutil.ToPtr(true),
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
	mockSqlManager.On("NewPooledSqlDb", mock.Anything, mock.Anything, mock.Anything).Return(&sqlmanager.SqlConnection{Db: mockSqlDb}, nil)
	mockSqlDb.On("Close").Return(nil)

	bbuilder := newInitStatementBuilder(mockSqlManager, mockJobClient, mockConnectionClient, nil, true, workflowId)
	_, err := bbuilder.RunSqlInitTableStatements(
		context.Background(),
		&RunSqlInitTableStatementsRequest{JobId: "123"},
		testutil.GetTestLogger(t),
	)
	assert.Nil(t, err)
}

func Test_getFilteredForeignToPrimaryTableMap(t *testing.T) {
	t.Parallel()
	tables := map[string]struct{}{
		"public.regions":     {},
		"public.jobs":        {},
		"public.countries":   {},
		"public.locations":   {},
		"public.dependents":  {},
		"public.departments": {},
		"public.employees":   {},
	}
	dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.countries": {
			{Columns: []string{"region_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.regions", Columns: []string{"region_id"}}},
		},
		"public.departments": {
			{Columns: []string{"location_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.locations", Columns: []string{"location_id"}}},
		},
		"public.dependents": {
			{Columns: []string{"dependent_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.employees", Columns: []string{"employees_id"}}},
		},
		"public.locations": {
			{Columns: []string{"country_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.countries", Columns: []string{"country_id"}}},
		},
		"public.employees": {
			{Columns: []string{"department_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.departments", Columns: []string{"department_id"}}},
			{Columns: []string{"job_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.jobs", Columns: []string{"job_id"}}},
			{Columns: []string{"manager_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.employees", Columns: []string{"employee_id"}}},
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
	t.Parallel()
	tables := map[string]struct{}{
		"public.countries": {},
	}
	dependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"public.countries": {
			{Columns: []string{"region_id"}, NotNullable: []bool{true}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.regions", Columns: []string{"region_id"}}}},

		"public.departments": {
			{Columns: []string{"location_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.locations", Columns: []string{"location_id"}}},
		},
		"public.dependents": {
			{Columns: []string{"dependent_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.employees", Columns: []string{"employees_id"}}},
		},
		"public.locations": {
			{Columns: []string{"country_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.countries", Columns: []string{"country_id"}}},
		},
		"public.employees": {
			{Columns: []string{"department_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.departments", Columns: []string{"department_id"}}},
			{Columns: []string{"job_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.jobs", Columns: []string{"job_id"}}},
			{Columns: []string{"manager_id"}, NotNullable: []bool{false}, ForeignKey: &sqlmanager_shared.ForeignKey{Table: "public.employees", Columns: []string{"employee_id"}}},
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
