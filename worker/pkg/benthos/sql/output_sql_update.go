package neosync_benthos_sql

import (
	"context"
	"fmt"
	"sync"

	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/nucleuscloud/neosync/backend/pkg/sqldbtx"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	querybuilder "github.com/nucleuscloud/neosync/worker/pkg/query-builder"
	"github.com/redpanda-data/benthos/v4/public/service"
)

type SqlDbtx interface {
	sqldbtx.DBTX
	Close() error
}

type DbPoolProvider interface {
	GetDb(driver, dsn string) (SqlDbtx, error)
}

type ConnectionProvider interface {
	GetDb(ctx context.Context, connectionId string) (SqlDbtx, error)
	GetDriver(connectionId string) (string, error)
}

func sqlUpdateOutputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("connection_id")).
		Field(service.NewStringField("schema")).
		Field(service.NewStringField("table")).
		Field(service.NewStringListField("columns")).
		Field(service.NewStringListField("where_columns")).
		Field(service.NewBoolField("skip_foreign_key_violations").Optional().Default(false)).
		Field(service.NewIntField("max_in_flight").Default(64)).
		Field(service.NewBatchPolicyField("batching"))
}

// Registers an output on a benthos environment called pooled_sql_raw
func RegisterPooledSqlUpdateOutput(env *service.Environment, dbprovider ConnectionProvider) error {
	return env.RegisterBatchOutput(
		"pooled_sql_update", sqlUpdateOutputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchOutput, service.BatchPolicy, int, error) {
			batchPolicy, err := conf.FieldBatchPolicy("batching")
			if err != nil {
				return nil, batchPolicy, -1, err
			}

			maxInFlight, err := conf.FieldInt("max_in_flight")
			if err != nil {
				return nil, service.BatchPolicy{}, -1, err
			}
			out, err := newUpdateOutput(conf, mgr, dbprovider)
			if err != nil {
				return nil, service.BatchPolicy{}, -1, err
			}
			return out, batchPolicy, maxInFlight, nil
		},
	)
}

var _ service.BatchOutput = &pooledUpdateOutput{}

type pooledUpdateOutput struct {
	driver       string
	connectionId string
	provider     ConnectionProvider
	dbMut        sync.RWMutex
	db           SqlDbtx
	logger       *service.Logger

	schema                   string
	table                    string
	columns                  []string
	whereCols                []string
	skipForeignKeyViolations bool
}

func newUpdateOutput(conf *service.ParsedConfig, mgr *service.Resources, provider ConnectionProvider) (*pooledUpdateOutput, error) {
	connectionId, err := conf.FieldString("connection_id")
	if err != nil {
		return nil, err
	}

	schema, err := conf.FieldString("schema")
	if err != nil {
		return nil, err
	}

	table, err := conf.FieldString("table")
	if err != nil {
		return nil, err
	}

	columns, err := conf.FieldStringList("columns")
	if err != nil {
		return nil, err
	}

	whereCols, err := conf.FieldStringList("where_columns")
	if err != nil {
		return nil, err
	}

	skipForeignKeyViolations, err := conf.FieldBool("skip_foreign_key_violations")
	if err != nil {
		return nil, err
	}

	driver, err := provider.GetDriver(connectionId)
	if err != nil {
		return nil, err
	}

	output := &pooledUpdateOutput{
		driver:                   driver,
		connectionId:             connectionId,
		logger:                   mgr.Logger(),
		provider:                 provider,
		schema:                   schema,
		table:                    table,
		columns:                  columns,
		whereCols:                whereCols,
		skipForeignKeyViolations: skipForeignKeyViolations,
	}
	return output, nil
}

func (s *pooledUpdateOutput) Connect(ctx context.Context) error {
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	if s.db != nil {
		return nil
	}

	db, err := s.provider.GetDb(ctx, s.connectionId)
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *pooledUpdateOutput) WriteBatch(ctx context.Context, batch service.MessageBatch) error {
	s.dbMut.RLock()
	if s.db == nil {
		s.dbMut.RUnlock()
		return service.ErrNotConnected
	}
	db := s.db
	defer s.dbMut.RUnlock()

	batchLen := len(batch)
	if batchLen == 0 {
		return nil
	}

	for _, msg := range batch {
		m, _ := msg.AsStructured()

		// msgMap has all the table columns and values not just the columns we are updating
		msgMap, ok := m.(map[string]any)
		if !ok {
			return fmt.Errorf("message returned non-map result: %T", msgMap)
		}

		query, err := querybuilder.BuildUpdateQuery(s.driver, s.schema, s.table, s.columns, s.whereCols, msgMap)
		if err != nil {
			return err
		}

		if _, err := db.ExecContext(ctx, query); err != nil {
			if !s.skipForeignKeyViolations || !neosync_benthos.IsForeignKeyViolationError(err.Error()) {
				return err
			}
		}
	}

	return nil
}

func (s *pooledUpdateOutput) Close(ctx context.Context) error {
	s.dbMut.RLock()
	if s.db == nil {
		s.dbMut.RUnlock()
		return nil
	}
	s.dbMut.RUnlock()

	// Take write lock to null out the connection
	s.dbMut.Lock()
	s.db = nil
	s.dbMut.Unlock()
	return nil
}
