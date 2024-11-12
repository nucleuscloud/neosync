package ee_sqlmanager_mssql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"golang.org/x/sync/errgroup"
)

type Manager struct {
	querier mssql_queries.Querier
	db      mysql_queries.DBTX
	close   func()
}

func NewManager(querier mssql_queries.Querier, db mysql_queries.DBTX, closer func()) *Manager {
	return &Manager{querier: querier, db: db, close: closer}
}

func (m *Manager) GetTableInitStatements(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableInitStatement, error) {
	if len(tables) == 0 {
		return []*sqlmanager_shared.TableInitStatement{}, nil
	}

	combined := []string{}
	schemaset := map[string]struct{}{}
	for _, table := range tables {
		combined = append(combined, table.String())
		schemaset[table.Schema] = struct{}{}
	}
	schemas := []string{}
	for schema := range schemaset {
		schemas = append(schemas, schema)
	}

	errgrp, errctx := errgroup.WithContext(ctx)

	colDefMap := map[string][]*mssql_queries.GetDatabaseTableSchemasBySchemasAndTablesRow{}
	errgrp.Go(func() error {
		columnDefs, err := m.querier.GetDatabaseTableSchemasBySchemasAndTables(errctx, m.db, combined)
		if err != nil {
			return err
		}
		for _, columnDefinition := range columnDefs {
			key := sqlmanager_shared.SchemaTable{Schema: columnDefinition.TableSchema, Table: columnDefinition.TableName}
			colDefMap[key.String()] = append(colDefMap[key.String()], columnDefinition)
		}
		return nil
	})

	constraintmap := map[string][]*mssql_queries.GetTableConstraintsBySchemasRow{}
	errgrp.Go(func() error {
		constraints, err := m.querier.GetTableConstraintsBySchemas(errctx, m.db, schemas) // todo: update this to only grab what is necessary instead of entire schema
		if err != nil {
			return err
		}
		for _, constraint := range constraints {
			key := sqlmanager_shared.SchemaTable{Schema: constraint.SchemaName, Table: constraint.TableName}
			constraintmap[key.String()] = append(constraintmap[key.String()], constraint)
		}
		return nil
	})

	indexmap := map[string][]string{}
	errgrp.Go(func() error {
		idxrecords, err := m.querier.GetIndicesBySchemasAndTables(errctx, m.db, combined)
		if err != nil {
			return err
		}
		for _, record := range idxrecords {
			key := sqlmanager_shared.SchemaTable{Schema: record.SchemaName, Table: record.TableName}
			indexmap[key.String()] = append(indexmap[key.String()], generateCreateIndexStatement(record))
		}
		return nil
	})

	if err := errgrp.Wait(); err != nil {
		return nil, err
	}

	output := []*sqlmanager_shared.TableInitStatement{}
	// using input here causes the output to always be consistent
	for _, schematable := range tables {
		key := schematable.String()
		tableData, ok := colDefMap[key]
		if !ok {
			continue
		}

		info := &sqlmanager_shared.TableInitStatement{
			CreateTableStatement: generateCreateTableStatement(tableData),
			AlterTableStatements: []*sqlmanager_shared.AlterTableStatement{},
			IndexStatements:      indexmap[key],
		}
		for _, constraint := range constraintmap[key] {
			if constraint.ConstraintType == "PRIMARY KEY" {
				// primary keys must be defined in create table statement
				continue
			}
			stmt := generateAddConstraintStatement(constraint)
			constraintType, err := sqlmanager_shared.ToConstraintType(toStandardConstraintType(constraint.ConstraintType))
			if err != nil {
				return nil, err
			}
			info.AlterTableStatements = append(info.AlterTableStatements, &sqlmanager_shared.AlterTableStatement{
				Statement:      stmt,
				ConstraintType: constraintType,
			})
		}
		output = append(output, info)
	}
	return output, nil
}

func toStandardConstraintType(constraintType string) string {
	switch constraintType {
	case "PRIMARY KEY":
		return "p"
	case "UNIQUE":
		return "u"
	case "FOREIGN KEY":
		return "f"
	case "CHECK":
		return "c"
	default:
		return ""
	}
}

