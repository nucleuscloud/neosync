package neosync_benthos_sql

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/Jeffail/shutdown"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
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
		Field(service.NewBloblangField("args_mapping").Optional())
}

// Registers an input on a benthos environment called pooled_sql_raw
func RegisterPooledSqlRawInput(env *service.Environment, dbprovider ConnectionProvider, stopActivityChannel chan<- error) error {
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
	provider     ConnectionProvider
	logger       *service.Logger
	connectionId string
	driver       string

	argsMapping *bloblang.Executor
	queryStatic string

	db    mysql_queries.DBTX
	dbMut sync.Mutex
	tx    *sql.Tx
	rows  *sql.Rows

	cursorName string

	recordMapper record_mapper_builder.DatabaseRecordMapper[any]

	shutSig *shutdown.Signaller

	stopActivityChannel chan<- error
}

func newInput(conf *service.ParsedConfig, mgr *service.Resources, dbprovider ConnectionProvider, channel chan<- error) (*pooledInput, error) {
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
		cursorName:          generateCursorName(connectionId, queryStatic),
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

	switch s.driver {
	case sqlmanager_shared.PostgresDriver:
		tx, err := db.BeginTx(ctx, &sql.TxOptions{
			ReadOnly: true,
		})
		if err != nil {
			return err
		}
		s.tx = tx
		s.logger.Debug("transaction started")
		cursorStmt := fmt.Sprintf("DECLARE %s CURSOR FOR %s", s.cursorName, s.queryStatic)
		_, err = s.tx.ExecContext(ctx, cursorStmt, args...)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				s.logger.Error(fmt.Sprintf("error rolling back transaction: %s", err.Error()))
			}
			return err
		}

		s.logger.Debug(fmt.Sprintf("cursor declared: %s", s.cursorName))

		err = s.fetchNextBatchFromCursor(ctx)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				s.logger.Error(fmt.Sprintf("error rolling back transaction: %s", err.Error()))
			}
			return err
		}
	default:
		rows, err := db.QueryContext(ctx, s.queryStatic, args...)
		if err != nil {
			if neosync_benthos.IsCriticalError(err.Error()) {
				s.logger.Error(fmt.Sprintf("Benthos input error - sending stop activity signal: %s ", err.Error()))
				s.stopActivityChannel <- err
			}
			return err
		}

		s.rows = rows
	}

	go func() {
		<-s.shutSig.HardStopChan()

		s.dbMut.Lock()
		if s.rows != nil {
			_ = s.rows.Close()
			s.rows = nil
		}
		if s.tx != nil {
			_ = s.tx.Rollback()
			s.tx = nil
		}
		// not closing the connection here as that is managed by an outside force
		s.db = nil
		s.dbMut.Unlock()

		s.shutSig.TriggerHasStopped()
	}()
	return nil
}

// fetchNextBatchFromCursor issues a FETCH command to retrieve the next batch of rows from the cursor.
func (s *pooledInput) fetchNextBatchFromCursor(ctx context.Context) error {
	fetchStmt := fmt.Sprintf("FETCH FORWARD %d FROM %s", 100, s.cursorName) //nolint:gosec // cursor name is hashed and sanitized
	rows, err := s.tx.QueryContext(ctx, fetchStmt)
	if err != nil {
		if neosync_benthos.IsCriticalError(err.Error()) {
			s.logger.Error(fmt.Sprintf("Benthos input error - sending stop activity signal: %s ", err.Error()))
			s.stopActivityChannel <- err
		}
		return err
	}
	// If no rows are returned, rows.Next() will be false.
	s.rows = rows
	return nil
}

// allow only letters, numbers, and underscores.
var cursorNameRegex = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func sanitizeCursorName(s string) string {
	return cursorNameRegex.ReplaceAllString(s, "")
}

// generateCursorName generates a unique cursor name using the provided schema and table names.
// It also sanitizes the input to allow only alphanumeric characters and underscores.
func generateCursorName(connectionId, query string) string {
	// Sanitize the connectionID for safe use in SQL identifiers.
	safeConnID := sanitizeCursorName(connectionId)

	// Create a short hash (first 8 hex digits) from the query string.
	hash := sha256.Sum256([]byte(query))
	shortHash := fmt.Sprintf("%x", hash)[:8]

	// Use current timestamp in nanoseconds.
	timestamp := time.Now().UnixNano()

	// Construct the cursor name.
	return fmt.Sprintf("cursor_%s_%s_%d", safeConnID, shortHash, timestamp)
}

func (s *pooledInput) Read(ctx context.Context) (*service.Message, service.AckFunc, error) {
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	if s.db == nil || (s.driver == sqlmanager_shared.PostgresDriver && s.tx == nil) {
		return nil, nil, service.ErrNotConnected
	}
	if s.rows == nil {
		return nil, nil, service.ErrEndOfInput
	}
	if !s.rows.Next() {
		// Check if any error occurred.
		if err := s.rows.Err(); err != nil {
			_ = s.rows.Close()
			s.rows = nil
			return nil, nil, err
		}
		if s.driver == sqlmanager_shared.PostgresDriver && s.tx != nil {
			_ = s.rows.Close()
			// Fetch next batch for Postgres
			if err := s.fetchNextBatchFromCursor(ctx); err != nil {
				_ = s.tx.Rollback()
				return nil, nil, err
			}
			// Check if the new batch has rows
			if s.rows == nil || !s.rows.Next() {
				// No more rows; close cursor and commit transaction
				_ = s.rows.Close()
				s.rows = nil
				_ = s.tx.Commit()
				s.tx = nil
				return nil, nil, service.ErrEndOfInput
			}
		} else {
			// For non-Postgres drivers, simply close and return EndOfInput
			_ = s.rows.Close()
			s.rows = nil
			return nil, nil, service.ErrEndOfInput
		}
	}

	obj, err := s.recordMapper.MapRecord(s.rows)
	if err != nil {
		_ = s.rows.Close()
		s.rows = nil
		if s.tx != nil {
			_ = s.tx.Rollback()
			s.tx = nil
		}
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
	defer s.dbMut.Unlock()

	isNil := s.db == nil || (s.driver == sqlmanager_shared.PostgresDriver && s.tx == nil)
	if isNil {
		return nil
	}

	if s.rows != nil {
		_ = s.rows.Close()
		s.rows = nil
	}

	if s.tx != nil {
		// Rollback the transaction to clean up resources if it's still active.
		// For a read-only transaction, rollback is a safe way to end it.
		if err := s.tx.Rollback(); err != nil {
			s.logger.Error(fmt.Sprintf("error rolling back transaction on close: %s", err.Error()))
		}
		s.tx = nil
	}

	s.db = nil // not closing here since it's managed by the pool

	select {
	case <-s.shutSig.HasStoppedChan():
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
