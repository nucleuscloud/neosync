package neosync_benthos_sql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/Jeffail/shutdown"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	database_record_mapper "github.com/nucleuscloud/neosync/internal/database-record-mapper"
	record_mapper_builder "github.com/nucleuscloud/neosync/internal/database-record-mapper/builder"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/redpanda-data/benthos/v4/public/service"
)

func sqlRawInputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("connection_id")).
		Field(service.NewStringField("query")).
		Field(service.NewBloblangField("args_mapping").Optional()).
		Field(service.NewBloblangField("expected_total_rows").Optional())
}

// Registers an input on a benthos environment called pooled_sql_raw
func RegisterPooledSqlRawInput(
	env *service.Environment,
	dbprovider ConnectionProvider,
	stopActivityChannel chan<- error,
	onHasMorePages func(ok bool),
) error {
	return env.RegisterInput(
		"pooled_sql_raw",
		sqlRawInputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Input, error) {
			input, err := newInput(conf, mgr, dbprovider, stopActivityChannel, onHasMorePages)
			if err != nil {
				return nil, err
			}
			return input, nil
		},
	)
}

type pooledInput struct {
	provider     ConnectionProvider
	logger       *service.Logger
	connectionId string
	driver       string

	argsMapping *bloblang.Executor
	queryStatic string

	db    mysql_queries.DBTX
	dbMut sync.Mutex

	rows *sql.Rows

	recordMapper record_mapper_builder.DatabaseRecordMapper[any]

	shutSig *shutdown.Signaller

	stopActivityChannel chan<- error
	onHasMorePages      func(ok bool)
	expectedTotalRows   *int
	rowsRead            int
}

func newInput(
	conf *service.ParsedConfig,
	mgr *service.Resources,
	dbprovider ConnectionProvider,
	channel chan<- error,
	onHasMoreResults func(ok bool),
) (*pooledInput, error) {
	connectionId, err := conf.FieldString("connection_id")
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

	var expectedTotalRows *int
	if conf.Contains("expected_total_rows") {
		totalRows, err := conf.FieldInt("expected_total_rows")
		if err != nil {
			return nil, err
		}
		expectedTotalRows = &totalRows
	}

	driver, err := dbprovider.GetDriver(connectionId)
	if err != nil {
		return nil, err
	}

	mapper, err := database_record_mapper.NewDatabaseRecordMapper(driver)
	if err != nil {
		return nil, err
	}

	return &pooledInput{
		logger:              mgr.Logger(),
		shutSig:             shutdown.NewSignaller(),
		connectionId:        connectionId,
		driver:              driver,
		queryStatic:         queryStatic,
		argsMapping:         argsMapping,
		provider:            dbprovider,
		stopActivityChannel: channel,
		recordMapper:        mapper,
		onHasMorePages:      onHasMoreResults,
		expectedTotalRows:   expectedTotalRows,
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

	db, err := s.provider.GetDb(ctx, s.connectionId)
	if err != nil {
		return nil
	}
	s.db = db
	s.logger.Debug(fmt.Sprintf("connected to database %s", s.connectionId))

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
	s.rowsRead = 0
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

	if s.db == nil {
		return nil, nil, service.ErrNotConnected
	}
	if s.rows == nil {
		if s.expectedTotalRows != nil && s.onHasMorePages != nil {
			s.onHasMorePages(s.rowsRead >= *s.expectedTotalRows)
		}
		return nil, nil, service.ErrEndOfInput
	}
	if !s.rows.Next() {
		// Check if any error occurred.
		if err := s.rows.Err(); err != nil {
			_ = s.rows.Close()
			s.rows = nil
			return nil, nil, err
		}
		// For non-Postgres drivers, simply close and return EndOfInput
		_ = s.rows.Close()
		s.rows = nil
		if s.expectedTotalRows != nil && s.onHasMorePages != nil {
			s.onHasMorePages(s.rowsRead >= *s.expectedTotalRows)
		}
		return nil, nil, service.ErrEndOfInput
	}

	obj, err := s.recordMapper.MapRecord(s.rows)
	if err != nil {
		_ = s.rows.Close()
		s.rows = nil
		return nil, nil, err
	}

	s.rowsRead++

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
	if isNil {
		s.dbMut.Unlock()
		return nil
	}

	if s.rows != nil {
		_ = s.rows.Close()
		s.rows = nil
	}

	s.db = nil // not closing here since it's managed by the pool

	s.dbMut.Unlock()

	select {
	case <-s.shutSig.HasStoppedChan():
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