func (m *Manager) GetSchemaInitStatements(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.InitSchemaStatements, error) {
	schemasMap := map[string]struct{}{}
	for _, t := range tables {
		schemasMap[t.Schema] = struct{}{}
	}
	schemas := []string{}
	for schema := range schemasMap {
		schemas = append(schemas, schema)
	}

	errgrp, errctx := errgroup.WithContext(ctx)
	dataTypeStmts := []string{}
	errgrp.Go(func() error {
		datatypeCfg, err := m.GetSchemaTableDataTypes(errctx, tables)
		if err != nil {
			return fmt.Errorf("unable to retrieve mssql schema table data types: %w", err)
		}
		dataTypeStmts = datatypeCfg.GetStatements()
		return nil
	})

	tableTriggerStmts := []string{}
	errgrp.Go(func() error {
		tableTriggers, err := m.GetSchemaTableTriggers(ctx, tables)
		if err != nil {
			return fmt.Errorf("unable to retrieve mssql schema table triggers: %w", err)
		}
		for _, ttrig := range tableTriggers {
			tableTriggerStmts = append(tableTriggerStmts, ttrig.Definition)
		}
		return nil
	})

	createTables := []string{}
	nonFkAlterStmts := []string{}
	fkAlterStmts := []string{}
	idxStmts := []string{}
	errgrp.Go(func() error {
		initStatementCfgs, err := m.GetTableInitStatements(ctx, tables)
		if err != nil {
			return fmt.Errorf("unable to retrieve mssql schema table create statements: %w", err)
		}
		for _, stmtCfg := range initStatementCfgs {
			createTables = append(createTables, stmtCfg.CreateTableStatement)
			for _, alter := range stmtCfg.AlterTableStatements {
				if alter.ConstraintType == sqlmanager_shared.ForeignConstraintType {
					fkAlterStmts = append(fkAlterStmts, alter.Statement)
				} else {
					nonFkAlterStmts = append(nonFkAlterStmts, alter.Statement)
				}
			}
			idxStmts = append(idxStmts, stmtCfg.IndexStatements...)
		}
		return nil
	})

	tableViewsStmts := []string{}
	errgrp.Go(func() error {
		views, err := m.getViewsBySchemas(ctx, schemas)
		if err != nil {
			return fmt.Errorf("unable to retrieve mssql schema table triggers: %w", err)
		}
		for _, v := range views {
			tableViewsStmts = append(tableViewsStmts, v.Definition)
		}
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}

	return []*sqlmanager_shared.InitSchemaStatements{
		{Label: "data types", Statements: dataTypeStmts},
		{Label: "create table", Statements: slices.Concat(createTables, tableViewsStmts)},
		{Label: "non-fk alter table", Statements: nonFkAlterStmts},
		{Label: "fk alter table", Statements: fkAlterStmts},
		{Label: "table index", Statements: idxStmts},
		{Label: "table triggers", Statements: tableTriggerStmts},
	}, nil
}

func (m *Manager) GetCreateTableStatement(ctx context.Context, schema, table string) (string, error) {
	return "", errors.ErrUnsupported
}

func (m *Manager) GetSchemaTableDataTypes(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) (*sqlmanager_shared.SchemaTableDataTypeResponse, error) {
	if len(tables) == 0 {
		return &sqlmanager_shared.SchemaTableDataTypeResponse{}, nil
	}

	schemasMap := map[string]struct{}{}
	for _, t := range tables {
		schemasMap[t.Schema] = struct{}{}
	}
	schemas := []string{}
	for schema := range schemasMap {
		schemas = append(schemas, schema)
	}

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(3) // Limit this to effectively one set per schema

	output := &sqlmanager_shared.SchemaTableDataTypeResponse{}
	errgrp.Go(func() error {
		seqs, err := m.getSequencesBySchemas(errctx, schemas)
		if err != nil {
			return fmt.Errorf("unable to get mssql sequences by tables: %w", err)
		}
		output.Sequences = seqs
		return nil
	})
	errgrp.Go(func() error {
		funcs, err := m.getFunctionsBySchemas(errctx, schemas)
		if err != nil {
			return fmt.Errorf("unable to get mssql functions by tables: %w", err)
		}
		output.Functions = funcs
		return nil
	})
	errgrp.Go(func() error {
		datatypes, err := m.getDataTypesBySchemas(errctx, schemas)
		if err != nil {
			return fmt.Errorf("unable to get mssql data types by tables: %w", err)
		}
		output.Composites = datatypes.Composites
		output.Domains = datatypes.Domains
		return nil
	})
	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (m *Manager) GetSchemaTableTriggers(ctx context.Context, tables []*sqlmanager_shared.SchemaTable) ([]*sqlmanager_shared.TableTrigger, error) {
	if len(tables) == 0 {
		return []*sqlmanager_shared.TableTrigger{}, nil
	}

	combined := make([]string, 0, len(tables))
	for _, t := range tables {
		combined = append(combined, t.String())
	}

	rows, err := m.querier.GetCustomTriggersBySchemasAndTables(ctx, m.db, combined)
	if err != nil && !isNoRows(err) {
		return nil, err
	} else if err != nil && isNoRows(err) {
		return []*sqlmanager_shared.TableTrigger{}, nil
	}

	output := make([]*sqlmanager_shared.TableTrigger, 0, len(rows))
	for _, row := range rows {
		output = append(output, &sqlmanager_shared.TableTrigger{
			Schema:      row.SchemaName,
			Table:       row.TableName,
			TriggerName: row.TriggerName,
			Definition:  generateCreateTriggerStatement(row),
		})
	}
	return output, nil
}

func (m *Manager) getSequencesBySchemas(ctx context.Context, schemas []string) ([]*sqlmanager_shared.DataType, error) {
	rows, err := m.querier.GetCustomSequencesBySchemas(ctx, m.db, schemas)
	if err != nil && !isNoRows(err) {
		return nil, err
	} else if err != nil && isNoRows(err) {
		return []*sqlmanager_shared.DataType{}, nil
	}

	output := make([]*sqlmanager_shared.DataType, 0, len(rows))
	for _, row := range rows {
		output = append(output, &sqlmanager_shared.DataType{
			Schema:     row.SchemaName,
			Name:       row.SequenceName,
			Definition: generateCreateSequenceStatement(row),
		})
	}
	return output, nil
}

// todo remove this
func (m *Manager) GetSequencesByTables(ctx context.Context, schema string, tables []string) ([]*sqlmanager_shared.DataType, error) {
	rows, err := m.querier.GetCustomSequencesBySchemas(ctx, m.db, []string{schema})
	if err != nil && !isNoRows(err) {
		return nil, err
	} else if err != nil && isNoRows(err) {
		return []*sqlmanager_shared.DataType{}, nil
	}

	output := make([]*sqlmanager_shared.DataType, 0, len(rows))
	for _, row := range rows {
		output = append(output, &sqlmanager_shared.DataType{
			Schema:     row.SchemaName,
			Name:       row.SequenceName,
			Definition: generateCreateSequenceStatement(row),
		})
	}
	return output, nil
}

func (m *Manager) getFunctionsBySchemas(ctx context.Context, schemas []string) ([]*sqlmanager_shared.DataType, error) {
	rows, err := m.querier.GetCustomFunctionsBySchemas(ctx, m.db, schemas)
	if err != nil && !isNoRows(err) {
		return nil, err
	} else if err != nil && isNoRows(err) {
		return []*sqlmanager_shared.DataType{}, nil
	}

	output := make([]*sqlmanager_shared.DataType, 0, len(rows))
	for _, row := range rows {
		output = append(output, &sqlmanager_shared.DataType{
			Schema:     row.SchemaName,
			Name:       row.FunctionName,
			Definition: generateCreateFunctionStatement(row),
		})
	}
	return output, nil
}

func (m *Manager) getViewsBySchemas(ctx context.Context, schemas []string) ([]*sqlmanager_shared.DataType, error) {
	rows, err := m.querier.GetCustomViewsBySchemas(ctx, m.db, schemas)
	if err != nil && !isNoRows(err) {
		return nil, err
	} else if err != nil && isNoRows(err) {
		return []*sqlmanager_shared.DataType{}, nil
	}

	output := make([]*sqlmanager_shared.DataType, 0, len(rows))
	for _, row := range rows {
		output = append(output, &sqlmanager_shared.DataType{
			Schema:     row.SchemaName,
			Name:       row.ViewName,
			Definition: generateCreateViewStatement(row),
		})
	}
	return output, nil
}

type datatypes struct {
	Composites []*sqlmanager_shared.DataType
	Domains    []*sqlmanager_shared.DataType
}

func (m *Manager) getDataTypesBySchemas(ctx context.Context, schemas []string) (*datatypes, error) {
	rows, err := m.querier.GetDataTypesBySchemas(ctx, m.db, schemas)
	if err != nil && !isNoRows(err) {
		return nil, err
	} else if err != nil && isNoRows(err) {
		return &datatypes{}, nil
	}

	output := &datatypes{}

	for _, row := range rows {
		dt := &sqlmanager_shared.DataType{
			Schema:     row.SchemaName,
			Name:       row.TypeName,
			Definition: generateCreateDataTypeStatement(row),
		}
		switch row.Type {
		case "composite":
			output.Composites = append(output.Composites, dt)
		case "domain":
			output.Domains = append(output.Domains, dt)
		}
	}
	return output, nil
}

func isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
