package databaserecordmapper

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/database-record-mapper/builder"
	"github.com/nucleuscloud/neosync/internal/database-record-mapper/dynamodb"
	"github.com/nucleuscloud/neosync/internal/database-record-mapper/mongodb"
	"github.com/nucleuscloud/neosync/internal/database-record-mapper/mssql"
	"github.com/nucleuscloud/neosync/internal/database-record-mapper/mysql"
	"github.com/nucleuscloud/neosync/internal/database-record-mapper/postgres"
)

func NewDatabaseRecordMapper(dbType string) (builder.DatabaseRecordMapper[any], error) {
	switch dbType {
	case sqlmanager_shared.PostgresDriver:
		return postgres.NewPostgresBuilder(), nil
	case sqlmanager_shared.MysqlDriver:
		return mysql.NewMySQLBuilder(), nil
	case sqlmanager_shared.MssqlDriver:
		return mssql.NewMSSQLBuilder(), nil
	case "dynamodb":
		return dynamodb.NewDynamoBuilder(), nil
	case "mongodb":
		return mongodb.NewMongoBuilder(), nil
	default:
		return nil, fmt.Errorf("database type %s not supported", dbType)
	}
}

func NewDatabaseRecordMapperFromConnection(connection *mgmtv1alpha1.Connection) (builder.DatabaseRecordMapper[any], error) {
	switch connection.GetConnectionConfig().GetConfig().(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		return NewDatabaseRecordMapper(sqlmanager_shared.PostgresDriver)
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		return NewDatabaseRecordMapper(sqlmanager_shared.MysqlDriver)
	case *mgmtv1alpha1.ConnectionConfig_MssqlConfig:
		return NewDatabaseRecordMapper(sqlmanager_shared.MssqlDriver)
	case *mgmtv1alpha1.ConnectionConfig_MongoConfig:
		return NewDatabaseRecordMapper("mongodb")
	case *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
		return NewDatabaseRecordMapper("dynamodb")
	default:
		return nil, fmt.Errorf("unsupported connection type: %T for database record mapper", connection.GetConnectionConfig().GetConfig())
	}
}
