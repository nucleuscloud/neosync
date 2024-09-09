package neosync_benthos_sql

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Jeffail/shutdown"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	sqlserverutil "github.com/nucleuscloud/neosync/internal/sqlserver"
	querybuilder "github.com/nucleuscloud/neosync/worker/pkg/query-builder"
	"github.com/warpstreamlabs/bento/public/bloblang"
	"github.com/warpstreamlabs/bento/public/service"
)

func sqlInsertOutputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("driver")).
		Field(service.NewStringField("dsn")).
		Field(service.NewStringField("schema")).
		Field(service.NewStringField("table")).
		Field(service.NewStringListField("columns")).
		Field(service.NewBloblangField("args_mapping").Optional()).
		Field(service.NewBoolField("on_conflict_do_nothing").Optional().Default(false)).
		Field(service.NewBoolField("truncate_on_retry").Optional().Default(false)).
		Field(service.NewIntField("max_in_flight").Default(64)).
		Field(service.NewBatchPolicyField("batching")).
		Field(service.NewStringField("prefix").Optional()).
		Field(service.NewStringField("suffix").Optional()).
		Field(service.NewStringListField("identity_columns").Optional())
}

// Registers an output on a benthos environment called pooled_sql_raw
func RegisterPooledSqlInsertOutput(env *service.Environment, dbprovider DbPoolProvider, isRetry bool) error {
	return env.RegisterBatchOutput(
		"pooled_sql_insert", sqlInsertOutputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchOutput, service.BatchPolicy, int, error) {
			batchPolicy, err := conf.FieldBatchPolicy("batching")
			if err != nil {
				return nil, batchPolicy, -1, err
			}

			maxInFlight, err := conf.FieldInt("max_in_flight")
			if err != nil {
				return nil, service.BatchPolicy{}, -1, err
			}
			out, err := newInsertOutput(conf, mgr, dbprovider, isRetry)
			if err != nil {
				return nil, service.BatchPolicy{}, -1, err
			}
			return out, batchPolicy, maxInFlight, nil
		},
	)
}

func init() {
	dbprovider := NewDbPoolProvider()
	err := service.RegisterBatchOutput(
		"pooled_sql_insert", sqlInsertOutputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchOutput, service.BatchPolicy, int, error) {
			batchPolicy, err := conf.FieldBatchPolicy("batching")
			if err != nil {
				return nil, batchPolicy, -1, err
			}

			maxInFlight, err := conf.FieldInt("max_in_flight")
			if err != nil {
				return nil, service.BatchPolicy{}, -1, err
			}
			out, err := newInsertOutput(conf, mgr, dbprovider, false)
			if err != nil {
				return nil, service.BatchPolicy{}, -1, err
			}
			return out, batchPolicy, maxInFlight, nil
		})
	if err != nil {
		panic(err)
	}
}

var _ service.BatchOutput = &pooledInsertOutput{}

type pooledInsertOutput struct {
	driver   string
	dsn      string
	provider DbPoolProvider
	dbMut    sync.RWMutex
	db       mysql_queries.DBTX
	logger   *service.Logger

	schema              string
	table               string
	columns             []string
	identityColumns     []string
	onConflictDoNothing bool
	truncateOnRetry     bool
	prefix              *string
	suffix              *string

	argsMapping *bloblang.Executor
	shutSig     *shutdown.Signaller
	isRetry     bool
}

func newInsertOutput(conf *service.ParsedConfig, mgr *service.Resources, provider DbPoolProvider, isRetry bool) (*pooledInsertOutput, error) {
	driver, err := conf.FieldString("driver")
	if err != nil {
		return nil, err
	}
	dsn, err := conf.FieldString("dsn")
	if err != nil {
		return nil, err
	}

	schema, err := conf.FieldString("schema")
	if err != nil {
		return nil, err
	}

	table, err := conf.FieldString("table")
	if err != nil {
		return nil, err
	}

	columns, err := conf.FieldStringList("columns")
	if err != nil {
		return nil, err
	}

	onConflictDoNothing, err := conf.FieldBool("on_conflict_do_nothing")
	if err != nil {
		return nil, err
	}

	truncateOnRetry, err := conf.FieldBool("truncate_on_retry")
	if err != nil {
		return nil, err
	}

	var prefix *string
	if conf.Contains("prefix") {
		prefixStr, err := conf.FieldString("prefix")
		if err != nil {
			return nil, err
		}
		prefix = &prefixStr
	}

	var suffix *string
	if conf.Contains("suffix") {
		suffixStr, err := conf.FieldString("suffix")
		if err != nil {
			return nil, err
		}
		suffix = &suffixStr
	}

	var identityColumns []string
	if conf.Contains("identity_columns") {
		identityCols, err := conf.FieldStringList("identity_columns")
		if err != nil {
			return nil, err
		}
		identityColumns = identityCols
	}

	var argsMapping *bloblang.Executor
	if conf.Contains("args_mapping") {
		if argsMapping, err = conf.FieldBloblang("args_mapping"); err != nil {
			return nil, err
		}
	}

	output := &pooledInsertOutput{
		driver:              driver,
		dsn:                 dsn,
		logger:              mgr.Logger(),
		shutSig:             shutdown.NewSignaller(),
		argsMapping:         argsMapping,
		provider:            provider,
		schema:              schema,
		table:               table,
		columns:             columns,
		identityColumns:     identityColumns,
		onConflictDoNothing: onConflictDoNothing,
		truncateOnRetry:     truncateOnRetry,
		prefix:              prefix,
		suffix:              suffix,
		isRetry:             isRetry,
	}
	return output, nil
}

