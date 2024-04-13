package neosync_benthos_sql

// combo of generate, sql select and mapping

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	"github.com/nucleuscloud/neosync/worker/internal/benthos/shutdown"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

func sqlSelectGenerateInputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("driver")).
		Field(service.NewStringField("dsn")).
		Field(service.NewStringField("query")).
		Field(service.NewAnyMapField("table_columns_map")).
		Field(service.NewIntField("count")).
		Field(service.NewBloblangField("args_mapping").Optional())
}

// Registers an input on a benthos environment called pooled_sql_raw
func RegisterSqlSelectGenerateInput(env *service.Environment, dbprovider DbPoolProvider, stopActivityChannel chan error) error {
	return env.RegisterInput(
		"pooled_sql_select_generate", sqlSelectGenerateInputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Input, error) {
			input, err := newSqlSelectGenerateInput(conf, mgr, dbprovider, stopActivityChannel)
			if err != nil {
				return nil, err
			}
			return input, nil
		},
	)
}

type sqlSelectGenerateInput struct {
	driver       string
	dsn          string
	tableColsMap map[string][]string
	provider     DbPoolProvider
	logger       *service.Logger

	argsMapping *bloblang.Executor

	db    mysql_queries.DBTX
	dbMut sync.Mutex
	// rows      *sql.Rows
	remaining  int
	index      int
	joinedRows []map[string]any

	shutSig *shutdown.Signaller

	stopActivityChannel chan error
}

func newSqlSelectGenerateInput(conf *service.ParsedConfig, mgr *service.Resources, dbprovider DbPoolProvider, channel chan error) (*sqlSelectGenerateInput, error) {
	driver, err := conf.FieldString("driver")
	if err != nil {
		return nil, err
	}
	dsn, err := conf.FieldString("dsn")
	if err != nil {
		return nil, err
	}

	count, err := conf.FieldInt("count")
	if err != nil {
		return nil, err
	}

	tmpMap, err := conf.FieldAnyMap("table_columns_map")
	if err != nil {
		return nil, err
	}
	tableColsMap := map[string][]string{}
	for k, v := range tmpMap {
		val, err := v.FieldStringList()
		if err != nil {
			return nil, err
		}
		tableColsMap[k] = val
	}
	var argsMapping *bloblang.Executor
	if conf.Contains("args_mapping") {
		argsMapping, err = conf.FieldBloblang("args_mapping")
		if err != nil {
			return nil, err
		}
	}

	return &sqlSelectGenerateInput{
		logger:              mgr.Logger(),
		shutSig:             shutdown.NewSignaller(),
		driver:              driver,
		dsn:                 dsn,
		tableColsMap:        tableColsMap,
		argsMapping:         argsMapping,
		provider:            dbprovider,
		stopActivityChannel: channel,
		remaining:           count,
		index:               0,
	}, nil
}

var _ service.Input = &pooledInput{}

func (s *sqlSelectGenerateInput) Connect(ctx context.Context) error {
	s.logger.Debug("connecting to pooled input")
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	if s.db != nil {
		return nil
	}

	seed, err := transformer_utils.GenerateCryptoSeed()
	if err != nil {
		return err
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

	sqlRandomStr := "RANDOM()"
	if s.driver == "mysql" {
		sqlRandomStr = "RAND()"
	}

	tables := []string{}
	for t := range s.tableColsMap {
		tables = append(tables, t)
	}

	randomLimit, err := transformer_utils.GenerateRandomInt64InValueRange(10, 100)
	if err != nil {
		return err
	}

	joinedRows := []map[string]any{}

	table := tables[0]
	otherTables := tables[1:]

	cols := s.tableColsMap[table]
	// update col names to be that of destination table or should it be handled on insert
	selectColumns := make([]any, len(cols))
	for i, col := range cols {
		selectColumns[i] = col
	}
	orderBy := exp.NewOrderedExpression(exp.NewLiteralExpression(sqlRandomStr), exp.AscDir, exp.NullsLastSortType)
	builder := goqu.Dialect(s.driver)
	selectBuilder := builder.From(table).Select(selectColumns...).Order(orderBy).Limit(uint(randomLimit))
	selectSql, _, err := selectBuilder.ToSQL()
	if err != nil {
		return err
	}

	// select rows from

	rows, err := db.QueryContext(ctx, selectSql)
	if err != nil {
		if !neosync_benthos.IsMaxConnectionError(err.Error()) {
			s.logger.Error(fmt.Sprintf("Benthos input error - sending stop activity signal: %s ", err.Error()))
			s.stopActivityChannel <- err
		}
		return err
	}

	rowObjList, err := sqlRowsToMapList(rows)
	if err != nil {
		_ = rows.Close()
		return err
	}

	for _, r := range rowObjList {
		randLimit, err := transformer_utils.GenerateRandomInt64InValueRange(0, 3)
		if err != nil {
			return err
		}
		otherTableRows := [][]map[string]any{}
		for _, t := range otherTables {
			cols := s.tableColsMap[t]
			selectColumns := make([]any, len(cols))
			for i, col := range cols {
				selectColumns[i] = col
			}
			selectBuilder := builder.From(table).Select(selectColumns...).Order(orderBy).Limit(uint(randLimit))
			selectSql, _, err := selectBuilder.ToSQL()
			if err != nil {
				return err
			}
			rows, err := db.QueryContext(ctx, selectSql)
			if err != nil {
				if !neosync_benthos.IsMaxConnectionError(err.Error()) {
					s.logger.Error(fmt.Sprintf("Benthos input error - sending stop activity signal: %s ", err.Error()))
					s.stopActivityChannel <- err
				}
				return err
			}
			rowObjList, err := sqlRowsToMapList(rows)
			if err != nil {
				_ = rows.Close()
				return err
			}
			otherTableRows = append(otherTableRows, rowObjList)
		}

	}

	s.joinedRows = joinedRows
	go func() {
		<-s.shutSig.CloseNowChan()

		s.dbMut.Lock()
		if rows != nil {
			_ = rows.Close()
			rows = nil
		}
		// not closing the connection here as that is managed by an outside force
		s.db = nil
		s.dbMut.Unlock()

		s.shutSig.ShutdownComplete()
	}()
	return nil
}

func (s *sqlSelectGenerateInput) Read(ctx context.Context) (*service.Message, service.AckFunc, error) {
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

func (s *sqlSelectGenerateInput) Close(ctx context.Context) error {
	s.shutSig.CloseNow()
	s.dbMut.Lock()
	isNil := s.db == nil
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

func sqlRowsToMapList(rows *sql.Rows) ([]map[string]any, error) {
	results := []map[string]any{}
	for rows.Next() {
		obj, err := sqlRowToMap(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, obj)
	}
	return results, nil
}

func combineRows(maps []map[string]any) map[string]any {
	result := make(map[string]any)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func combineRowLists(rows [][]map[string]any) []map[string]any {
	results := []map[string]any{}
	rowCount := len(rows[0])
	for i := 0; i < rowCount; i++ {
		rowsToCombine := []map[string]any{}
		for _, r := range rows {
			rowsToCombine = append(rowsToCombine, r[i])
		}
		results = append(results, combineRows(rowsToCombine))
	}
	return results
}
