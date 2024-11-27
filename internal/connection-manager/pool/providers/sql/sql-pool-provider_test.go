package pool_sql_provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewConnectionProvider(t *testing.T) {
	assert.NotNil(t, NewConnectionProvider(nil, nil, nil, nil))
}

// func Test_newPoolProvider_GetDb(t *testing.T) {
// 	provider := NewProvider(func(dsn string) (neosync_benthos_sql.SqlDbtx, error) {
// 		return neosync_benthos_sql.NewMockSqlDbtx(t), nil
// 	})
// 	assert.NotNil(t, provider)
// 	db, err := provider.GetDb("foo", "bar")
// 	assert.NoError(t, err)
// 	assert.NotNil(t, db)
// }
