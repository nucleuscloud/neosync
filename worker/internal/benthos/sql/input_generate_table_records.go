package neosync_benthos_sql

// combo of generate, sql select and mapping

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/benthosdev/benthos/v4/public/component/input"
	"github.com/benthosdev/benthos/v4/public/component/interop"
	"github.com/benthosdev/benthos/v4/public/service"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	"github.com/nucleuscloud/neosync/worker/internal/benthos/shutdown"
)

func generateTableRecordsInputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("driver")).
		Field(service.NewStringField("dsn")).
		Field(service.NewStringField("query")).
		Field(service.NewAnyMapField("table_columns_map")).
		Field(service.NewIntField("count")).
		Field(service.NewBloblangField("args_mapping").Optional())
}
func RegisterGenerateTableRecordsInput(env *service.Environment, dbprovider DbPoolProvider, stopActivityChannel chan error) error {
	return env.RegisterBatchInput(
		"generate_table_records", generateTableRecordsInputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
			// input, err := newSqlSelectGenerateInput(conf, mgr, dbprovider, stopActivityChannel)
			// if err != nil {
			// 	return nil, err
			// }
			// return input, nil
			nm := interop.UnwrapManagement(mgr)
			b, err := newGenerateReaderFromParsed(conf, nm)
			if err != nil {
				return nil, err
			}
			i, err := input.NewAsyncReader("generate", input.NewAsyncPreserver(b), nm)
			if err != nil {
				return nil, err
			}
			return interop.NewUnwrapInternalInput(i), nil
		},
	)
}

//------------------------------------------------------------------------------

type generateReader struct {
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

func newGenerateReaderFromParsed(conf *service.ParsedConfig, mgr *service.Resources) (*generateReader, error) {
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

	return &generateReader{
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

// Connect establishes a Bloblang reader.
func (b *generateReader) Connect(ctx context.Context) error {
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
		if !neosync_benthos.IsMaxConnectionError(err.Error()) {
			s.logger.Error(fmt.Sprintf("Benthos input error - sending stop activity signal: %s ", err.Error()))
			s.stopActivityChannel <- err
		}
		return err
	}
	jsonF, _ := json.MarshalIndent(rows, "", " ")
	fmt.Printf("\n rows: %s \n", string(jsonF))

	s.rows = rows
	go func() {
		<-s.shutSig.CloseNowChan()

		s.dbMut.Lock()
		if s.rows != nil {
			_ = s.rows.Close()
			s.rows = nil
		}
		// not closing the connection here as that is managed by an outside force
		s.db = nil
		s.dbMut.Unlock()

		s.shutSig.ShutdownComplete()
	}()
	return nil
	return nil
}

// ReadBatch a new bloblang generated message.
func (b *generateReader) ReadBatch(ctx context.Context) (message.Batch, input.AsyncAckFn, error) {
	batchSize := b.batchSize
	if b.limited {
		if b.remaining <= 0 {
			return nil, nil, component.ErrTypeClosed
		}
		if b.remaining < batchSize {
			batchSize = b.remaining
		}
	}

	if !b.firstIsFree && b.timer != nil {
		select {
		case t, open := <-b.timer.C:
			if !open {
				return nil, nil, component.ErrTypeClosed
			}
			if b.schedule != nil {
				b.timer.Reset(getDurationTillNextSchedule(t, *b.schedule, b.location))
			}
		case <-ctx.Done():
			return nil, nil, component.ErrTimeout
		}
	}
	b.firstIsFree = false

	batch := make(message.Batch, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		p, err := b.exec.MapPart(0, batch)
		if err != nil {
			return nil, nil, err
		}
		if p != nil {
			if b.limited {
				b.remaining--
			}
			batch = append(batch, p)
		}
	}
	if len(batch) == 0 {
		return nil, nil, component.ErrTimeout
	}
	return batch, func(context.Context, error) error { return nil }, nil
}

// CloseAsync shuts down the bloblang reader.
func (b *generateReader) Close(ctx context.Context) (err error) {
	if b.timer != nil {
		b.timer.Stop()
	}
	return
}
