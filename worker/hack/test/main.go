package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	dbschemas_postgres "github.com/nucleuscloud/neosync/worker/internal/dbschemas/postgres"
)

func main() {
	pool, err := pgxpool.New(context.Background(), "postgres://postgres:foofar@localhost:5433/nucleus?sslmode=disable")
	if err != nil {
		panic(err)
	}

	stmt, err := dbschemas_postgres.GetTableCreateStatement(context.Background(), pool, &dbschemas_postgres.GetTableCreateStatementRequest{
		Schema: "neosync_api",
		Table:  "user_identity_provider_associations",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(stmt)
}
