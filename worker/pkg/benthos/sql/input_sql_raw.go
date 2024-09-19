package neosync_benthos_sql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/Jeffail/shutdown"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	myutil "github.com/nucleuscloud/neosync/internal/mysql"
	pgutil "github.com/nucleuscloud/neosync/internal/postgres"
	sqlserverutil "github.com/nucleuscloud/neosync/internal/sqlserver"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/warpstreamlabs/bento/public/bloblang"
	"github.com/warpstreamlabs/bento/public/service"
)

func sqlRawInputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("driver")).
		Field(service.NewStringField("dsn")).
		Field(service.NewStringField("query")).
		Field(service.NewBloblangField("args_mapping").Optional())
}

// Registers an input on a benthos environment called pooled_sql_raw
func RegisterPooledSqlRawInput(env *service.Environment, dbprovider DbPoolProvider, stopActivityChannel chan<- error) error {
	return env.RegisterInput(
		"pooled_sql_raw", sqlRawInputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Input, error) {
			input, err := newInput(conf, mgr, dbprovider, stopActivityChannel)
			if err != nil {
				return nil, err
			}
			return input, nil
		},
	)
}

type pooledInput struct {
	driver   string
	dsn      string
	provider DbPoolProvider
	logger   *service.Logger

	argsMapping *bloblang.Executor
	queryStatic string

	db    mysql_queries.DBTX
	dbMut sync.Mutex
	rows  *sql.Rows

	shutSig *shutdown.Signaller

	stopActivityChannel chan<- error
}

func newInput(conf *service.ParsedConfig, mgr *service.Resources, dbprovider DbPoolProvider, channel chan<- error) (*pooledInput, error) {
	driver, err := conf.FieldString("driver")
	if err != nil {
		return nil, err
	}

	dsn, err := conf.FieldString("dsn")
	if err != nil {
		return nil, err
	}

	queryStatic, err := conf.FieldString("query")
	if err != nil {
		return nil, err
	}

	var argsMapping *bloblang.Executor
	if conf.Contains("args_mapping") {
		argsMapping, err = conf.FieldBloblang("args_mapping")
		if err != nil {
			return nil, err
		}
	}

	return &pooledInput{
		logger:              mgr.Logger(),
		shutSig:             shutdown.NewSignaller(),
		driver:              driver,
		dsn:                 dsn,
		queryStatic:         queryStatic,
		argsMapping:         argsMapping,
		provider:            dbprovider,
		stopActivityChannel: channel,
	}, nil
}

var _ service.Input = &pooledInput{}

func (s *pooledInput) Connect(ctx context.Context) error {
	s.logger.Debug("connecting to pooled input")
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	if s.db != nil {
		return nil
	}

	db, err := s.provider.GetDb(s.driver, s.dsn)
	if err != nil {
		return nil
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

	rows, err := db.QueryContext(ctx, s.queryStatic, args...)
	if err != nil {
		if neosync_benthos.IsCriticalError(err.Error()) {
			s.logger.Error(fmt.Sprintf("Benthos input error - sending stop activity signal: %s ", err.Error()))
			s.stopActivityChannel <- err
		}
		return err
	}

	s.rows = rows
	go func() {
		<-s.shutSig.HardStopChan()

		s.dbMut.Lock()
		if s.rows != nil {
			_ = s.rows.Close()
			s.rows = nil
		}
		// not closing the connection here as that is managed by an outside force
		s.db = nil
		s.dbMut.Unlock()

		s.shutSig.TriggerHasStopped()
	}()
	return nil
}

func (s *pooledInput) Read(ctx context.Context) (*service.Message, service.AckFunc, error) {
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	if s.db == nil && s.rows == nil {
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

	obj, err := sqlRowToMap(s.rows, s.driver)
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
	s.shutSig.TriggerHardStop()
	s.dbMut.Lock()
	isNil := s.db == nil
	s.dbMut.Unlock()
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

func sqlRowToMap(rows *sql.Rows, driver string) (map[string]any, error) {
	switch driver {
	case sqlmanager_shared.PostgresDriver:
		return pgutil.SqlRowToPgTypesMap(rows)
	case sqlmanager_shared.MssqlDriver:
		return sqlserverutil.SqlRowToSqlServerTypesMap(rows)
	default:
		return myutil.MysqlSqlRowToMap(rows)
	}
}
