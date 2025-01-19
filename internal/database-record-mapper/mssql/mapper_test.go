package mssql

import (
	"testing"
	"time"

	mssql "github.com/microsoft/go-mssqldb"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseRowValues(t *testing.T) {
	t.Parallel()

	testTime := time.Now()
	testUUID := mssql.UniqueIdentifier{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF, 0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF}
	testBinary := []byte{0x12, 0x34}
	testVarBinary := []byte{0x56, 0x78}
	testString := []byte("test string")

	values := []any{
		testTime,      // datetime
		&testUUID,     // uniqueidentifier
		testBinary,    // binary
		testVarBinary, // varbinary
		testString,    // other bytes
		42,            // default case
	}

	columnNames := []string{
		"datetime_col",
		"uuid_col",
		"binary_col",
		"varbinary_col",
		"string_col",
		"int_col",
	}

	columnDbTypes := []string{
		"DATETIME",
		"UNIQUEIDENTIFIER",
		"BINARY",
		"VARBINARY",
		"VARCHAR",
		"INT",
	}

	result, err := parseRowValues(values, columnNames, columnDbTypes)
	require.NoError(t, err)

	// Test datetime handling
	dt, err := neosynctypes.NewDateTimeFromMssql(testTime)
	assert.NoError(t, err)
	assert.Equal(t, dt, result["datetime_col"])

	// Test UUID handling
	assert.Equal(t, testUUID.String(), result["uuid_col"])

	// Test binary handling
	expectedBinary, err := neosynctypes.NewBinaryFromMssql(testBinary)
	assert.NoError(t, err)
	assert.Equal(t, expectedBinary, result["binary_col"])

	// Test varbinary handling
	expectedBits, err := neosynctypes.NewBitsFromMssql(testVarBinary)
	assert.NoError(t, err)
	assert.Equal(t, expectedBits, result["varbinary_col"])

	// Test string bytes handling
	assert.Equal(t, string(testString), result["string_col"])

	// Test default case
	assert.Equal(t, 42, result["int_col"])
}
