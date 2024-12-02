package sqlmanager

import sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"

type SqlConnection struct {
	database SqlDatabase
	driver   string
}

func (s *SqlConnection) Db() SqlDatabase {
	return s.database
}
func (s *SqlConnection) Driver() string {
	return s.driver
}

func NewPostgresSqlConnection(database SqlDatabase) *SqlConnection {
	return newSqlConnection(database, sqlmanager_shared.PostgresDriver)
}

func NewMysqlSqlConnection(database SqlDatabase) *SqlConnection {
	return newSqlConnection(database, sqlmanager_shared.MysqlDriver)
}

func NewMssqlSqlConnection(database SqlDatabase) *SqlConnection {
	return newSqlConnection(database, sqlmanager_shared.MssqlDriver)
}

func newSqlConnection(database SqlDatabase, driver string) *SqlConnection {
	return &SqlConnection{database: database, driver: driver}
}
