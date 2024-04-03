package neosync_benthos_sql

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/benthosdev/benthos/v4/public/service"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	"github.com/nucleuscloud/neosync/worker/internal/benthos/shutdown"
)

func sqlRawOutputSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		Field(service.NewStringField("driver")).
		Field(service.NewStringField("dsn")).
		Field(service.NewStringField("query")).
		Field(service.NewBoolField("unsafe_dynamic_query").Default(false)).
		Field(service.NewBloblangField("args_mapping").Optional()).
		Field(service.NewIntField("max_in_flight").Default(64)).
		Field(service.NewBatchPolicyField("batching")).
		Field(service.NewStringField("init_statement").Optional())
}

// type DbPoolProvider interface {
// 	GetDb(driver, dsn string) (mysql_queries.DBTX, error)
// }

// Registers an output on a benthos environment called pooled_sql_raw
func RegisterPooledSqlRawOutput(env *service.Environment, dbprovider DbPoolProvider) error {
	return env.RegisterBatchOutput(
		"pooled_sql_insert", sqlRawOutputSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchOutput, service.BatchPolicy, int, error) {
			batchPolicy, err := conf.FieldBatchPolicy("batching")
			if err != nil {
				return nil, batchPolicy, -1, err
			}

			maxInFlight, err := conf.FieldInt("max_in_flight")
			if err != nil {
				return nil, service.BatchPolicy{}, -1, err
			}
			out, err := newOutput(conf, mgr, dbprovider)
			if err != nil {
				return nil, service.BatchPolicy{}, -1, err
			}
			return out, batchPolicy, maxInFlight, nil
		},
	)
}

var _ service.BatchOutput = &pooledInsertOutput{}

type pooledInsertOutput struct {
	driver   string
	dsn      string
	provider DbPoolProvider
	dbMut    sync.RWMutex
	db       mysql_queries.DBTX
	logger   *service.Logger

	queryStatic string
	queryDyn    *service.InterpolatedString

	argsMapping *bloblang.Executor
	shutSig     *shutdown.Signaller
}

func newInsertOutput(conf *service.ParsedConfig, mgr *service.Resources, provider DbPoolProvider) (*pooledInsertOutput, error) {
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

	var queryDyn *service.InterpolatedString
	if unsafeDyn, err := conf.FieldBool("unsafe_dynamic_query"); err != nil {
		return nil, err
	} else if unsafeDyn {
		if queryDyn, err = conf.FieldInterpolatedString("query"); err != nil {
			return nil, err
		}
	}

	var argsMapping *bloblang.Executor
	if conf.Contains("args_mapping") {
		if argsMapping, err = conf.FieldBloblang("args_mapping"); err != nil {
			return nil, err
		}
	}

	output := &pooledInsertOutput{
		driver:      driver,
		dsn:         dsn,
		logger:      mgr.Logger(),
		shutSig:     shutdown.NewSignaller(),
		queryStatic: queryStatic,
		queryDyn:    queryDyn,
		argsMapping: argsMapping,
		provider:    provider,
	}
	return output, nil
}

