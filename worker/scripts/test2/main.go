package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	// _ "github.com/lib/pq"
	_ "github.com/jackc/pgx/v5/stdlib"
	pgutil "github.com/nucleuscloud/neosync/internal/postgres"
	querybuilder "github.com/nucleuscloud/neosync/worker/pkg/query-builder"
)

func main() {
	// Replace these with your actual database connection details
	connStr := "postgresql://postgres:foofar@localhost:5434/nucleus?sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	sqlQuery := `SELECT
	  id,
	  int_array,
	  smallint_array,
	  bigint_array,
	  real_array,
	  double_array,
	  text_array,
	  varchar_array,
	  char_array,
	  boolean_array,
	  date_array,
	  time_array,
	  timestamp_array,
	  timestamptz_array,
	  interval_array,
	  inet_array,
	  cidr_array,
	  point_array,
	  line_array,
	  lseg_array,
	  path_array,
	  polygon_array,
	  circle_array,
	  uuid_array,
	  json_array,
	  jsonb_array,
	  bit_array,
	  varbit_array,
	  numeric_array,
	  money_array,
	  xml_array,
	  int_double_array
	FROM public.array_types_example;`
	// sqlQuery := `select * from public."matrix_data";`

	// Execute a query (adjust this query to match your database schema)
	rows, err := db.Query(sqlQuery)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		panic(err)
	}

	// Process the results
	insertValues := [][]any{}
	for rows.Next() {
		result, err := SqlRowToPgTypesMap(rows)
		if err != nil {
			log.Fatal(err)
		}
		// jsonF, _ := json.MarshalIndent(result, "", " ")
		fmt.Println(result)
		// fmt.Printf("%s \n", string(jsonF))
		args := []any{}
		for _, c := range columnNames {
			args = append(args, result[c])
		}
		insertValues = append(insertValues, args)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println()
	jsonF, _ := json.MarshalIndent(insertValues, "", " ")
	fmt.Printf("insertValues: %s \n", string(jsonF))
	onConflict := false
	insertQuery, err := querybuilder.BuildInsertQuery("pgx", "public.array_types_example", columnNames, insertValues, &onConflict)
	if err != nil {
		panic(err)
	}

	destConnStr := "postgresql://postgres:foofar@localhost:5435/nucleus?sslmode=disable"
	destDb, err := sql.Open("pgx", destConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer destDb.Close()

	if _, err := destDb.Exec(insertQuery); err != nil {
		panic(err)
	}
	fmt.Println("inserted data")
}

func SqlRowToPgTypesMap(rows *sql.Rows) (map[string]any, error) {
	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	cTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	colTypes := map[string]*sql.ColumnType{}
	for _, ct := range cTypes {
		colTypes[ct.Name()] = ct
	}

	// m := pgtype.NewMap()
	// m.RegisterType(&Array[any])

	values := make([]any, len(columnNames))
	valuesWrapped := make([]any, 0, len(columnNames))
	for i := range values {
		valuesWrapped = append(valuesWrapped, &values[i])
	}
	if err := rows.Scan(valuesWrapped...); err != nil {
		return nil, err
	}

	jObj := map[string]any{}
	for i, v := range values {

		col := columnNames[i]
		dbTypeName := colTypes[col].DatabaseTypeName()
		// if col == "integer_array_col" {
		// fmt.Printf("%s %+v %T \n", col, v, v)
		// fmt.Printf("%+v \n", &valuesWrapped[i])
		// }
		jObj[col] = v
		if isPgArrayType(dbTypeName) {
			valStr, ok := v.(string)
			if !ok {
				jObj[col] = v
			}
			fmt.Println()
			fmt.Println(valStr)
			parser := pgutil.NewArrayParser()
			a := parser.Parse(valStr)
			fmt.Printf("%+v \n", a)
			jObj[col] = a
		} else {
			jObj[col] = v
		}
	}

	return jObj, nil
}

func isPgArrayType(dbTypeName string) bool {
	return strings.HasPrefix(dbTypeName, "_")
}
