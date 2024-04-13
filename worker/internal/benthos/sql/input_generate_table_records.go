package neosync_benthos_sql

// combo of generate, sql select and mapping

import (
	"context"
	"sync"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/benthosdev/benthos/v4/public/component/input"
	"github.com/benthosdev/benthos/v4/public/component/interop"
	"github.com/benthosdev/benthos/v4/public/service"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
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
			b, err := newGenerateReaderFromParsed(conf, nm, dbprovider, stopActivityChannel)
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
func (b *generateReader) ReadBatch(ctx context.Context) (service.MessageBatch, input.AsyncAckFn, error) {
	if b.remaining <= 0 {
		return nil, nil, service.ErrEndOfInput
	}

	// b.firstIsFree = false

	batch := make(service.MessageBatch, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		p, err := b.exec.MapPart(0, batch)
		if err != nil {
			return nil, nil, err
		}
		if p != nil {
			b.remaining--
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
