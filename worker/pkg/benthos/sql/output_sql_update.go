package neosync_benthos_sql

import (
	"context"
	"database/sql"
	"fmt"
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

type SqlDbtx interface {
	mysql_queries.DBTX

	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
}

type DbPoolProvider interface {
	GetDb(driver, dsn string) (SqlDbtx, error)
}

type ConnectionProvider interface {
	GetDb(ctx context.Context, connectionId string) (SqlDbtx, error)
	GetDriver(connectionId string) (string, error)
}

func sqlUpdateOutputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("connection_id")).
		Field(service.NewStringField("schema")).
		Field(service.NewStringField("table")).
		Field(service.NewStringListField("columns")).
		Field(service.NewStringListField("where_columns")).
		Field(service.NewBoolField("skip_foreign_key_violations").Optional().Default(false)).
		Field(service.NewIntField("max_in_flight").Default(64)).
		Field(service.NewBatchPolicyField("batching")).
		Field(service.NewStringField("max_retry_attempts").Default(3)).
		Field(service.NewStringField("retry_attempt_delay").Default("300ms"))
}

// Registers an output on a benthos environment called pooled_sql_raw
func RegisterPooledSqlUpdateOutput(env *service.Environment, dbprovider ConnectionProvider) error {
	return env.RegisterBatchOutput(
		"pooled_sql_update", sqlUpdateOutputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchOutput, service.BatchPolicy, int, error) {
			batchPolicy, err := conf.FieldBatchPolicy("batching")
			if err != nil {
				return nil, batchPolicy, -1, err
			}

			maxInFlight, err := conf.FieldInt("max_in_flight")
			if err != nil {
				return nil, service.BatchPolicy{}, -1, err
			}
			out, err := newUpdateOutput(conf, mgr, dbprovider)
			if err != nil {
				return nil, service.BatchPolicy{}, -1, err
			}
			return out, batchPolicy, maxInFlight, nil
		},
	)
}

var _ service.BatchOutput = &pooledUpdateOutput{}

type pooledUpdateOutput struct {
	driver       string
	connectionId string
	provider     ConnectionProvider
	dbMut        sync.RWMutex
	db           SqlDbtx
	logger       *service.Logger

	schema                   string
	table                    string
	columns                  []string
	whereCols                []string
	skipForeignKeyViolations bool

	shutSig *shutdown.Signaller

	maxRetryAttempts uint
	retryDelay       time.Duration
}

func newUpdateOutput(conf *service.ParsedConfig, mgr *service.Resources, provider ConnectionProvider) (*pooledUpdateOutput, error) {
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

	whereCols, err := conf.FieldStringList("where_columns")
	if err != nil {
		return nil, err
	}

	skipForeignKeyViolations, err := conf.FieldBool("skip_foreign_key_violations")
	if err != nil {
		return nil, err
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

	output := &pooledUpdateOutput{
		driver:                   driver,
		connectionId:             connectionId,
		logger:                   mgr.Logger(),
		shutSig:                  shutdown.NewSignaller(),
		provider:                 provider,
		schema:                   schema,
		table:                    table,
		columns:                  columns,
		whereCols:                whereCols,
		skipForeignKeyViolations: skipForeignKeyViolations,
		maxRetryAttempts:         retryAttempts,
		retryDelay:               retryDelay,
	}
	return output, nil
}

func (s *pooledUpdateOutput) Connect(ctx context.Context) error {
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

func (s *pooledUpdateOutput) WriteBatch(ctx context.Context, batch service.MessageBatch) error {
	s.dbMut.RLock()
	defer s.dbMut.RUnlock()

	batchLen := len(batch)
	if batchLen == 0 {
		return nil
	}

	for _, msg := range batch {
		m, _ := msg.AsStructured()

		// msgMap has all the table columns and values not just the columns we are updating
		msgMap, ok := m.(map[string]any)
		if !ok {
			return fmt.Errorf("message returned non-map result: %T", msgMap)
		}

		query, err := querybuilder.BuildUpdateQuery(s.driver, s.schema, s.table, s.columns, s.whereCols, msgMap)
		if err != nil {
			return err
		}
		if err := s.execWithRetry(ctx, query); err != nil {
			if !s.skipForeignKeyViolations || !neosync_benthos.IsForeignKeyViolationError(err.Error()) {
				return err
			}
		}
	}

	return nil
}

func (s *pooledUpdateOutput) Close(ctx context.Context) error {
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

func (s *pooledUpdateOutput) execWithRetry(
	ctx context.Context,
	query string,
) error {
	config := &retryConfig{
		MaxAttempts: s.maxRetryAttempts,
		RetryDelay:  s.retryDelay,
		Logger:      s.logger,
		ShouldRetry: isDeadlockError,
	}

	operation := func(ctx context.Context) error {
		_, err := s.db.ExecContext(ctx, query)
		return err
	}

	return retryWithConfig(ctx, config, operation)
}
