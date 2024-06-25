// go:build ignore
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	pg_query "github.com/pganalyze/pg_query_go/v5"
	pgquery "github.com/wasilibs/go-pgquery"
)

type Input struct {
	BasePath string `json:"base_path"`
	SqlPath  string `json:"sql_path"`
	// TransformerMap map[string]map[string]*mgmtv1alpha1.JobMappingTransformer
	OutputPath string `json:"output_path"`
}

type Column struct {
	Name string
}

type Table struct {
	Schema  string
	Name    string
	Columns []*Column
}

type JobMapping struct {
	Schema      string
	Table       string
	Column      string
	Transformer string
	Config      string
}

func parseSQLSchema(sql string) ([]*Table, error) {
	tree, err := pgquery.Parse(sql)
	if err != nil {
		return nil, err
	}

	tables := []*Table{}
	var schema string
	for _, stmt := range tree.GetStmts() {
		s := stmt.GetStmt()
		switch s.Node.(type) {
		case *pg_query.Node_CreateSchemaStmt:
			schema = s.GetCreateSchemaStmt().GetSchemaname()
		case *pg_query.Node_CreateStmt:
			table := s.GetCreateStmt().GetRelation().GetRelname()
			columns := []*Column{}
			for _, col := range s.GetCreateStmt().GetTableElts() {
				columns = append(columns, &Column{
					Name: col.GetColumnDef().Colname,
				})
			}
			tables = append(tables, &Table{
				Schema:  schema,
				Name:    table,
				Columns: columns,
			})
		}
	}
	return tables, nil
}

func generateJobMapping(tables []*Table) []*mgmtv1alpha1.JobMapping {
	mappings := []*mgmtv1alpha1.JobMapping{}
	for _, t := range tables {
		for _, c := range t.Columns {
			mappings = append(mappings, &mgmtv1alpha1.JobMapping{
				Schema: t.Schema,
				Table:  t.Name,
				Column: c.Name,
				Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
				},
			})

		}
	}
	return mappings
}

func formatJobMappings(mappings []*mgmtv1alpha1.JobMapping) (string, error) {
	const tmpl = `
package main

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

var JobMappings = []*mgmtv1alpha1.JobMapping{
	{{- range . }}
	{
		Schema: "{{ .Schema }}",
		Table:  "{{ .Table }}",
		Column: "{{ .Column }}",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{{- end }}
}
`
	t := template.Must(template.New("jobmappings").Parse(tmpl))
	var out bytes.Buffer
	err := t.Execute(&out, mappings)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func main() {
	sql := `
	CREATE SCHEMA IF NOT EXISTS vfk_hr;
	SET search_path TO vfk_hr;

	CREATE TABLE IF NOT EXISTS regions (
		region_id SERIAL PRIMARY KEY,
		region_name CHARACTER VARYING (25)
	);

	CREATE TABLE IF NOT EXISTS countries (
		country_id CHARACTER (2) PRIMARY KEY,
		country_name CHARACTER VARYING (40),
		region_id INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS locations (
		location_id SERIAL PRIMARY KEY,
		street_address CHARACTER VARYING (40),
		postal_code CHARACTER VARYING (12),
		city CHARACTER VARYING (30) NOT NULL,
		state_province CHARACTER VARYING (25),
		country_id CHARACTER (2) NOT NULL
	);

	CREATE TABLE IF NOT EXISTS departments (
		department_id SERIAL PRIMARY KEY,
		department_name CHARACTER VARYING (30) NOT NULL,
		location_id INTEGER
	);

	CREATE TABLE IF NOT EXISTS jobs (
		job_id SERIAL PRIMARY KEY,
		job_title CHARACTER VARYING (35) NOT NULL,
		min_salary NUMERIC (8, 2),
		max_salary NUMERIC (8, 2)
	);

	CREATE TABLE IF NOT EXISTS employees (
		employee_id SERIAL PRIMARY KEY,
		first_name CHARACTER VARYING (20),
		last_name CHARACTER VARYING (25) NOT NULL,
		email CHARACTER VARYING (100) NOT NULL,
		phone_number CHARACTER VARYING (20),
		hire_date DATE NOT NULL,
		job_id INTEGER NOT NULL,
		salary NUMERIC (8, 2) NOT NULL,
		manager_id INTEGER,
		department_id INTEGER
	);

	CREATE TABLE IF NOT EXISTS dependents (
		dependent_id SERIAL PRIMARY KEY,
		first_name CHARACTER VARYING (50) NOT NULL,
		last_name CHARACTER VARYING (50) NOT NULL,
		relationship CHARACTER VARYING (25) NOT NULL,
		employee_id INTEGER NOT NULL
	);
	`

	tables, err := parseSQLSchema(sql)
	if err != nil {
		fmt.Println("Error parsing SQL schema:", err)
		return
	}

	jobMapping := generateJobMapping(tables)

	formattedJobMappings, err := formatJobMappings(jobMapping)
	if err != nil {
		fmt.Println("Error formatting job mappings:", err)
		return
	}

	file, err := os.Create("/Users/alisha/Projects/neosync/worker/scripts/generators/mock.go")
	if err != nil {
		fmt.Println("Error creating jobmapping.go file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(formattedJobMappings)
	if err != nil {
		fmt.Println("Error writing to jobmapping.go file:", err)
		return
	}
	return
}
