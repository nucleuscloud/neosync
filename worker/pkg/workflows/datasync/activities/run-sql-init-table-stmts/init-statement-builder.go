package runsqlinittablestmts_activity

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_mssql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mssql"
	sqlmanager_mysql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mysql"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	connectionmanager "github.com/nucleuscloud/neosync/internal/connection-manager"
	"github.com/nucleuscloud/neosync/internal/ee/license"
	ee_sqlmanager_mssql "github.com/nucleuscloud/neosync/internal/ee/mssql-manager"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

const (
	batchSizeConst = 20
)

type initStatementBuilder struct {
	sqlmanager sql_manager.SqlManagerClient
	jobclient  mgmtv1alpha1connect.JobServiceClient
	connclient mgmtv1alpha1connect.ConnectionServiceClient
	eelicense  license.EEInterface
	workflowId string
}

func newInitStatementBuilder(
	sqlmanagerclient sql_manager.SqlManagerClient,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	eelicense license.EEInterface,
	workflowId string,
) *initStatementBuilder {
	return &initStatementBuilder{
		sqlmanager: sqlmanagerclient,
		jobclient:  jobclient,
		connclient: connclient,
		eelicense:  eelicense,
		workflowId: workflowId,
	}
}

func (b *initStatementBuilder) RunSqlInitTableStatements(
	ctx context.Context,
	req *RunSqlInitTableStatementsRequest,
	session connectionmanager.SessionInterface,
	slogger *slog.Logger,
) (*RunSqlInitTableStatementsResponse, error) {
	job, err := b.getJobById(ctx, req.JobId)
	if err != nil {
		return nil, fmt.Errorf("unable to get job by id: %w", err)
	}

	sourceConnection, err := shared.GetJobSourceConnection(ctx, job.GetSource(), b.connclient)
	if err != nil {
		return nil, fmt.Errorf("unable to get connection by id: %w", err)
	}

	sourceConnectionType := shared.GetConnectionType(sourceConnection)
	slogger = slogger.With(
		"sourceConnectionType", sourceConnectionType,
	)

	if job.GetSource().GetOptions().GetAiGenerate() != nil {
		sourceConnection, err = shared.GetConnectionById(ctx, b.connclient, *job.GetSource().GetOptions().GetAiGenerate().FkSourceConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get connection by id: %w", err)
		}
	}

	if sourceConnection.GetConnectionConfig().GetMongoConfig() != nil || sourceConnection.GetConnectionConfig().GetDynamodbConfig() != nil {
		return &RunSqlInitTableStatementsResponse{}, nil
	}

	sourcedb, err := b.sqlmanager.NewSqlConnection(ctx, session, sourceConnection, slogger)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}
	defer sourcedb.Db().Close()

	uniqueTables := shared.GetUniqueTablesMapFromJob(job)
	uniqueSchemas := shared.GetUniqueSchemasFromJob(job)

	initSchemaRunContext := []*InitSchemaRunContext{}

	destConns := []*sqlmanager.SqlConnection{}
	defer func() {
		for _, destconn := range destConns {
			destconn.Db().Close()
		}
	}()

	for _, destination := range job.Destinations {
		destinationConnection, err := shared.GetConnectionById(ctx, b.connclient, destination.ConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get destination connection by id (%s): %w", destination.ConnectionId, err)
		}
		destinationConnectionType := shared.GetConnectionType(destinationConnection)
		slogger = slogger.With(
			"destinationConnectionType", destinationConnectionType,
		)
		if destinationConnection.GetConnectionConfig().GetAwsS3Config() != nil || destinationConnection.GetConnectionConfig().GetGcpCloudstorageConfig() != nil {
			// nothing to do for Bucket destinations
			continue
		}
		sqlopts, err := shared.GetSqlJobDestinationOpts(destination.GetOptions())
		if err != nil {
			return nil, err
		}

		if job.GetSource().GetOptions().GetAiGenerate() != nil {
			fkSrcConnId := job.GetSource().GetOptions().GetAiGenerate().GetFkSourceConnectionId()
			if fkSrcConnId == destination.GetConnectionId() && sqlopts.InitSchema {
				slogger.Warn("cannot init schema when destination connection is the same as the foreign key source connection")
				sqlopts.InitSchema = false
			}
		}

		if job.GetSource().GetOptions().GetGenerate() != nil {
			fkSrcConnId := job.GetSource().GetOptions().GetGenerate().GetFkSourceConnectionId()
			if fkSrcConnId == destination.GetConnectionId() && sqlopts.InitSchema {
				slogger.Warn("cannot init schema when destination connection is the same as the foreign key source connection")
				sqlopts.InitSchema = false
			}
		}

		if !sqlopts.TruncateCascade && !sqlopts.TruncateBeforeInsert && !sqlopts.InitSchema {
			slogger.Info("skipping truncate and schema init as none were set to true")
			continue
		}

		destdb, err := b.sqlmanager.NewSqlConnection(ctx, session, destinationConnection, slogger)
		if err != nil {
			return nil, fmt.Errorf("unable to create new sql db: %w", err)
		}
		destConns = append(destConns, destdb)

		switch destinationConnection.ConnectionConfig.Config.(type) {
		case *mgmtv1alpha1.ConnectionConfig_PgConfig:
			destPgConfig := destinationConnection.ConnectionConfig.GetPgConfig()
			if destPgConfig == nil {
				continue
			}

			if sqlopts.InitSchema {
				tables := []*sqlmanager_shared.SchemaTable{}
				for tableKey := range uniqueTables {
					schema, table := sqlmanager_shared.SplitTableKey(tableKey)
					tables = append(tables, &sqlmanager_shared.SchemaTable{Schema: schema, Table: table})
				}

				initblocks, err := sourcedb.Db().GetSchemaInitStatements(ctx, tables)
				if err != nil {
					return nil, err
				}

				for _, block := range initblocks {
					slogger.Info(fmt.Sprintf("[%s] found %d statements to execute during schema initialization", block.Label, len(block.Statements)))
					if len(block.Statements) == 0 {
						continue
					}
					err = destdb.Db().BatchExec(ctx, batchSizeConst, block.Statements, &sqlmanager_shared.BatchExecOpts{})
					if err != nil {
						slogger.Error(fmt.Sprintf("unable to exec pg %s statements: %s", block.Label, err.Error()))
						if block.Label != sqlmanager_postgres.SchemasLabel {
							return nil, fmt.Errorf("unable to exec pg %s statements: %w", block.Label, err)
						}
					}
				}
			}
			// truncate statements
			if sqlopts.TruncateCascade {
				tableTruncateStmts := []string{}
				for table := range uniqueTables {
					schema, table := sqlmanager_shared.SplitTableKey(table)
					stmt, err := sqlmanager_postgres.BuildPgTruncateCascadeStatement(schema, table)
					if err != nil {
						return nil, err
					}
					tableTruncateStmts = append(tableTruncateStmts, stmt)
				}
				slogger.Info(fmt.Sprintf("executing %d sql statements that will truncate cascade tables", len(tableTruncateStmts)))
				err = destdb.Db().BatchExec(ctx, batchSizeConst, tableTruncateStmts, &sqlmanager_shared.BatchExecOpts{})
				if err != nil {
					return nil, fmt.Errorf("unable to exec truncate cascade statements: %w", err)
				}
			} else if sqlopts.TruncateBeforeInsert {
				tableDependencies, err := sourcedb.Db().GetTableConstraintsBySchema(ctx, uniqueSchemas)
				if err != nil {
					return nil, fmt.Errorf("unable to retrieve database foreign key constraints: %w", err)
				}
				slogger.Info(fmt.Sprintf("found %d foreign key constraints for database", len(tableDependencies.ForeignKeyConstraints)))
				tablePrimaryDependencyMap := getFilteredForeignToPrimaryTableMap(tableDependencies.ForeignKeyConstraints, uniqueTables)
				orderedTablesResp, err := tabledependency.GetTablesOrderedByDependency(tablePrimaryDependencyMap)
				if err != nil {
					return nil, err
				}

				slogger.Info(fmt.Sprintf("executing %d sql statements that will truncate tables", len(orderedTablesResp.OrderedTables)))
				truncateStmt, err := sqlmanager_postgres.BuildPgTruncateStatement(orderedTablesResp.OrderedTables)
				if err != nil {
					return nil, fmt.Errorf("unable to build postgres truncate statement: %w", err)
				}
				err = destdb.Db().Exec(ctx, truncateStmt)
				if err != nil {
					return nil, fmt.Errorf("unable to exec ordered truncate statements: %w", err)
				}
			}
			if sqlopts.TruncateBeforeInsert || sqlopts.TruncateCascade {
				// reset serial counts
				// identity counts are automatically reset with truncate identity restart clause
				schemaTableMap := map[string][]string{}
				for schemaTable := range uniqueTables {
					schema, table := sqlmanager_shared.SplitTableKey(schemaTable)
					schemaTableMap[schema] = append(schemaTableMap[schema], table)
				}

				resetSeqStmts := []string{}
				for schema, tables := range schemaTableMap {
					sequences, err := sourcedb.Db().GetSequencesByTables(ctx, schema, tables)
					if err != nil {
						return nil, err
					}
					for _, seq := range sequences {
						resetSeqStmts = append(resetSeqStmts, sqlmanager_postgres.BuildPgResetSequenceSql(seq.Name))
					}
				}
				if len(resetSeqStmts) > 0 {
					err = destdb.Db().BatchExec(ctx, 10, resetSeqStmts, &sqlmanager_shared.BatchExecOpts{})
					if err != nil {
						// handle not found errors
						if !strings.Contains(err.Error(), `does not exist`) {
							return nil, fmt.Errorf("unable to exec postgres sequence reset statements: %w", err)
						}
					}
				}
			}
		case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
			if sqlopts.InitSchema {
				if sqlopts.InitSchema {
					tables := []*sqlmanager_shared.SchemaTable{}
					for tableKey := range uniqueTables {
						schema, table := sqlmanager_shared.SplitTableKey(tableKey)
						tables = append(tables, &sqlmanager_shared.SchemaTable{Schema: schema, Table: table})
					}

					initblocks, err := sourcedb.Db().GetSchemaInitStatements(ctx, tables)
					if err != nil {
						return nil, err
					}

					for _, block := range initblocks {
						slogger.Info(fmt.Sprintf("[%s] found %d statements to execute during schema initialization", block.Label, len(block.Statements)))
						if len(block.Statements) == 0 {
							continue
						}
						err = destdb.Db().BatchExec(ctx, batchSizeConst, block.Statements, &sqlmanager_shared.BatchExecOpts{})
						if err != nil {
							slogger.Error(fmt.Sprintf("unable to exec mysql %s statements: %s", block.Label, err.Error()))
							if block.Label != sqlmanager_mysql.SchemasLabel {
								return nil, fmt.Errorf("unable to exec mysql %s statements: %w", block.Label, err)
							}
						}
					}
				}
			}
			// truncate statements
			if sqlopts.TruncateBeforeInsert {
				tableTruncate := []string{}
				for table := range uniqueTables {
					schema, table := sqlmanager_shared.SplitTableKey(table)
					stmt, err := sqlmanager_mysql.BuildMysqlTruncateStatement(schema, table)
					if err != nil {
						return nil, err
					}
					tableTruncate = append(tableTruncate, stmt)
				}
				slogger.Info(fmt.Sprintf("executing %d sql statements that will truncate tables", len(tableTruncate)))
				disableFkChecks := sqlmanager_shared.DisableForeignKeyChecks
				err = destdb.Db().BatchExec(ctx, batchSizeConst, tableTruncate, &sqlmanager_shared.BatchExecOpts{Prefix: &disableFkChecks})
				if err != nil {
					return nil, err
				}
			}
			destdb.Db().Close()
		case *mgmtv1alpha1.ConnectionConfig_AwsS3Config, *mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig:
			// nothing to do here
		case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
			// init statements
			if sqlopts.InitSchema {
				if !b.eelicense.IsValid() {
					return nil, fmt.Errorf("invalid or non-existent Neosync License. SQL Server schema init requires valid Enterprise license.")
				}
				tables := []*sqlmanager_shared.SchemaTable{}
				for tableKey := range uniqueTables {
					schema, table := sqlmanager_shared.SplitTableKey(tableKey)
					tables = append(tables, &sqlmanager_shared.SchemaTable{Schema: schema, Table: table})
				}

				initblocks, err := sourcedb.Db().GetSchemaInitStatements(ctx, tables)
				if err != nil {
					return nil, err
				}

				initErrors := []*InitSchemaError{}
				for _, block := range initblocks {
					slogger.Info(fmt.Sprintf("[%s] found %d statements to execute during schema initialization", block.Label, len(block.Statements)))
					if len(block.Statements) == 0 {
						continue
					}
					for _, stmt := range block.Statements {
						err = destdb.Db().Exec(ctx, stmt)
						if err != nil {
							slogger.Error(fmt.Sprintf("unable to exec mssql %s statements: %s", block.Label, err.Error()))
							initErrors = append(initErrors, &InitSchemaError{
								Statement: stmt,
								Error:     err.Error(),
							})
							if block.Label != ee_sqlmanager_mssql.SchemasLabel && block.Label != ee_sqlmanager_mssql.ViewsFunctionsLabel && block.Label != ee_sqlmanager_mssql.TableIndexLabel {
								return nil, fmt.Errorf("unable to exec mssql %s statements: %w", block.Label, err)
							}
						}
					}
				}
				initSchemaRunContext = append(initSchemaRunContext, &InitSchemaRunContext{
					ConnectionId: destination.ConnectionId,
					Errors:       initErrors,
				})
			}

			// truncate statements
			if sqlopts.TruncateBeforeInsert {
				tableDependencies, err := sourcedb.Db().GetTableConstraintsBySchema(ctx, uniqueSchemas)
				if err != nil {
					return nil, fmt.Errorf("unable to retrieve database foreign key constraints: %w", err)
				}
				slogger.Info(fmt.Sprintf("found %d foreign key constraints for database", len(tableDependencies.ForeignKeyConstraints)))
				tablePrimaryDependencyMap := getFilteredForeignToPrimaryTableMap(tableDependencies.ForeignKeyConstraints, uniqueTables)
				orderedTablesResp, err := tabledependency.GetTablesOrderedByDependency(tablePrimaryDependencyMap)
				if err != nil {
					return nil, err
				}

				orderedTableDelete := []string{}
				for i := len(orderedTablesResp.OrderedTables) - 1; i >= 0; i-- {
					st := orderedTablesResp.OrderedTables[i]
					stmt, err := sqlmanager_mssql.BuildMssqlDeleteStatement(st.Schema, st.Table)
					if err != nil {
						return nil, err
					}
					orderedTableDelete = append(orderedTableDelete, stmt)
				}

				slogger.Info(fmt.Sprintf("executing %d sql statements that will delete from tables", len(orderedTableDelete)))
				err = destdb.Db().BatchExec(ctx, 10, orderedTableDelete, &sqlmanager_shared.BatchExecOpts{})
				if err != nil {
					return nil, fmt.Errorf("unable to exec ordered delete from statements: %w", err)
				}

				// reset identity column counts
				schemaColMap, err := sourcedb.Db().GetSchemaColumnMap(ctx)
				if err != nil {
					return nil, err
				}

				identityStmts := []string{}
				for table, cols := range schemaColMap {
					if _, ok := uniqueTables[table]; !ok {
						continue
					}
					for _, c := range cols {
						if c.IdentityGeneration != nil && *c.IdentityGeneration != "" {
							schema, table := sqlmanager_shared.SplitTableKey(table)
							identityResetStatement := sqlmanager_mssql.BuildMssqlIdentityColumnResetStatement(schema, table, c.IdentitySeed, c.IdentityIncrement)
							identityStmts = append(identityStmts, identityResetStatement)
						}
					}
				}
				if len(identityStmts) > 0 {
					err = destdb.Db().BatchExec(ctx, 10, identityStmts, &sqlmanager_shared.BatchExecOpts{})
					if err != nil {
						return nil, fmt.Errorf("unable to exec identity reset statements: %w", err)
					}
				}
			}
		default:
			return nil, fmt.Errorf("unsupported destination connection config: %T", destinationConnection.ConnectionConfig.Config)
		}
	}

	err = b.setInitSchemaRunCtx(ctx, initSchemaRunContext, job.AccountId)
	if err != nil {
		return nil, err
	}

	return &RunSqlInitTableStatementsResponse{}, nil
}

