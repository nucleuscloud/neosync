package neosync_benthos_sql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	"github.com/nucleuscloud/neosync/worker/internal/benthos/shutdown"
)

type SqlDbtx interface {
	mysql_queries.DBTX

	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
}

type DbPoolProvider interface {
	GetDb(driver, dsn string) (SqlDbtx, error)
}

func sqlUpdateOutputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("driver")).
		Field(service.NewStringField("dsn")).
		Field(service.NewStringField("schema")).
		Field(service.NewStringField("table")).
		Field(service.NewStringListField("columns")).
		Field(service.NewStringListField("where_columns")).
		Field(service.NewBloblangField("args_mapping").Optional()).
		Field(service.NewIntField("max_in_flight").Default(64)).
		Field(service.NewBatchPolicyField("batching"))
}

// Registers an output on a benthos environment called pooled_sql_raw
func RegisterPooledSqlUpdateOutput(env *service.Environment, dbprovider DbPoolProvider) error {
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
	driver   string
	dsn      string
	provider DbPoolProvider
	dbMut    sync.RWMutex
	// db       mysql_queries.DBTX
	db     SqlDbtx
	logger *service.Logger

	schema    string
	table     string
	columns   []string
	whereCols []string

	argsMapping *bloblang.Executor
	shutSig     *shutdown.Signaller
}

func newUpdateOutput(conf *service.ParsedConfig, mgr *service.Resources, provider DbPoolProvider) (*pooledUpdateOutput, error) {
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

	whereCols, err := conf.FieldStringList("where_columns")
	if err != nil {
		return nil, err
	}

	var argsMapping *bloblang.Executor
	if conf.Contains("args_mapping") {
		if argsMapping, err = conf.FieldBloblang("args_mapping"); err != nil {
			return nil, err
		}
	}

	output := &pooledUpdateOutput{
		driver:      driver,
		dsn:         dsn,
		logger:      mgr.Logger(),
		shutSig:     shutdown.NewSignaller(),
		argsMapping: argsMapping,
		provider:    provider,
		schema:      schema,
		table:       table,
		columns:     columns,
		whereCols:   whereCols,
	}
	return output, nil
}

func (s *pooledUpdateOutput) Connect(ctx context.Context) error {
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

	go func() {
		<-s.shutSig.CloseNowChan()

		s.dbMut.Lock()
		// not closing the connection here as that is managed by an outside force
		s.db = nil
		s.dbMut.Unlock()

		s.shutSig.ShutdownComplete()
	}()
	return nil
}

// how to handle defaults??
func (s *pooledUpdateOutput) WriteBatch(ctx context.Context, batch service.MessageBatch) error {
	s.dbMut.RLock()
	defer s.dbMut.RUnlock()

	batchLen := len(batch)
	if batchLen == 0 {
		return nil
	}
	builder := goqu.Dialect(s.driver)
	table := goqu.S(s.schema).Table(s.table)

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

		allCols := []string{}
		allCols = append(allCols, s.columns...)
		allCols = append(allCols, s.whereCols...)

		colValMap := map[string]any{}
		for idx, col := range allCols {
			colValMap[col] = args[idx]
		}

		updateRecord := goqu.Record{}
		for _, col := range s.columns {
			val := colValMap[col]
			// set any default transformations
			if val == "DEFAULT" {
				updateRecord[col] = goqu.L("DEFAULT")
			} else {
				updateRecord[col] = val
			}
		}

		where := []exp.Expression{}
		for _, col := range s.whereCols {
			val := colValMap[col]
			where = append(where, goqu.Ex{col: val})
		}

		update := builder.Update(table).
			Set(updateRecord).
			Where(where...)

		query, args, err := update.ToSQL()
		if err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}

	return nil
}

func (s *pooledUpdateOutput) Close(ctx context.Context) error {
	s.shutSig.CloseNow()
	s.dbMut.RLock()
	isNil := s.db == nil
	s.dbMut.RUnlock()
	if isNil {
		return nil
	}
	select {
	case <-s.shutSig.HasClosedChan():
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
