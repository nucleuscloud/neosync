package neosync_benthos_sql

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Jeffail/shutdown"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	querybuilder "github.com/nucleuscloud/neosync/worker/pkg/query-builder"
	"github.com/warpstreamlabs/bento/public/bloblang"
	"github.com/warpstreamlabs/bento/public/service"
)

func sqlInsertOutputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("connection_id")).
		Field(service.NewStringField("schema")).
		Field(service.NewStringField("table")).
		Field(service.NewStringListField("columns")).
		Field(service.NewStringListField("column_data_types")).
		Field(service.NewAnyMapField("column_default_properties")).
		Field(service.NewBloblangField("args_mapping").Optional()).
		Field(service.NewBoolField("on_conflict_do_nothing").Optional().Default(false)).
		Field(service.NewBoolField("skip_foreign_key_violations").Optional().Default(false)).
		Field(service.NewBoolField("raw_insert_mode").Optional().Default(false)).
		Field(service.NewBoolField("truncate_on_retry").Optional().Default(false)).
		Field(service.NewIntField("max_in_flight").Default(64)).
		Field(service.NewBatchPolicyField("batching")).
		Field(service.NewStringField("prefix").Optional()).
		Field(service.NewStringField("suffix").Optional()).
		Field(service.NewStringField("max_retry_attempts").Default(3)).
		Field(service.NewStringField("retry_attempt_delay").Default("300ms"))
}

// Registers an output on a benthos environment called pooled_sql_raw
func RegisterPooledSqlInsertOutput(env *service.Environment, dbprovider ConnectionProvider, isRetry bool, logger *slog.Logger) error {
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
			out, err := newInsertOutput(conf, mgr, dbprovider, isRetry, logger)
			if err != nil {
				return nil, service.BatchPolicy{}, -1, err
			}
			return out, batchPolicy, maxInFlight, nil
		},
	)
}

var _ service.BatchOutput = &pooledInsertOutput{}

type pooledInsertOutput struct {
	driver       string
	connectionId string
	provider     ConnectionProvider
	dbMut        sync.RWMutex
	db           mysql_queries.DBTX
	logger       *service.Logger
	slogger      *slog.Logger

	schema                   string
	table                    string
	columns                  []string
	columnDataTypes          []string
	columnDefaultProperties  map[string]*neosync_benthos.ColumnDefaultProperties
	onConflictDoNothing      bool
	skipForeignKeyViolations bool
	rawInsertMode            bool
	truncateOnRetry          bool
	prefix                   *string
	suffix                   *string

	argsMapping *bloblang.Executor
	shutSig     *shutdown.Signaller
	isRetry     bool

	maxRetryAttempts uint
	retryDelay       time.Duration
}

func newInsertOutput(conf *service.ParsedConfig, mgr *service.Resources, provider ConnectionProvider, isRetry bool, logger *slog.Logger) (*pooledInsertOutput, error) {
	connectionId, err := conf.FieldString("connection_id")
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

	columnDataTypes, err := conf.FieldStringList("column_data_types")
	if err != nil {
		return nil, err
	}

	columnDefaultPropertiesConfig, err := conf.FieldAnyMap("column_default_properties")
	if err != nil {
		return nil, err
	}

	columnDefaultProperties := map[string]*neosync_benthos.ColumnDefaultProperties{}
	for key, properties := range columnDefaultPropertiesConfig {
		props, err := properties.FieldAny()
		if err != nil {
			return nil, err
		}
		jsonData, err := json.Marshal(props)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal properties for key %s: %w", key, err)
		}

		var colDefaults neosync_benthos.ColumnDefaultProperties
		if err := json.Unmarshal(jsonData, &colDefaults); err != nil {
			return nil, fmt.Errorf("failed to unmarshal properties for key %s: %w", key, err)
		}

		columnDefaultProperties[key] = &colDefaults
	}

	onConflictDoNothing, err := conf.FieldBool("on_conflict_do_nothing")
	if err != nil {
		return nil, err
	}

	skipForeignKeyViolations, err := conf.FieldBool("skip_foreign_key_violations")
	if err != nil {
		return nil, err
	}

	rawInsertMode, err := conf.FieldBool("raw_insert_mode")
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

	var argsMapping *bloblang.Executor
	if conf.Contains("args_mapping") {
		if argsMapping, err = conf.FieldBloblang("args_mapping"); err != nil {
			return nil, err
		}
	}

	retryAttemptsConf, err := conf.FieldInt("max_retry_attempts")
	if err != nil {
		return nil, err
	}
	retryAttempts := uint(1)
	if retryAttemptsConf > 1 {
		retryAttempts = uint(retryAttemptsConf)
	}
	retryAttemptDelay, err := conf.FieldString("retry_attempt_delay")
	if err != nil {
		return nil, err
	}
	retryDelay, err := time.ParseDuration(retryAttemptDelay)
	if err != nil {
		return nil, err
	}
	driver, err := provider.GetDriver(connectionId)
	if err != nil {
		return nil, err
	}

	output := &pooledInsertOutput{
		connectionId:             connectionId,
		driver:                   driver,
		logger:                   mgr.Logger(),
		slogger:                  logger,
		shutSig:                  shutdown.NewSignaller(),
		argsMapping:              argsMapping,
		provider:                 provider,
		schema:                   schema,
		table:                    table,
		columns:                  columns,
		columnDataTypes:          columnDataTypes,
		columnDefaultProperties:  columnDefaultProperties,
		onConflictDoNothing:      onConflictDoNothing,
		skipForeignKeyViolations: skipForeignKeyViolations,
		rawInsertMode:            rawInsertMode,
		truncateOnRetry:          truncateOnRetry,
		prefix:                   prefix,
		suffix:                   suffix,
		isRetry:                  isRetry,
		maxRetryAttempts:         retryAttempts,
		retryDelay:               retryDelay,
	}
	return output, nil
}