type InitSchemaRunContext struct {
	ConnectionId string
	Errors       []*InitSchemaError
}
type InitSchemaError struct {
	Statement string
	Error     string
}

func (b *initStatementBuilder) setInitSchemaRunCtx(
	ctx context.Context,
	initschemaRunContexts []*InitSchemaRunContext,
	accountId string,
) error {
	bits, err := json.Marshal(initschemaRunContexts)
	if err != nil {
		return fmt.Errorf("failed to marshal init schema run context: %w", err)
	}
	_, err = b.jobclient.SetRunContext(ctx, connect.NewRequest(&mgmtv1alpha1.SetRunContextRequest{
		Id: &mgmtv1alpha1.RunContextKey{
			JobRunId:   b.workflowId,
			ExternalId: "init-schema-report",
			AccountId:  accountId,
		},
		Value: bits,
	}))
	if err != nil {
		return fmt.Errorf("failed to set init schema run context: %w", err)
	}
	return nil
}

func (b *initStatementBuilder) getJobById(
	ctx context.Context,
	jobId string,
) (*mgmtv1alpha1.Job, error) {
	getjobResp, err := b.jobclient.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: jobId,
	}))
	if err != nil {
		return nil, err
	}
	return getjobResp.Msg.Job, nil
}

// filtered by tables found in job mappings
func getFilteredForeignToPrimaryTableMap(td map[string][]*sqlmanager_shared.ForeignConstraint, uniqueTables map[string]struct{}) map[string][]string {
	dpMap := map[string][]string{}
	for table := range uniqueTables {
		_, dpOk := dpMap[table]
		if !dpOk {
			dpMap[table] = []string{}
		}
		constraints, ok := td[table]
		if !ok {
			continue
		}
		for _, dep := range constraints {
			_, ok := uniqueTables[dep.ForeignKey.Table]
			// only add to map if dependency is an included table
			if ok {
				dpMap[table] = append(dpMap[table], dep.ForeignKey.Table)
			}
		}
	}
	return dpMap
}
