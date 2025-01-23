package sqldbtx

import (
	"context"
	"database/sql"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
)

// Used by sql connector and other modules that need to interact with the database along with other methods on the sql.DB object
type DBTX interface {
	mysql_queries.DBTX

	PingContext(context.Context) error
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}
