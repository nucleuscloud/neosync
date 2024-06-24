package datasync_workflow

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	mockTemporalClient "github.com/nucleuscloud/neosync/worker/internal/mocks/go.temporal.io/sdk/client"
	genbenthosconfigs_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/gen-benthos-configs"
	runsqlinittablestmts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/run-sql-init-table-stmts"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	sync_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync"
	syncactivityopts_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-activity-opts"
	syncrediscleanup_activity "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/sync-redis-clean-up"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/sdk/testsuite"
)

// func (s *IntegrationTestSuite) Test_Workflow_VirtualForeignKeys_Passthrough() {
// 	s.SetupTestByFolder("virtual-foreign-keys")
// 	mux := http.NewServeMux()
// 	mux.Handle(mgmtv1alpha1connect.JobServiceGetJobProcedure, connect.NewUnaryHandler(
// 		mgmtv1alpha1connect.JobServiceGetJobProcedure,
// 		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetJobRequest]) (*connect.Response[mgmtv1alpha1.GetJobResponse], error) {
// 			return connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
// 				Job: &mgmtv1alpha1.Job{
// 					Id: "115aaf2c-776e-4847-8268-d914e3c15968",
// 					Source: &mgmtv1alpha1.JobSource{
// 						Options: &mgmtv1alpha1.JobSourceOptions{
// 							Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
// 								Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
// 									ConnectionId: "c9b6ce58-5c8e-4dce-870d-96841b19d988",
// 								},
// 							},
// 						},
// 					},
// 					Mappings: []*mgmtv1alpha1.JobMapping{
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "regions",
// 							Column: "region_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "regions",
// 							Column: "region_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "regions",
// 							Column: "region_name",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "countries",
// 							Column: "country_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "countries",
// 							Column: "country_name",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "countries",
// 							Column: "region_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "locations",
// 							Column: "location_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "locations",
// 							Column: "street_address",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "locations",
// 							Column: "postal_code",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "locations",
// 							Column: "city",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "locations",
// 							Column: "state_province",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "locations",
// 							Column: "country_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "departments",
// 							Column: "department_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "departments",
// 							Column: "department_name",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "departments",
// 							Column: "location_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "jobs",
// 							Column: "job_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "jobs",
// 							Column: "job_title",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "jobs",
// 							Column: "min_salary",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "jobs",
// 							Column: "max_salary",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "dependents",
// 							Column: "dependent_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "dependents",
// 							Column: "first_name",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "dependents",
// 							Column: "last_name",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "dependents",
// 							Column: "relationship",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "dependents",
// 							Column: "employee_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "employees",
// 							Column: "employee_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "employees",
// 							Column: "first_name",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "employees",
// 							Column: "last_name",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "employees",
// 							Column: "email",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "employees",
// 							Column: "phone_number",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "employees",
// 							Column: "hire_date",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "employees",
// 							Column: "job_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "employees",
// 							Column: "salary",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "employees",
// 							Column: "manager_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 						{
// 							Schema: "vfk_hr",
// 							Table:  "employees",
// 							Column: "department_id",
// 							Transformer: &mgmtv1alpha1.JobMappingTransformer{
// 								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
// 							},
// 						},
// 					},
// 					VirtualForeignKeys: []*mgmtv1alpha1.VirtualForeignConstraint{
// 						{
// 							Schema:  "vfk_hr",
// 							Table:   "countries",
// 							Columns: []string{"region_id"},
// 							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
// 								Schema:  "vfk_hr",
// 								Table:   "regions",
// 								Columns: []string{"region_id"},
// 							},
// 						},
// 						{
// 							Schema:  "vfk_hr",
// 							Table:   "departments",
// 							Columns: []string{"location_id"},
// 							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
// 								Schema:  "vfk_hr",
// 								Table:   "locations",
// 								Columns: []string{"location_id"},
// 							},
// 						},
// 						{
// 							Schema:  "vfk_hr",
// 							Table:   "dependents",
// 							Columns: []string{"employee_id"},
// 							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
// 								Schema:  "vfk_hr",
// 								Table:   "employees",
// 								Columns: []string{"employee_id"},
// 							},
// 						},
// 						{
// 							Schema:  "vfk_hr",
// 							Table:   "employees",
// 							Columns: []string{"manager_id"},
// 							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
// 								Schema:  "vfk_hr",
// 								Table:   "employees",
// 								Columns: []string{"employee_id"},
// 							},
// 						},
// 						{
// 							Schema:  "vfk_hr",
// 							Table:   "employees",
// 							Columns: []string{"department_id"},
// 							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
// 								Schema:  "vfk_hr",
// 								Table:   "departments",
// 								Columns: []string{"department_id"},
// 							},
// 						},
// 						{
// 							Schema:  "vfk_hr",
// 							Table:   "employees",
// 							Columns: []string{"job_id"},
// 							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
// 								Schema:  "vfk_hr",
// 								Table:   "jobs",
// 								Columns: []string{"job_id"},
// 							},
// 						},
// 						{
// 							Schema:  "vfk_hr",
// 							Table:   "locations",
// 							Columns: []string{"country_id"},
// 							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
// 								Schema:  "vfk_hr",
// 								Table:   "countries",
// 								Columns: []string{"country_id"},
// 							},
// 						},
// 					},
// 					Destinations: []*mgmtv1alpha1.JobDestination{
// 						{
// 							ConnectionId: "226add85-5751-4232-b085-a0ae93afc7ce",
// 						},
// 					},
// 				},
// 			}), nil
// 		},
// 	))

