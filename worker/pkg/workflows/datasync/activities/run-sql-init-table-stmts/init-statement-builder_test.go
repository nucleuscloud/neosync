package runsqlinittablestmts_activity

import (
	"context"
	"log/slog"
	"testing"

	"connectrpc.com/connect"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// todo figure out how to mock pgxpool
// func Test_InitStatementBuilder_Pg_Generate(t *testing.T) {
// 	mockJobClient := mgmtv1alpha1connect.NewMockJobServiceClient(t)
// 	mockConnectionClient := mgmtv1alpha1connect.NewMockConnectionServiceClient(t)

// 	dbtx := pg_queries.NewMockDBTX(t)
// 	pgcache := map[string]pg_queries.DBTX{
// 		"fake-prod-url": pg_queries.NewMockDBTX(t),
// 		"postgresql://postgres:foofar@localhost:5435/nucleus": dbtx,
// 	}
// 	pgquerier := pg_queries.NewMockQuerier(t)
// 	mysqlcache := map[string]mysql_queries.DBTX{}
// 	mysqlquerier := mysql_queries.NewMockQuerier(t)
// 	connectionId := "456"

// 	mockJobClient.On("GetJob", mock.Anything, mock.Anything).
// 		Return(connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
// 			Job: &mgmtv1alpha1.Job{
// 				Source: &mgmtv1alpha1.JobSource{
// 					Options: &mgmtv1alpha1.JobSourceOptions{
// 						Config: &mgmtv1alpha1.JobSourceOptions_Generate{
// 							Generate: &mgmtv1alpha1.GenerateSourceOptions{
// 								Schemas: []*mgmtv1alpha1.GenerateSourceSchemaOption{
// 									{
// 										Schema: "public",
// 										Tables: []*mgmtv1alpha1.GenerateSourceTableOption{
// 											{
// 												Table:    "users",
// 												RowCount: 10,
// 											},
// 										},
// 									},
// 								},
// 								FkSourceConnectionId: &connectionId,
// 							},
// 						},
// 					},
// 				},
// 				Mappings: []*mgmtv1alpha1.JobMapping{
// 					{
// 						Schema: "public",
// 						Table:  "users",
// 						Column: "id",
// 						Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 							Source: "generate_uuid",
// 							Config: &mgmtv1alpha1.TransformerConfig{
// 								Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
// 									GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
// 										IncludeHyphens: true,
// 									},
// 								},
// 							},
// 						},
// 					},
// 					{
// 						Schema: "public",
// 						Table:  "users",
// 						Column: "name",
// 						Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 							Source: "generate_full_name",
// 							Config: &mgmtv1alpha1.TransformerConfig{
// 								Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
// 									GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
// 								},
// 							},
// 						},
// 					},
// 				},
// 				Destinations: []*mgmtv1alpha1.JobDestination{
// 					{
// 						ConnectionId: "456",
// 						Options: &mgmtv1alpha1.JobDestinationOptions{
// 							Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
// 								PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
// 									TruncateTable: &mgmtv1alpha1.PostgresTruncateTableConfig{
// 										TruncateBeforeInsert: true,
// 										Cascade:              true,
// 									},
// 									InitTableSchema: true,
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		}), nil)

// 	mockConnectionClient.On(
// 		"GetConnection",
// 		mock.Anything,
// 		connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
// 			Id: "456",
// 		}),
// 	).Return(connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
// 		Connection: &mgmtv1alpha1.Connection{
// 			Id:   "456",
// 			Name: "stage",
// 			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
// 				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
// 					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
// 						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
// 							Url: "postgresql://postgres:foofar@localhost:5435/nucleus",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}), nil)
// 	var cmdtag pgconn.CommandTag
// 	dbtx.On("Exec", mock.Anything, mock.Anything).Return(cmdtag, nil)

// 	bbuilder := newInitStatementBuilder(pgcache, pgquerier, mysqlcache, mysqlquerier, mockJobClient, mockConnectionClient)
// 	_, err := bbuilder.RunSqlInitTableStatements(
// 		context.Background(),
// 		&RunSqlInitTableStatementsRequest{JobId: "123", WorkflowId: "123"},
// 		log.NewStructuredLogger(slog.Default()),
// 	)
// 	assert.Nil(t, err)
// }

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
