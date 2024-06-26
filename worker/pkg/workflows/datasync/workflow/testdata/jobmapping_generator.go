//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	pg_query "github.com/pganalyze/pg_query_go/v5"
	pgquery "github.com/wasilibs/go-pgquery"
)

type Input struct {
	Folder  string `json:"folder"`
	SqlFile string `json:"sql_file"`
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
				if col.GetColumnDef() != nil {
					columns = append(columns, &Column{
						Name: col.GetColumnDef().Colname,
					})
				}
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

type TemplateData struct {
	PackageName string
	Mappings    []*mgmtv1alpha1.JobMapping
}

func formatJobMappings(pkgName string, mappings []*mgmtv1alpha1.JobMapping) (string, error) {
	const tmpl = `
package {{ .PackageName }}

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

func getDefaultJobMappings()[]*mgmtv1alpha1.JobMapping {
  return []*mgmtv1alpha1.JobMapping{
		{{- range .Mappings }}
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
}

`
	data := TemplateData{
		PackageName: pkgName,
		Mappings:    mappings,
	}
	t := template.Must(template.New("jobmappings").Parse(tmpl))
	var out bytes.Buffer
	err := t.Execute(&out, data)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func main() {
	args := os.Args
	// if len(args) < 4 {
	// 	panic("must provide necessary args")
	// }

	configFile := args[1]
	gopackage := args[2]

	packageSplit := strings.Split(gopackage, "_")
	goPkg := packageSplit[len(packageSplit)-1]

	jsonFile, err := os.Open(configFile)
	if err != nil {
		fmt.Println("failed to open file: %s", err)
		return
	}
	defer jsonFile.Close()

	// Read the file content into a byte slice
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("failed to read file: %s", err)
		return
	}

	// Unmarshal the byte slice into the struct
	var inputs []Input
	if err := json.Unmarshal(byteValue, &inputs); err != nil {
		fmt.Println("failed to unmarshal JSON: %s", err)
		return
	}
	for _, input := range inputs {
		goPkgName := strings.ReplaceAll(fmt.Sprintf("%s_%s", goPkg, input.Folder), "-", "")
		sqlFile, err := os.Open(fmt.Sprintf("%s/%s", input.Folder, input.SqlFile))
		if err != nil {
			fmt.Println("failed to open file: %s", err)
		}

		// Read the file content into a byte slice
		byteValue, err := io.ReadAll(sqlFile)
		if err != nil {
			fmt.Println("failed to read file: %s", err)
		}

		// Convert the byte slice to a string
		sqlContent := string(byteValue)
		sqlFile.Close()

		tables, err := parseSQLSchema(sqlContent)
		if err != nil {
			fmt.Println("Error parsing SQL schema:", err)
			return
		}

		jobMapping := generateJobMapping(tables)

		formattedJobMappings, err := formatJobMappings(goPkgName, jobMapping)
		if err != nil {
			fmt.Println("Error formatting job mappings:", err)
			return
		}

		output := fmt.Sprintf("%s/job_mappings.go", input.Folder)
		outputFile, err := os.Create(output)
		if err != nil {
			fmt.Println("Error creating jobmapping.go file:", err)
			return
		}

		_, err = outputFile.WriteString(formattedJobMappings)
		if err != nil {
			fmt.Println("Error writing to jobmapping.go file:", err)
			return
		}
		outputFile.Close()
	}

	return
}
