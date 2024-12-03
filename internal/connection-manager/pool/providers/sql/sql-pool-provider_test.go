package pool_sql_provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewConnectionProvider(t *testing.T) {
	assert.NotNil(t, NewConnectionProvider(nil, nil, nil, nil))
}
