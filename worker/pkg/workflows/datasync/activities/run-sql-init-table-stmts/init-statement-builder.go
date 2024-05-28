package runsqlinittablestmts_activity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

const (
	batchSizeConst = 20
)

type initStatementBuilder struct {
	sqlmanager sql_manager.SqlManagerClient
	jobclient  mgmtv1alpha1connect.JobServiceClient
	connclient mgmtv1alpha1connect.ConnectionServiceClient
}

func newInitStatementBuilder(
	sqlmanager sql_manager.SqlManagerClient,
	jobclient mgmtv1alpha1connect.JobServiceClient,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
) *initStatementBuilder {
	return &initStatementBuilder{
		sqlmanager: sqlmanager,
		jobclient:  jobclient,
		connclient: connclient,
	}
}

func (b *initStatementBuilder) RunSqlInitTableStatements(
	ctx context.Context,
	req *RunSqlInitTableStatementsRequest,
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
	if job.GetSource().GetOptions().GetAiGenerate() != nil {
		sourceConnection, err = shared.GetConnectionById(ctx, b.connclient, *job.GetSource().GetOptions().GetAiGenerate().FkSourceConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get connection by id: %w", err)
		}
	}

	sourcedb, err := b.sqlmanager.NewPooledSqlDb(ctx, slogger, sourceConnection)
	if err != nil {
		return nil, fmt.Errorf("unable to create new sql db: %w", err)
	}
	defer sourcedb.Db.Close()

	uniqueTables := shared.GetUniqueTablesMapFromJob(job)
	uniqueSchemas := shared.GetUniqueSchemasFromJob(job)

	for _, destination := range job.Destinations {
		destinationConnection, err := shared.GetConnectionById(ctx, b.connclient, destination.ConnectionId)
		if err != nil {
			return nil, fmt.Errorf("unable to get destination connection by id (%s): %w", destination.ConnectionId, err)
		}
		if destinationConnection.GetConnectionConfig().GetAwsS3Config() != nil {
			// nothing to do for AWS S3 destination
			continue
		}
		sqlopts, err := getSqlJobDestinationOpts(destination.GetOptions())
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

		destdb, err := b.sqlmanager.NewPooledSqlDb(ctx, slogger, destinationConnection)
		if err != nil {
			return nil, fmt.Errorf("unable to create new sql db: %w", err)
		}

		switch destinationConnection.ConnectionConfig.Config.(type) {
		case *mgmtv1alpha1.ConnectionConfig_PgConfig:
			if sqlopts.InitSchema {
				tables := []*sql_manager.SchemaTable{}
				for tableKey := range uniqueTables {
					schema, table := shared.SplitTableKey(tableKey)
					tables = append(tables, &sql_manager.SchemaTable{Schema: schema, Table: table})
				}
				initStatementCfgs, err := sourcedb.Db.GetTableInitStatements(ctx, tables)
				if err != nil {
					return nil, err
				}
				createTables := []string{}
				nonFkAlterStmts := []string{}
				fkAlterStmts := []string{}
				for _, stmtCfg := range initStatementCfgs {
					createTables = append(createTables, stmtCfg.CreateTableStatement)
					for _, alter := range stmtCfg.AlterTableStatements {
						if alter.ConstraintType == sql_manager.ForeignConstraintType {
							fkAlterStmts = append(fkAlterStmts, alter.Statement)
						} else {
							nonFkAlterStmts = append(nonFkAlterStmts, alter.Statement)
						}
					}
				}

				destPgConfig := destinationConnection.ConnectionConfig.GetPgConfig()
				if destPgConfig == nil {
					continue
				}
				if len(createTables) > 0 {
					err = destdb.Db.BatchExec(ctx, batchSizeConst, createTables, &sql_manager.BatchExecOpts{})
					if err != nil {
						destdb.Db.Close()
						return nil, fmt.Errorf("unable to exec pg create table statements: %w", err)
					}
				}

				if len(nonFkAlterStmts) > 0 {
					err = destdb.Db.BatchExec(ctx, batchSizeConst, nonFkAlterStmts, &sql_manager.BatchExecOpts{})
					if err != nil {
						destdb.Db.Close()
						return nil, fmt.Errorf("unable to exec pg alter table statements: %w", err)
					}
				}
				if len(fkAlterStmts) > 0 {
					err = destdb.Db.BatchExec(ctx, batchSizeConst, fkAlterStmts, &sql_manager.BatchExecOpts{})
					if err != nil {
						destdb.Db.Close()
						return nil, fmt.Errorf("unable to exec pg fk alter table statements: %w", err)
					}
				}
			}
			// truncate statements
			if sqlopts.TruncateCascade {
				tableTruncateStmts := []string{}
				for table := range uniqueTables {
					split := strings.Split(table, ".")
					stmt, err := sql_manager.BuildPgTruncateCascadeStatement(split[0], split[1])
					if err != nil {
						destdb.Db.Close()
						return nil, err
					}
					tableTruncateStmts = append(tableTruncateStmts, stmt)
				}
				slogger.Info(fmt.Sprintf("executing %d sql statements that will truncate cascade tables", len(tableTruncateStmts)))
				err = destdb.Db.BatchExec(ctx, batchSizeConst, tableTruncateStmts, &sql_manager.BatchExecOpts{})
				if err != nil {
					destdb.Db.Close()
					return nil, fmt.Errorf("unable to exec truncate cascade statements: %w", err)
				}
			} else if sqlopts.TruncateBeforeInsert {
				tableDependencies, err := sourcedb.Db.GetForeignKeyConstraintsMap(ctx, uniqueSchemas)
				if err != nil {
					return nil, fmt.Errorf("unable to retrieve database foreign key constraints: %w", err)
				}
				slogger.Info(fmt.Sprintf("found %d foreign key constraints for database", len(tableDependencies)))
				tablePrimaryDependencyMap := getFilteredForeignToPrimaryTableMap(tableDependencies, uniqueTables)
				orderedTablesResp, err := tabledependency.GetTablesOrderedByDependency(tablePrimaryDependencyMap)
				if err != nil {
					destdb.Db.Close()
					return nil, err
				}

				orderedTableTruncate := []string{}
				for _, table := range orderedTablesResp.OrderedTables {
					split := strings.Split(table, ".")
					orderedTableTruncate = append(orderedTableTruncate, fmt.Sprintf(`%q.%q`, split[0], split[1]))
				}
				slogger.Info(fmt.Sprintf("executing %d sql statements that will truncate tables", len(orderedTableTruncate)))
				truncateStmt := sql_manager.BuildPgTruncateStatement(orderedTableTruncate)
				err = destdb.Db.Exec(ctx, truncateStmt)
				if err != nil {
					destdb.Db.Close()
					return nil, fmt.Errorf("unable to exec ordered truncate statements: %w", err)
				}
			}
			destdb.Db.Close()
		case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
			if sqlopts.InitSchema {
				tableDependencies, err := sourcedb.Db.GetForeignKeyConstraintsMap(ctx, uniqueSchemas)
				if err != nil {
					return nil, fmt.Errorf("unable to retrieve database foreign key constraints: %w", err)
				}
				slogger.Info(fmt.Sprintf("found %d foreign key constraints for database", len(tableDependencies)))
				tableForeignDependencyMap := getFilteredForeignToPrimaryTableMap(tableDependencies, uniqueTables)
				orderedTablesResp, err := tabledependency.GetTablesOrderedByDependency(tableForeignDependencyMap)
				if err != nil {
					destdb.Db.Close()
					return nil, err
				}
				if orderedTablesResp.HasCycles {
					destdb.Db.Close()
					return nil, errors.New("init schema: unable to handle circular dependencies")
				}

				tableCreateStmts := []string{}
				for _, table := range orderedTablesResp.OrderedTables {
					split := strings.Split(table, ".")
					// todo: make this more efficient to reduce amount of times we have to connect to the source database
					initStmt, err := sourcedb.Db.GetCreateTableStatement(
						ctx,
						split[0],
						split[1],
					)
					if err != nil {
						destdb.Db.Close()
						return nil, fmt.Errorf("unable to build init statement for postgres: %w", err)
					}
					tableCreateStmts = append(tableCreateStmts, initStmt)
				}
				slogger.Info(fmt.Sprintf("executing %d sql statements that will initialize tables", len(tableCreateStmts)))
				err = destdb.Db.BatchExec(ctx, batchSizeConst, tableCreateStmts, &sql_manager.BatchExecOpts{})
				if err != nil {
					destdb.Db.Close()
					return nil, fmt.Errorf("unable to exec postgres table create statements: %w", err)
				}
			}
			// truncate statements
			if sqlopts.TruncateBeforeInsert {
				tableTruncate := []string{}
				for table := range uniqueTables {
					split := strings.Split(table, ".")
					stmt, err := sql_manager.BuildMysqlTruncateStatement(split[0], split[1])
					if err != nil {
						destdb.Db.Close()
						return nil, err
					}
					tableTruncate = append(tableTruncate, stmt)
				}
				slogger.Info(fmt.Sprintf("executing %d sql statements that will truncate tables", len(tableTruncate)))
				disableFkChecks := sql_manager.DisableForeignKeyChecks
				err = destdb.Db.BatchExec(ctx, batchSizeConst, tableTruncate, &sql_manager.BatchExecOpts{Prefix: &disableFkChecks})
				if err != nil {
					destdb.Db.Close()
					return nil, err
				}
			}
			destdb.Db.Close()
		case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
			// nothing to do here
		default:
			return nil, fmt.Errorf("unsupported destination connection config")
		}
	}

	return &RunSqlInitTableStatementsResponse{}, nil
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
func getFilteredForeignToPrimaryTableMap(td map[string][]*sql_manager.ForeignConstraint, uniqueTables map[string]struct{}) map[string][]string {
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

type sqlJobDestinationOpts struct {
	TruncateBeforeInsert bool
	TruncateCascade      bool
	InitSchema           bool
}

func getSqlJobDestinationOpts(
	options *mgmtv1alpha1.JobDestinationOptions,
) (*sqlJobDestinationOpts, error) {
	if options == nil {
		return &sqlJobDestinationOpts{}, nil
	}
	switch opts := options.GetConfig().(type) {
	case *mgmtv1alpha1.JobDestinationOptions_PostgresOptions:
		return &sqlJobDestinationOpts{
			TruncateBeforeInsert: opts.PostgresOptions.GetTruncateTable().GetTruncateBeforeInsert(),
			TruncateCascade:      opts.PostgresOptions.GetTruncateTable().GetCascade(),
			InitSchema:           opts.PostgresOptions.GetInitTableSchema(),
		}, nil
	case *mgmtv1alpha1.JobDestinationOptions_MysqlOptions:
		return &sqlJobDestinationOpts{
			TruncateBeforeInsert: opts.MysqlOptions.GetTruncateTable().GetTruncateBeforeInsert(),
			InitSchema:           opts.MysqlOptions.GetInitTableSchema(),
		}, nil
	default:
		return nil, errors.New("unsupported job destination options type")
	}
}