// 	mux.Handle(mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure, connect.NewUnaryHandler(
// 		mgmtv1alpha1connect.ConnectionServiceGetConnectionProcedure,
// 		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetConnectionRequest]) (*connect.Response[mgmtv1alpha1.GetConnectionResponse], error) {
// 			if r.Msg.GetId() == "c9b6ce58-5c8e-4dce-870d-96841b19d988" {
// 				return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
// 					Connection: &mgmtv1alpha1.Connection{
// 						Id:   "c9b6ce58-5c8e-4dce-870d-96841b19d988",
// 						Name: "source",
// 						ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
// 							Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
// 								PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
// 									ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
// 										Url: s.sourceDsn,
// 									},
// 								},
// 							},
// 						},
// 					},
// 				}), nil
// 			}
// 			if r.Msg.GetId() == "226add85-5751-4232-b085-a0ae93afc7ce" {
// 				return connect.NewResponse(&mgmtv1alpha1.GetConnectionResponse{
// 					Connection: &mgmtv1alpha1.Connection{
// 						Id:   "226add85-5751-4232-b085-a0ae93afc7ce",
// 						Name: "target",
// 						ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
// 							Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
// 								PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
// 									ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
// 										Url: s.targetDsn,
// 									},
// 								},
// 							},
// 						},
// 					},
// 				}), nil
// 			}
// 			return nil, nil
// 		},
// 	))
// 	srv := startHTTPServer(s.T(), mux)
// 	executeWorkflow(s.T(), srv, s.redisUrl, "115aaf2c-776e-4847-8268-d914e3c15968")

// 	tables := []string{"regions", "countries", "locations", "departments", "dependents", "locations", "jobs", "employees"}
// 	for _, t := range tables {
// 		rows, err := s.targetPgPool.Query(s.ctx, fmt.Sprintf("select * from vfk_hr.%s;", t))
// 		require.NoError(s.T(), err)
// 		for rows.Next() {
// 			values, err := rows.Values()
// 			count := 0
// 			for i := range values {
// 				count = i
// 			}
// 			require.Greater(s.T(), count, 0)
// 			require.NoError(s.T(), err)
// 		}
// 	}

// 	s.TearDownTestByFolder("virtual-foreign-keys")
// }

