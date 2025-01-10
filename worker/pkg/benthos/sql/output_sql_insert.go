package neosync_benthos_sql

import (
	"context"
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
	"github.com/warpstreamlabs/bento/public/service"
)

func sqlInsertOutputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("connection_id")).
		Field(service.NewStringField("schema")).
		Field(service.NewStringField("table")).
		Field(service.NewBoolField("on_conflict_do_nothing").Optional().Default(false)).
		Field(service.NewBoolField("on_conflict_do_update").Optional().Default(false)).
		Field(service.NewBoolField("skip_foreign_key_violations").Optional().Default(false)).
		Field(service.NewBoolField("truncate_on_retry").Optional().Default(false)).
		Field(service.NewBoolField("should_override_column_default").Optional().Default(false)).
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
	onConflictDoNothing      bool
	skipForeignKeyViolations bool
	truncateOnRetry          bool
	queryBuilder             querybuilder.InsertQueryBuilder

	shutSig *shutdown.Signaller
	isRetry bool

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

	onConflictDoNothing, err := conf.FieldBool("on_conflict_do_nothing")
	if err != nil {
		return nil, err
	}

	onConflictDoUpdate, err := conf.FieldBool("on_conflict_do_update")
	if err != nil {
		return nil, err
	}

	skipForeignKeyViolations, err := conf.FieldBool("skip_foreign_key_violations")
	if err != nil {
		return nil, err
	}

	truncateOnRetry, err := conf.FieldBool("truncate_on_retry")
	if err != nil {
		return nil, err
	}

	shouldOverrideColumnDefault, err := conf.FieldBool("should_override_column_default")
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

	options := []querybuilder.InsertOption{
		querybuilder.WithPrefix(prefix),
		querybuilder.WithSuffix(suffix),
	}

	if onConflictDoNothing {
		options = append(options, querybuilder.WithOnConflictDoNothing())
	}
	if onConflictDoUpdate {
		options = append(options, querybuilder.WithOnConflictDoUpdate())
	}
	if shouldOverrideColumnDefault {
		options = append(options, querybuilder.WithShouldOverrideColumnDefault())
	}
	builder, err := querybuilder.GetInsertBuilder(
		logger,
		driver,
		schema,
		table,
		options...,
	)
	if err != nil {
		return nil, err
	}

	output := &pooledInsertOutput{
		connectionId:             connectionId,
		driver:                   driver,
		logger:                   mgr.Logger(),
		slogger:                  logger,
		shutSig:                  shutdown.NewSignaller(),
		provider:                 provider,
		schema:                   schema,
		table:                    table,
		onConflictDoNothing:      onConflictDoNothing,
		queryBuilder:             builder,
		skipForeignKeyViolations: skipForeignKeyViolations,
		truncateOnRetry:          truncateOnRetry,
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

	rows := []map[string]any{}
	for _, msg := range batch {
		m, _ := msg.AsStructured()
		msgMap, ok := m.(map[string]any)
		if !ok {
			return fmt.Errorf("message returned non-map result: %T", msgMap)
		}

		rows = append(rows, msgMap)
	}
	if len(rows) == 0 {
		s.logger.Debug("no rows to insert")
		return nil
	}

	insertQuery, args, err := s.queryBuilder.BuildInsertQuery(rows)
	if err != nil {
		return err
	}

	if _, err := s.db.ExecContext(ctx, insertQuery, args...); err != nil {
		shouldRetry := isDeadlockError(err) || (s.skipForeignKeyViolations && neosync_benthos.IsForeignKeyViolationError(err.Error()))
		if !shouldRetry {
			return err
		}

		err = s.RetryInsertRowByRow(ctx, s.queryBuilder, rows)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *pooledInsertOutput) RetryInsertRowByRow(
	ctx context.Context,
	builder querybuilder.InsertQueryBuilder,
	rows []map[string]any,
) error {
	fkErrorCount := 0
	insertCount := 0
	for _, row := range rows {
		insertQuery, args, err := builder.BuildInsertQuery([]map[string]any{row})
		if err != nil {
			return err
		}
		err = s.execWithRetry(ctx, insertQuery, args)
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
