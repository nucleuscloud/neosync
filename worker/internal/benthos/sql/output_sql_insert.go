package neosync_benthos_sql

import (
	"context"
	"fmt"
	"sync"

	"github.com/Jeffail/shutdown"
	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
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
		Field(service.NewBatchPolicyField("batching"))
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
	onConflictDoNothing bool
	truncateOnRetry     bool

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
		onConflictDoNothing: onConflictDoNothing,
		truncateOnRetry:     truncateOnRetry,
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
		builder := goqu.Dialect(s.driver)
		table := goqu.S(s.schema).Table(s.table)
		truncate := builder.Truncate(table)
		query, _, err := truncate.ToSQL()
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
	builder := goqu.Dialect(s.driver)
	table := goqu.S(s.schema).Table(s.table)
	insertCols := make([]any, len(s.columns))
	for i, col := range s.columns {
		insertCols[i] = col
	}
	insert := builder.Insert(table).Cols(insertCols...)
	rows := [][]interface{}{} //nolint:gofmt

	for i := range batch {
		if s.argsMapping == nil {
			continue
		}
		resMsg, err := batch.BloblangQuery(i, s.argsMapping)
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

		// set any default transformations
		for idx, a := range args {
			if a == "DEFAULT" {
				args[idx] = goqu.L("DEFAULT")
			}
		}

		rows = append(rows, args)
	}
	// add rows to the dataset
	for _, row := range rows {
		insert = insert.Vals(row)
	}
	// adds on conflict do nothing to insert query
	if s.onConflictDoNothing {
		insert = insert.OnConflict(goqu.DoNothing())
	}

	query, args, err := insert.ToSQL()
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		fmt.Println("failed to insert", query, args)
		return err
	}
	return nil
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