func (s *pooledInsertOutput) Connect(ctx context.Context) error {
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	if s.db != nil {
		return nil
	}

	db, err := s.provider.GetDb(ctx, s.connectionId)
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

	rows := [][]any{}
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

	// keep same index and order of columns slice
	columnDefaults := make([]*neosync_benthos.ColumnDefaultProperties, len(s.columns))
	for idx, cName := range s.columns {
		defaults, ok := s.columnDefaultProperties[cName]
		if !ok {
			defaults = &neosync_benthos.ColumnDefaultProperties{}
		}
		columnDefaults[idx] = defaults
	}

	options := []querybuilder.InsertOption{
		querybuilder.WithColumnDataTypes(s.columnDataTypes),
		querybuilder.WithColumnDefaults(columnDefaults),
		querybuilder.WithPrefix(s.prefix),
		querybuilder.WithSuffix(s.suffix),
	}

	if s.onConflictDoNothing {
		options = append(options, querybuilder.WithOnConflictDoNothing())
	}
	if s.rawInsertMode {
		options = append(options, querybuilder.WithRawInsertMode())
	}
	builder, err := querybuilder.GetInsertBuilder(
		s.slogger,
		s.driver,
		s.schema,
		s.table,
		s.columns,
		options...,
	)
	if err != nil {
		return err
	}

	insertQuery, args, err := builder.BuildInsertQuery(rows)
	if err != nil {
		return err
	}

	if _, err := s.db.ExecContext(ctx, insertQuery, args...); err != nil {
		shouldRetry := isDeadlockError(err) || (s.skipForeignKeyViolations && neosync_benthos.IsForeignKeyViolationError(err.Error()))
		if !shouldRetry {
			return err
		}

		err = s.RetryInsertRowByRow(ctx, builder, insertQuery, s.columns, rows, columnDefaults)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *pooledInsertOutput) RetryInsertRowByRow(
	ctx context.Context,
	builder querybuilder.InsertQueryBuilder,
	insertQuery string,
	columns []string,
	rows [][]any,
	columnDefaults []*neosync_benthos.ColumnDefaultProperties,
) error {
	fkErrorCount := 0
	insertCount := 0
	preparedInsert, err := builder.BuildPreparedInsertQuerySingleRow()
	if err != nil {
		return err
	}
	args := builder.BuildPreparedInsertArgs(rows)
	for _, row := range args {
		err = s.execWithRetry(ctx, preparedInsert, row)
		if err != nil && neosync_benthos.IsForeignKeyViolationError(err.Error()) {
			fkErrorCount++
		} else if err != nil && !neosync_benthos.IsForeignKeyViolationError(err.Error()) {
			return err
		}
		if err == nil {
			insertCount++
		}
	}
	s.logger.Infof("Completed batch insert with %d foreign key violations. Skipped rows: %d, Successfully inserted: %d", fkErrorCount, fkErrorCount, insertCount)
	return nil
}

func (s *pooledInsertOutput) execWithRetry(
	ctx context.Context,
	query string,
	args []any,
) error {
	config := &retryConfig{
		MaxAttempts: s.maxRetryAttempts,
		RetryDelay:  s.retryDelay,
		Logger:      s.logger,
		ShouldRetry: isDeadlockError,
	}

	operation := func(ctx context.Context) error {
		_, err := s.db.ExecContext(ctx, query, args...)
		return err
	}

	return retryWithConfig(ctx, config, operation)
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
