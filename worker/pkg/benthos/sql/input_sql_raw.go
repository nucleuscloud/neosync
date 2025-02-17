package neosync_benthos_sql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	continuation_token "github.com/nucleuscloud/neosync/internal/continuation-token"
	database_record_mapper "github.com/nucleuscloud/neosync/internal/database-record-mapper"
	record_mapper_builder "github.com/nucleuscloud/neosync/internal/database-record-mapper/builder"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/redpanda-data/benthos/v4/public/service"
)

func sqlRawInputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("connection_id")).
		Field(service.NewStringField("query")).
		Field(service.NewStringField("paged_query").Optional()).
		Field(service.NewIntField("expected_total_rows").Optional()).
		Field(service.NewStringListField("order_by_columns").Default([]string{}))
}

// Registers an input on a benthos environment called pooled_sql_raw
func RegisterPooledSqlRawInput(
	env *service.Environment,
	dbprovider ConnectionProvider,
	stopActivityChannel chan<- error,
	onHasMorePages OnHasMorePagesFn,
	continuationToken *continuation_token.ContinuationToken,
) error {
	return env.RegisterInput(
		"pooled_sql_raw",
		sqlRawInputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Input, error) {
			input, err := newInput(conf, mgr, dbprovider, stopActivityChannel, onHasMorePages, continuationToken)
			if err != nil {
				return nil, err
			}
			// return service.AutoRetryNacksToggled(conf, input)
			return input, nil
		},
	)
}

type OnHasMorePagesFn func(lastReadOrderValues []any)

type pooledInput struct {
	provider     ConnectionProvider
	logger       *service.Logger
	connectionId string
	driver       string

	queryStatic      string
	pagedQueryStatic *string
	db               mysql_queries.DBTX
	dbMut            sync.Mutex

	rows *sql.Rows

	recordMapper record_mapper_builder.DatabaseRecordMapper[any]

	stopActivityChannel chan<- error
	onHasMorePages      OnHasMorePagesFn
	expectedTotalRows   *int
	rowsRead            int
	orderByColumns      []string
	lastReadOrderValues []any
	continuationToken   *continuation_token.ContinuationToken
}

func newInput(
	conf *service.ParsedConfig,
	mgr *service.Resources,
	dbprovider ConnectionProvider,
	channel chan<- error,
	onHasMorePages OnHasMorePagesFn,
	continuationToken *continuation_token.ContinuationToken,
) (*pooledInput, error) {
	connectionId, err := conf.FieldString("connection_id")
	if err != nil {
		return nil, err
	}

	queryStatic, err := conf.FieldString("query")
	if err != nil {
		return nil, err
	}

	var pagedQueryStatic *string
	if conf.Contains("paged_query") {
		pquery, err := conf.FieldString("paged_query")
		if err != nil {
			return nil, err
		}
		pagedQueryStatic = &pquery
	}

	var expectedTotalRows *int
	if conf.Contains("expected_total_rows") {
		totalRows, err := conf.FieldInt("expected_total_rows")
		if err != nil {
			return nil, err
		}
		expectedTotalRows = &totalRows
	}

	orderByColumns, err := conf.FieldStringList("order_by_columns")
	if err != nil {
		return nil, err
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
		connectionId:        connectionId,
		driver:              driver,
		queryStatic:         queryStatic,
		pagedQueryStatic:    pagedQueryStatic,
		provider:            dbprovider,
		stopActivityChannel: channel,
		recordMapper:        mapper,
		onHasMorePages:      onHasMorePages,
		expectedTotalRows:   expectedTotalRows,
		orderByColumns:      orderByColumns,
		lastReadOrderValues: []any{},
		continuationToken:   continuationToken,
	}, nil
}

var _ service.Input = &pooledInput{}

