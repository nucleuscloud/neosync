package mssql_queries

type Querier interface{}

var _ Querier = (*Queries)(nil)