func (s *pooledInsertOutput) Connect(ctx context.Context) error {
	s.dbMut.Lock()
	defer s.dbMut.Unlock()

	if s.db != nil {
		return nil
	}

	db, err := s.provider.GetDb(s.driver, s.dsn)
	if err != nil {
		return err
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

func (s *pooledInsertOutput) WriteBatch(ctx context.Context, batch service.MessageBatch) error {
	s.dbMut.RLock()
	defer s.dbMut.RUnlock()

	batchLen := len(batch)
	if batchLen == 0 {
		return nil
	}
	fmt.Println()
	fmt.Println(batchLen)
	batchedArgs := []any{}
	queryStr := s.queryStatic
	if s.queryDyn != nil {
		var err error
		if queryStr, err = batch.TryInterpolatedString(0, s.queryDyn); err != nil {
			return fmt.Errorf("query interpolation error: %w", err)
		}
	}
	query := extendQueryParameters(queryStr, batchLen)
	for i := range batch {
		// var args []any
		if s.argsMapping == nil {
			continue
		}
		// if s.argsMapping != nil {
		resMsg, err := batch.BloblangQuery(i, s.argsMapping)
		if err != nil {
			return err
		}

		iargs, err := resMsg.AsStructured()
		if err != nil {
			return err
		}

		args, ok := iargs.([]any)
		if !ok {
			return fmt.Errorf("mapping returned non-array result: %T", iargs)
		}
		batchedArgs = append(batchedArgs, args...)
		// }

		// queryStr := s.queryStatic
		// if s.queryDyn != nil {
		// 	var err error
		// 	if queryStr, err = batch.TryInterpolatedString(i, s.queryDyn); err != nil {
		// 		return fmt.Errorf("query interpolation error: %w", err)
		// 	}
		// }
		// handle postgres and mysql

		// if _, err := s.db.ExecContext(ctx, queryStr, args...); err != nil {
		// 	fmt.Println()
		// 	fmt.Println(queryStr)
		// 	fmt.Printf("%+v\n", args)
		// 	fmt.Println()
		// 	return err
		// }
	}
	fmt.Println()
	fmt.Println(query)
	fmt.Printf("%+v\n", batchedArgs)
	fmt.Println()
	fmt.Println()
	if _, err := s.db.ExecContext(ctx, query, batchedArgs...); err != nil {
		return err
	}
	return nil
}

func (s *pooledInsertOutput) Close(ctx context.Context) error {
	s.shutSig.CloseNow()
	s.dbMut.RLock()
	isNil := s.db == nil
	s.dbMut.RUnlock()
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

// func extendQueryParameters(query string, batchLen int) string {
// 	// Find the starting position of the "VALUES" section.
// 	valuesPos := strings.Index(query, "VALUES")
// 	if valuesPos == -1 {
// 		// If "VALUES" is not found, return the original query as it might not be a valid INSERT statement.
// 		return query
// 	}

// 	// Extract the section of the query that contains the parameters.
// 	paramSectionStart := strings.Index(query[valuesPos:], "(") + valuesPos
// 	paramSectionEnd := strings.Index(query[valuesPos:], ")") + valuesPos
// 	if paramSectionStart == -1 || paramSectionEnd == -1 || paramSectionStart >= paramSectionEnd {
// 		// If the parentheses are not found or are in the wrong order, return the original query.
// 		return query
// 	}
// 	paramSection := query[paramSectionStart+1 : paramSectionEnd]

// 	// Split the existing parameters by comma and count them.
// 	existingParams := strings.Split(paramSection, ",")
// 	existingParamCount := len(existingParams)
// 	totalParamCount := batchLen * existingParamCount

// 	// If the existing number of parameters is already equal to or greater than the required count, return the original query.
// 	if existingParamCount >= totalParamCount {
// 		return query
// 	}

// 	// Generate additional parameters and append them to the existing ones.
// 	for i := existingParamCount + 1; i <= totalParamCount; i++ {
// 		if i > existingParamCount {
// 			paramSection += ","
// 		}
// 		paramSection += fmt.Sprintf("$%d", i)
// 	}

//		// Reconstruct the query with the extended parameter section.
//		return query[:paramSectionStart+1] + paramSection + query[paramSectionEnd:]
//	}
func extendQueryParameters(query string, batchLen int) string {
	// Find the starting position of the "VALUES" section.
	valuesPos := strings.Index(query, "VALUES")
	if valuesPos == -1 {
		return query // Not a valid INSERT statement for this operation.
	}

	// Extract the section of the query that contains the parameters.
	paramSectionStart := strings.Index(query[valuesPos:], "(") + valuesPos
	paramSectionEnd := strings.LastIndex(query, ")")
	if paramSectionStart == -1 || paramSectionEnd == -1 || paramSectionStart >= paramSectionEnd {
		return query // Invalid or unexpected query format.
	}

	existingParamSection := query[paramSectionStart+1 : paramSectionEnd]
	existingParams := strings.Split(existingParamSection, ",")

	// Determine the number of columns per row by counting the commas in the first row of parameters.
	columnCount := len(existingParams)

	// Calculate how many additional parameters and rows are needed.
	existingRowCount := (len(existingParams) + columnCount - 1) / columnCount
	// requiredRowCount := (paramCount + columnCount - 1) / columnCount

	// Build the new parameter section by repeating the pattern.
	var newParams []string
	for i := existingRowCount * columnCount; i < batchLen*columnCount; i++ {
		newParams = append(newParams, fmt.Sprintf("$%d", i+1))
	}

	// Split the new parameters into rows and append them to the existing parameter section.
	for i := 0; i < len(newParams); i += columnCount {
		end := i + columnCount
		if end > len(newParams) {
			end = len(newParams)
		}
		rowParams := strings.Join(newParams[i:end], ",")
		existingParamSection += "), (" + rowParams
	}

	// Reconstruct the query with the new parameter section.
	newQuery := query[:paramSectionStart+1] + existingParamSection + query[paramSectionEnd:]
	return newQuery
}
