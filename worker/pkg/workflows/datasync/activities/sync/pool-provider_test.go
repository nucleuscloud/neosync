package sync_activity

// import (
// 	"testing"

// 	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
// 	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/internal/benthos/sql"
// 	"github.com/stretchr/testify/assert"
// )

// func Test_newPoolProvider(t *testing.T) {
// 	assert.NotNil(t, newPoolProvider(nil))
// }

// func Test_newPoolProvider_GetDb(t *testing.T) {
// 	provider := newPoolProvider(func(dsn string) (neosync_benthos_sql.SqlDbtx, error) {
// 		return mysql_queries.NewMockDBTX(t), nil
// 	})
// 	assert.NotNil(t, provider)
// 	db, err := provider.GetDb("foo", "bar")
// 	assert.NoError(t, err)
// 	assert.NotNil(t, db)
// }