func (s *pooledInput) Connect(ctx context.Context) error {
	s.logger.Debug("connecting to pooled input")
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	db, err := s.provider.GetDb(ctx, s.connectionId)
	if err != nil {
		return err
	}
	s.db = db
	s.logger.Debug(fmt.Sprintf("connected to database %s", s.connectionId))

	query := s.queryStatic
	var args []any
	if s.pagedQueryStatic != nil && s.continuationToken != nil && s.expectedTotalRows != nil {
		if len(s.orderByColumns) != len(s.continuationToken.Contents.LastReadOrderValues) {
			columnMisMatchErr := fmt.Errorf("order by columns and last read order values must be the same length")
			s.logger.Error(columnMisMatchErr.Error())
			s.stopActivityChannel <- columnMisMatchErr
			return columnMisMatchErr
		}
		s.logger.Debug("using paged query")
		query = *s.pagedQueryStatic

		// Build arguments for lexicographical ordering
		// To retain this ordering, we need to build the args in a way that match the OR'd statements from the prepared query
		// For lexi ordering the total number of args follows the algorithm: (n (n + 1)) / 2 where n is the number of order by columns
		// For example, if we have 2 order by columns, we need 3 args:
		// First OR condition: [id_value]
		// Second OR condition: [id_value, name_value]
		// Third OR condition: [id_value, name_value, email_value]
		// And so on...
		lastValues := s.continuationToken.Contents.LastReadOrderValues
		for i := 0; i < len(s.orderByColumns); i++ {
			// For each OR condition, add values up to and including current position
			for j := 0; j <= i; j++ {
				args = append(args, lastValues[j])
			}
		}
		args = append(args, *s.expectedTotalRows)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		if neosync_benthos.IsCriticalError(err.Error()) {
			s.logger.Error(fmt.Sprintf("Benthos input error - sending stop activity signal: %s ", err.Error()))
			s.stopActivityChannel <- err
		}
		return err
	}

	s.rows = rows
	s.rowsRead = 0
	return nil
}

func (s *pooledInput) Read(ctx context.Context) (*service.Message, service.AckFunc, error) {
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	if s.db == nil {
		return nil, nil, service.ErrNotConnected
	}
	if s.rows == nil {
		if s.expectedTotalRows != nil && s.onHasMorePages != nil && len(s.orderByColumns) > 0 {
			// emit order by column values if ok
			s.logger.Debug(fmt.Sprintf("rows read: %d, expected total rows: %d", s.rowsRead, *s.expectedTotalRows))
			if s.rowsRead >= *s.expectedTotalRows {
				s.logger.Debug("emitting order by column values")
				s.onHasMorePages(s.lastReadOrderValues)
			}
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
		if s.expectedTotalRows != nil && s.onHasMorePages != nil && len(s.orderByColumns) > 0 {
			// emit order by column values if ok
			s.logger.Debug(fmt.Sprintf("[ROW END] rows read: %d, expected total rows: %d", s.rowsRead, *s.expectedTotalRows))
			if s.rowsRead >= *s.expectedTotalRows {
				s.logger.Debug("[ROW END] emitting order by column values")
				s.onHasMorePages(s.lastReadOrderValues)
			}
		}
		return nil, nil, service.ErrEndOfInput
	}

	obj, err := s.recordMapper.MapRecord(s.rows)
	if err != nil {
		_ = s.rows.Close()
		s.rows = nil
		return nil, nil, err
	}

	// store last order by columns values
	lastReadOrderValues := make([]any, len(s.orderByColumns))
	for i, col := range s.orderByColumns {
		val, ok := obj[col]
		if !ok {
			_ = s.rows.Close()
			s.rows = nil
			return nil, nil, fmt.Errorf("order by column %s not found", col)
		}
		lastReadOrderValues[i] = val
	}
	if len(lastReadOrderValues) > 0 {
		s.logger.Debug(fmt.Sprintf("last read order values: %v", lastReadOrderValues))
		s.lastReadOrderValues = lastReadOrderValues
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
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	if s.rows != nil {
		_ = s.rows.Close()
		s.rows = nil
	}

	if s.db != nil {
		s.db = nil // not closing here since it's managed by the pool
	}
	return nil
}
