package main

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
	// import the dialect
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

func main() {
	builder := goqu.Dialect("postgres")

	table := goqu.S("public").Table("users")

	query := builder.From(table).Select("id", "name", "last_updated")
	query = query.Where(goqu.Ex{
		"name": "123",
		"id":   "1",
	})

	sql, params, err := query.Prepared(true).ToSQL()
	if err != nil {
		panic(err)
	}
	fmt.Println("sql", sql)
	fmt.Println("params", params)
}
