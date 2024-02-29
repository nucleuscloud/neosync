package sql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/nucleuscloud/neosync/worker/internal/benthos/shutdown"
)

func sqlRawOutputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("query")).
		Field(service.NewBoolField("unsafe_dynamic_query").Default(false)).
		Field(service.NewBloblangField("args_mapping").Optional()).
		Field(service.NewIntField("max_in_flight").Default(64)).
		Field(service.NewBatchPolicyField("batching")).
		Field(service.NewStringField("init_statement").Optional())
}

// Registers an output on a benthos environment called pooled_sql_raw
func RegisterPooledSqlRawOutput(env *service.Environment, db *sql.DB) error {
	return env.RegisterBatchOutput(
		"pooled_sql_raw", sqlRawOutputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (out service.BatchOutput, batchPolicy service.BatchPolicy, maxInFlight int, err error) {
			if batchPolicy, err = conf.FieldBatchPolicy("batching"); err != nil {
				return
			}
			if maxInFlight, err = conf.FieldInt("max_in_flight"); err != nil {
				return
			}
			out, err = newOutput(conf, mgr)
			return
		},
	)
}

var _ service.BatchOutput = &pooledOutput{}

type pooledOutput struct {
	db     *sql.DB
	dbMut  sync.RWMutex
	logger *service.Logger

	queryStatic string
	queryDyn    *service.InterpolatedString

	argsMapping *bloblang.Executor
	shutSig     *shutdown.Signaller
}

func newOutput(conf *service.ParsedConfig, mgr *service.Resources) (*pooledOutput, error) {
	queryStatic, err := conf.FieldString("query")
	if err != nil {
		return nil, err
	}

	var queryDyn *service.InterpolatedString
	if unsafeDyn, err := conf.FieldBool("unsafe_dynamic_query"); err != nil {
		return nil, err
	} else if unsafeDyn {
		if queryDyn, err = conf.FieldInterpolatedString("query"); err != nil {
			return nil, err
		}
	}

	var argsMapping *bloblang.Executor
	if conf.Contains("args_mapping") {
		if argsMapping, err = conf.FieldBloblang("args_mapping"); err != nil {
			return nil, err
		}
	}

	output := &pooledOutput{
		logger:      mgr.Logger(),
		shutSig:     shutdown.NewSignaller(),
		queryStatic: queryStatic,
		queryDyn:    queryDyn,
		argsMapping: argsMapping,
	}
	return output, nil
}

func (s *pooledOutput) WithDb(db *sql.DB) {
	s.db = db
}

func (s *pooledOutput) Connect(ctx context.Context) error {
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	// nothing to do here since the database is already connected

	go func() {
		<-s.shutSig.CloseNowChan()

		s.shutSig.ShutdownComplete()
	}()
	return nil
}

func (s *pooledOutput) WriteBatch(ctx context.Context, batch service.MessageBatch) error {
	s.dbMut.RLock()
	defer s.dbMut.RUnlock()

	for i := range batch {
		var args []any
		if s.argsMapping != nil {
			resMsg, err := batch.BloblangQuery(i, s.argsMapping)
			if err != nil {
				return err
			}

			iargs, err := resMsg.AsStructured()
			if err != nil {
				return err
			}

			var ok bool
			if args, ok = iargs.([]any); !ok {
				return fmt.Errorf("mapping returned non-array result: %T", iargs)
			}
		}

		queryStr := s.queryStatic
		if s.queryDyn != nil {
			var err error
			if queryStr, err = batch.TryInterpolatedString(i, s.queryDyn); err != nil {
				return fmt.Errorf("query interpolation error: %w", err)
			}
		}

		if _, err := s.db.ExecContext(ctx, queryStr, args...); err != nil {
			return err
		}
	}
	return nil
}

func (s *pooledOutput) Close(ctx context.Context) error {
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