func (s *IntegrationTestSuite) Test_Workflow_VirtualForeignKeys_Transform() {
	s.SetupTestByFolder("virtual-foreign-keys")
	// neosync api mocks
	mux := http.NewServeMux()
	mux.Handle(mgmtv1alpha1connect.JobServiceGetJobProcedure, connect.NewUnaryHandler(
		mgmtv1alpha1connect.JobServiceGetJobProcedure,
		func(ctx context.Context, r *connect.Request[mgmtv1alpha1.GetJobRequest]) (*connect.Response[mgmtv1alpha1.GetJobResponse], error) {
			return connect.NewResponse(&mgmtv1alpha1.GetJobResponse{
				Job: &mgmtv1alpha1.Job{
					Id: "fd4d8660-31a0-48b2-9adf-10f11b94898f",
					Source: &mgmtv1alpha1.JobSource{
						Options: &mgmtv1alpha1.JobSourceOptions{
							Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
								Postgres: &mgmtv1alpha1.PostgresSourceConnectionOptions{
									ConnectionId: "c9b6ce58-5c8e-4dce-870d-96841b19d988",
								},
							},
						},
					},
					Mappings: []*mgmtv1alpha1.JobMapping{
						{
							Schema: "vfk_hr",
							Table:  "regions",
							Column: "region_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "regions",
							Column: "region_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "regions",
							Column: "region_name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "countries",
							Column: "country_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
								Config: &mgmtv1alpha1.TransformerConfig{
									Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
										TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: `if (value == 'US') { return 'SU'; } return value;`},
									},
								},
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "countries",
							Column: "country_name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "countries",
							Column: "region_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "locations",
							Column: "location_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "locations",
							Column: "street_address",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "locations",
							Column: "postal_code",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "locations",
							Column: "city",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "locations",
							Column: "state_province",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "locations",
							Column: "country_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "departments",
							Column: "department_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "departments",
							Column: "department_name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "departments",
							Column: "location_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "jobs",
							Column: "job_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "jobs",
							Column: "job_title",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "jobs",
							Column: "min_salary",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "jobs",
							Column: "max_salary",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "dependents",
							Column: "dependent_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "dependents",
							Column: "first_name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "dependents",
							Column: "last_name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "dependents",
							Column: "relationship",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "dependents",
							Column: "employee_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "employees",
							Column: "employee_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "employees",
							Column: "first_name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "employees",
							Column: "last_name",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "employees",
							Column: "email",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "employees",
							Column: "phone_number",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "employees",
							Column: "hire_date",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "employees",
							Column: "job_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "employees",
							Column: "salary",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "employees",
							Column: "manager_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
						{
							Schema: "vfk_hr",
							Table:  "employees",
							Column: "department_id",
							Transformer: &mgmtv1alpha1.JobMappingTransformer{
								Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
							},
						},
					},
					VirtualForeignKeys: []*mgmtv1alpha1.VirtualForeignConstraint{
						{
							Schema:  "vfk_hr",
							Table:   "countries",
							Columns: []string{"region_id"},
							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
								Schema:  "vfk_hr",
								Table:   "regions",
								Columns: []string{"region_id"},
							},
						},
						{
							Schema:  "vfk_hr",
							Table:   "departments",
							Columns: []string{"location_id"},
							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
								Schema:  "vfk_hr",
								Table:   "locations",
								Columns: []string{"location_id"},
							},
						},
						{
							Schema:  "vfk_hr",
							Table:   "dependents",
							Columns: []string{"employee_id"},
							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
								Schema:  "vfk_hr",
								Table:   "employees",
								Columns: []string{"employee_id"},
							},
						},
						{
							Schema:  "vfk_hr",
							Table:   "employees",
							Columns: []string{"manager_id"},
							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
								Schema:  "vfk_hr",
								Table:   "employees",
								Columns: []string{"employee_id"},
							},
						},
						{
							Schema:  "vfk_hr",
							Table:   "employees",
							Columns: []string{"department_id"},
							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
								Schema:  "vfk_hr",
								Table:   "departments",
								Columns: []string{"department_id"},
							},
						},
						{
							Schema:  "vfk_hr",
							Table:   "employees",
							Columns: []string{"job_id"},
							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
								Schema:  "vfk_hr",
								Table:   "jobs",
								Columns: []string{"job_id"},
							},
						},
						{
							Schema:  "vfk_hr",
							Table:   "locations",
							Columns: []string{"country_id"},
							ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
								Schema:  "vfk_hr",
								Table:   "countries",
								Columns: []string{"country_id"},
							},
						},
					},
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
										Url: s.sourceDsn,
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
										Url: s.targetDsn,
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
	srv := startHTTPServer(s.T(), mux)
	executeWorkflow(s.T(), srv, s.redisUrl, "fd4d8660-31a0-48b2-9adf-10f11b94898f")

	tables := []string{"regions", "countries", "locations", "departments", "dependents", "locations", "jobs", "employees"}
	for _, t := range tables {
		rows, err := s.targetPgPool.Query(s.ctx, fmt.Sprintf("select * from vfk_hr.%s;", t))
		require.NoError(s.T(), err)
		for rows.Next() {
			values, err := rows.Values()
			count := 0
			for i := range values {
				count = i
			}
			require.Greater(s.T(), count, 0)
			require.NoError(s.T(), err)
		}
	}

	s.TearDownTestByFolder("virtual-foreign-keys")
}

func executeWorkflow(
	t *testing.T,
	srv *httptest.Server,
	redisUrl string,
	jobId string,
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
	temporalClientMock := mockTemporalClient.NewMockClient(t)
	pgpoolmap := &sync.Map{}
	mysqlpoolmap := &sync.Map{}
	pgquerier := pg_queries.New()
	mysqlquerier := mysql_queries.New()
	sqlmanager := sql_manager.NewSqlManager(pgpoolmap, pgquerier, mysqlpoolmap, mysqlquerier, sqlconnector)

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
	syncActivity := sync_activity.New(connclient, &sync.Map{}, temporalClientMock, activityMeter, sync_activity.NewBenthosStreamManager(), disableReaper)
	retrieveActivityOpts := syncactivityopts_activity.New(jobclient)
	runSqlInitTableStatements := runsqlinittablestmts_activity.New(jobclient, connclient, sqlmanager)
	env.RegisterWorkflow(Workflow)
	env.RegisterActivity(syncActivity.Sync)
	env.RegisterActivity(retrieveActivityOpts.RetrieveActivityOptions)
	env.RegisterActivity(runSqlInitTableStatements.RunSqlInitTableStatements)
	env.RegisterActivity(syncrediscleanup_activity.DeleteRedisHash)
	env.RegisterActivity(genbenthosActivity.GenerateBenthosConfigs)
	env.SetTestTimeout(120 * time.Second) // increase the test timeout

	env.ExecuteWorkflow(Workflow, &WorkflowRequest{JobId: jobId})
	require.True(t, env.IsWorkflowCompleted())

	err := env.GetWorkflowError()
	require.Nil(t, err)
}

func startHTTPServer(tb testing.TB, h http.Handler) *httptest.Server {
	tb.Helper()
	srv := httptest.NewUnstartedServer(h)
	srv.EnableHTTP2 = true
	srv.Start()
	tb.Cleanup(srv.Close)
	return srv
}