func (s *pooledInsertOutput) Connect(ctx context.Context) error {
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	if s.db != nil {
		return nil
	}

	db, err := s.provider.GetDb(s.driver, s.dsn)
	if err != nil {
		return err
	}
	s.db = db

	// truncate table on retry
	if s.isRetry && s.truncateOnRetry && !s.onConflictDoNothing {
		s.logger.Info("retry: truncating table before inserting")
		query, err := querybuilder.BuildTruncateQuery(s.driver, fmt.Sprintf("%s.%s", s.schema, s.table))
		if err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	go func() {
		<-s.shutSig.HardStopChan()

		s.dbMut.Lock()
		// not closing the connection here as that is managed by an outside force
		s.db = nil
		s.dbMut.Unlock()

		s.shutSig.TriggerHasStopped()
	}()
	return nil
}

func (s *pooledInsertOutput) WriteBatch(ctx context.Context, batch service.MessageBatch) error {
	s.dbMut.RLock()
	defer s.dbMut.RUnlock()

	batchLen := len(batch)
	if batchLen == 0 {
		return nil
	}

	var executor *service.MessageBatchBloblangExecutor
	if s.argsMapping != nil {
		executor = batch.BloblangExecutor(s.argsMapping)
	}

	rows := [][]interface{}{} //nolint:gofmt
	for i := range batch {
		if s.argsMapping == nil {
			continue
		}
		resMsg, err := executor.Query(i)
		if err != nil {
			return err
		}

		iargs, err := resMsg.AsStructured()
		if err != nil {
			return err
		}

		args, ok := iargs.([]any)
		if !ok {
			return fmt.Errorf("mapping returned non-array result: %T", iargs)
		}

		rows = append(rows, args)
	}

	processedCols, processedRows := s.processRows(s.columns, rows)
	insertQuery, err := querybuilder.BuildInsertQuery(s.driver, fmt.Sprintf("%s.%s", s.schema, s.table), processedCols, processedRows, &s.onConflictDoNothing)
	if err != nil {
		return err
	}

	if s.driver == sqlmanager_shared.MssqlDriver && len(processedCols) == 0 {
		insertQuery = sqlserverutil.GeSqlServerDefaultValuesInsertSql(s.schema, s.table, len(rows))
	}

	query := s.buildQuery(insertQuery)
	if _, err := s.db.ExecContext(ctx, query); err != nil {
		return err
	}
	return nil
}

func (s *pooledInsertOutput) processRows(columnNames []string, dataRows [][]any) (columns []string, rows [][]any) {
	switch s.driver {
	case sqlmanager_shared.MssqlDriver:
		newDataRows := sqlserverutil.GoTypeToSqlServerType(dataRows)
		return sqlserverutil.FilterOutSqlServerDefaultIdentityColumns(s.driver, s.identityColumns, s.columns, newDataRows)
	default:
		return columnNames, dataRows
	}
}

func (s *pooledInsertOutput) buildQuery(insertQuery string) string {
	var query string
	if s.prefix != nil {
		query = *s.prefix
	}

	query += strings.TrimSuffix(insertQuery, ";") + ";"

	if s.suffix != nil {
		query += *s.suffix
	}
	return query
}

func (s *pooledInsertOutput) Close(ctx context.Context) error {
	s.shutSig.TriggerHardStop()
	s.dbMut.RLock()
	isNil := s.db == nil
	s.dbMut.RUnlock()
	if isNil {
		return nil
	}
	select {
	case <-s.shutSig.HasStoppedChan():
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
