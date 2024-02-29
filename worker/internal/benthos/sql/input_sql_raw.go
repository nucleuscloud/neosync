package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/nucleuscloud/neosync/worker/internal/benthos/shutdown"
)

// Registers an input on a benthos environment called pooled_sql_raw
func RegisterPooledSqlRawInput(env *service.Environment, db *sql.DB) error {
	spec := service.NewConfigSpec().
		Field(service.NewStringField("query")).
		Field(service.NewBloblangField("args_mapping").Optional())
	return env.RegisterInput(
		"pooled_sql_raw", spec,
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Input, error) {
			input, err := newInput(conf, mgr)
			if err != nil {
				return nil, err
			}
			input.WithSqlDb(db)
			return input, nil
		},
	)
}

type pooledInput struct {
	sqlpool *sql.DB
	logger  *service.Logger

	argsMapping *bloblang.Executor

	queryStatic string

	dbMut sync.Mutex
	rows  *sql.Rows

	shutSig *shutdown.Signaller
}

func newInput(conf *service.ParsedConfig, mgr *service.Resources) (*pooledInput, error) {
	input := &pooledInput{logger: mgr.Logger(), shutSig: shutdown.NewSignaller()}
	var err error
	if input.queryStatic, err = conf.FieldString("query"); err != nil {
		return nil, err
	}

	if conf.Contains("args_mapping") {
		if input.argsMapping, err = conf.FieldBloblang("args_mapping"); err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	return input, nil
}

func (s *pooledInput) WithSqlDb(db *sql.DB) {
	s.sqlpool = db
}

var _ service.Input = &pooledInput{}

func (s *pooledInput) Connect(ctx context.Context) error {
	s.logger.Info("connecting to pooled input")
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	if s.sqlpool == nil {
		return errors.New("must provide either sql or pgx pool to continue")
	}

	var args []any
	if s.argsMapping != nil {
		iargs, err := s.argsMapping.Query(nil)
		if err != nil {
			return err
		}
		var ok bool
		if args, ok = iargs.([]any); !ok {
			return fmt.Errorf("mapping returned non-array result: %T", iargs)
		}
	}

	rows, err := s.sqlpool.Query(s.queryStatic, args...)
	if err != nil {
		return err
	}
	s.rows = rows
	go func() {
		<-s.shutSig.CloseNowChan()

		s.dbMut.Lock()
		if s.rows != nil {
			_ = s.rows.Close()
			s.rows = nil
		}
		if s.sqlpool != nil {
			s.sqlpool = nil
		}
		s.dbMut.Unlock()

		s.shutSig.ShutdownComplete()
	}()

	return nil
}

func (s *pooledInput) Read(ctx context.Context) (*service.Message, service.AckFunc, error) {
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	if s.sqlpool == nil && s.rows == nil {
		return nil, nil, service.ErrNotConnected
	}
	if s.rows == nil {
		return nil, nil, service.ErrEndOfInput
	}
	if !s.rows.Next() {
		err := s.rows.Err()
		if err == nil {
			err = service.ErrEndOfInput
		}
		_ = s.rows.Close()
		s.rows = nil
		return nil, nil, err
	}

	obj, err := sqlRowToMap(s.rows)
	if err != nil {
		_ = s.rows.Close()
		s.rows = nil
		return nil, nil, err
	}

	msg := service.NewMessage(nil)
	msg.SetStructured(obj)
	return msg, emptyAck, nil
}

func emptyAck(ctx context.Context, err error) error {
	// Nacks are handled by AutoRetryNacks because we don't have an explicit
	// ack mechanism right now.
	return nil
}

func (s *pooledInput) Close(ctx context.Context) error {
	s.shutSig.CloseNow()
	s.dbMut.Lock()
	isNil := s.sqlpool == nil
	s.dbMut.Unlock()
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

func sqlRowToMap(rows *sql.Rows) (map[string]any, error) {
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	values := make([]any, len(columnNames))
	valuesWrapped := make([]any, 0, len(columnNames))
	for i := range values {
		valuesWrapped = append(valuesWrapped, &values[i])
	}
	if err := rows.Scan(valuesWrapped...); err != nil {
		return nil, err
	}
	jObj := map[string]any{}
	for i, v := range values {
		col := columnNames[i]
		switch t := v.(type) {
		case string:
			jObj[col] = t
		case []byte:
			jObj[col] = string(t)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			jObj[col] = t
		case float32, float64:
			jObj[col] = t
		case bool:
			jObj[col] = t
		default:
			jObj[col] = t
		}
	}
	return jObj, nil
}
