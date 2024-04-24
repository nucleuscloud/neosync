package neosync_benthos_sql

// combo of generate, sql select and mapping

import (
	"context"
	"database/sql"
	"encoding/json"
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

func generateTableRecordsInputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("driver")).
		Field(service.NewStringField("dsn")).
		Field(service.NewAnyMapField("table_columns_map")).
		Field(service.NewStringMapField("column_name_map").Optional().Example("{ schema.table.column: new_column_name }")).
		Field(service.NewIntField("count")).
		Fields(service.NewBloblangField("mapping").Optional())
}

func RegisterGenerateTableRecordsInput(env *service.Environment, dbprovider DbPoolProvider, stopActivityChannel chan error) error {
	return env.RegisterBatchInput(
		"generate_sql_select", generateTableRecordsInputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
			b, err := newGenerateReaderFromParsed(conf, mgr, dbprovider, stopActivityChannel)
			if err != nil {
				return nil, err
			}
			return service.AutoRetryNacksBatched(b), nil
		},
	)
}

//------------------------------------------------------------------------------

type generateReader struct {
	driver        string
	dsn           string
	tableColsMap  map[string][]string
	columnNameMap map[string]string
	provider      DbPoolProvider
	logger        *service.Logger

	mapping *bloblang.Executor

	db        mysql_queries.DBTX
	dbMut     sync.Mutex
	remaining int

	shutSig *shutdown.Signaller

	stopActivityChannel chan error
}

func newGenerateReaderFromParsed(conf *service.ParsedConfig, mgr *service.Resources, dbprovider DbPoolProvider, channel chan error) (*generateReader, error) {
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

	columnNameMap := map[string]string{}
	if conf.Contains("column_name_map") {
		columnNameMap, err = conf.FieldStringMap("column_name_map")
		if err != nil {
			return nil, err
		}
	}

	var mapping *bloblang.Executor
	if conf.Contains("mapping") {
		mapping, err = conf.FieldBloblang("mapping")
		if err != nil {
			return nil, err
		}
	}

	return &generateReader{
		logger:              mgr.Logger(),
		shutSig:             shutdown.NewSignaller(),
		driver:              driver,
		dsn:                 dsn,
		tableColsMap:        tableColsMap,
		columnNameMap:       columnNameMap,
		mapping:             mapping,
		provider:            dbprovider,
		stopActivityChannel: channel,
		remaining:           count,
	}, nil
}

// Connect establishes a Bloblang reader.
func (s *generateReader) Connect(ctx context.Context) error {
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

// ReadBatch a new bloblang generated message.
func (s *generateReader) ReadBatch(ctx context.Context) (service.MessageBatch, service.AckFunc, error) {
	if s.remaining <= 0 {
		return nil, nil, service.ErrEndOfInput
	}

	sqlRandomStr := "RANDOM()"
	if s.driver == "mysql" {
		sqlRandomStr = "RAND()"
	}

	tables := []string{}
	for t := range s.tableColsMap {
		tables = append(tables, t)
	}

	randomLrgLimit, err := transformer_utils.GenerateRandomInt64InValueRange(10, 50)
	if err != nil {
		return nil, nil, err
	}

	table := tables[0]
	// need to remove self circular dependent tables
	otherTables := tables[1:]

	cols := s.tableColsMap[table]
	// update col names to be that of destination table or should it be handled on insert
	selectColumns := make([]any, len(cols))
	for i, col := range cols {
		as, ok := s.columnNameMap[fmt.Sprintf("%s.%s", table, col)]
		if ok {
			selectColumns[i] = goqu.I(col).As(as)
		} else {
			selectColumns[i] = col
		}
	}
	rows, err := s.queryDatabase(sqlRandomStr, table, randomLrgLimit, selectColumns)
	if err != nil {
		return nil, nil, err
	}

	batch := service.MessageBatch{}
	for _, r := range rows {
		randomSmLimit, err := transformer_utils.GenerateRandomInt64InValueRange(0, 3)
		if err != nil {
			return nil, nil, err
		}
		otherTableRows := [][]map[string]any{}
		for _, t := range otherTables {
			cols := s.tableColsMap[t]
			selectColumns := make([]any, len(cols))
			for i, col := range cols {
				tn := fmt.Sprintf("%s.%s", t, col)
				as, ok := s.columnNameMap[tn]
				if ok {
					selectColumns[i] = goqu.I(col).As(as)
				} else {
					selectColumns[i] = col
				}
			}
			newRows, err := s.queryDatabase(sqlRandomStr, t, randomSmLimit, selectColumns)
			if err != nil {
				return nil, nil, err
			}
			if len(newRows) != 0 {
				// how to handle tables that don't have enough data
				otherTableRows = append(otherTableRows, newRows)
			}
		}
		combinedRows := combineRowLists(otherTableRows)
		if len(combinedRows) > 0 {
			for _, cr := range combinedRows {
				var args map[string]any
				if s.mapping != nil {
					args, err = s.queryBloblangMapping()
					if err != nil {
						return nil, nil, err
					}
				}

				newRow := combineRows([]map[string]any{r, cr, args})
				rowStr, err := json.Marshal(newRow)
				if err != nil {
					return nil, nil, err
				}
				if s.remaining < 1 {
					return batch, func(context.Context, error) error { return nil }, nil
				}

				msg := service.NewMessage(rowStr)
				batch = append(batch, msg)
				s.remaining--
			}
		} else {
			newRow := r
			if s.mapping != nil {
				args, err := s.queryBloblangMapping()
				if err != nil {
					return nil, nil, err
				}
				newRow = combineRows([]map[string]any{r, args})
			}

			rowStr, err := json.Marshal(newRow)
			if err != nil {
				return nil, nil, err
			}
			if s.remaining < 1 {
				return batch, func(context.Context, error) error { return nil }, nil
			}
			msg := service.NewMessage(rowStr)
			batch = append(batch, msg)
			s.remaining--
		}
	}

	return batch, func(context.Context, error) error { return nil }, nil
}

func (s *generateReader) Close(ctx context.Context) (err error) {
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

func (s *generateReader) queryBloblangMapping() (map[string]any, error) {
	var iargs any
	var err error
	if iargs, err = s.mapping.Query(nil); err != nil {
		return nil, err
	}

	var ok bool
	var args map[string]any
	if args, ok = iargs.(map[string]any); !ok {
		err = fmt.Errorf("mapping returned non-array result: %T", iargs)
		return nil, err
	}
	return args, nil
}

func (s *generateReader) queryDatabase(sqlRandomStr, table string, limit int64, columns []any) ([]map[string]any, error) {
	orderBy := exp.NewOrderedExpression(exp.NewLiteralExpression(sqlRandomStr), exp.AscDir, exp.NullsLastSortType)
	builder := goqu.Dialect(s.driver)
	selectBuilder := builder.From(table).Select(columns...).Order(orderBy).Limit(uint(limit))
	selectSql, _, err := selectBuilder.ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, selectSql)
	if err != nil {
		if !neosync_benthos.ShouldTerminate(err.Error()) {
			s.logger.Error(fmt.Sprintf("Benthos input error - sending stop activity signal: %s ", err.Error()))
			s.stopActivityChannel <- err
		}
		return nil, err
	}

	rowObjList, err := sqlRowsToMapList(rows)
	if err != nil {
		_ = rows.Close()
		return nil, err
	}
	return rowObjList, nil
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
	if len(rows) == 0 {
		return []map[string]any{}
	}
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
