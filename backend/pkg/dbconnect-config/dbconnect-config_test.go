package dbconnectconfig

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GeneralDbConnectionConfig_Helper_Methods(t *testing.T) {
	cfg := GeneralDbConnectConfig{
		driver:      "postgres",
		host:        "localhost",
		port:        ptr(int32(5432)),
		database:    ptr("mydb"),
		user:        "test-user",
		pass:        "test-pass",
		queryParams: url.Values{"sslmode": []string{"verify"}},
	}
	require.Equal(t, cfg.GetDriver(), "postgres")
	require.Equal(t, cfg.GetHost(), "localhost")
	require.Equal(t, *cfg.GetPort(), int32(5432))
	require.Equal(t, cfg.GetUser(), "test-user")

	cfg.SetHost("foo")
	cfg.SetPort(5433)
	require.Equal(t, cfg.GetHost(), "foo")
	require.Equal(t, *cfg.GetPort(), int32(5433))
}

func Test_GeneralDbConnectionConfig_String(t *testing.T) {
	type testcase struct {
		name     string
		input    GeneralDbConnectConfig
		expected string
	}
	testcases := []testcase{
		{
			name:     "empty",
			input:    GeneralDbConnectConfig{},
			expected: "",
		},
		{
			name: "postgres",
			input: GeneralDbConnectConfig{
				driver:      "postgres",
				host:        "localhost",
				port:        ptr(int32(5432)),
				database:    ptr("mydb"),
				user:        "test-user",
				pass:        "test-pass",
				queryParams: url.Values{"sslmode": []string{"verify"}},
			},
			expected: "postgres://test-user:test-pass@localhost:5432/mydb?sslmode=verify",
		},
		{
			name: "mysql",
			input: GeneralDbConnectConfig{
				driver:        "mysql",
				host:          "localhost",
				port:          ptr(int32(3309)),
				database:      ptr("mydb"),
				user:          "test-user",
				pass:          "test-pass",
				mysqlProtocol: ptr("tcp"),
				queryParams:   url.Values{"foo": []string{"bar"}},
			},
			expected: "test-user:test-pass@tcp(localhost:3309)/mydb?foo=bar",
		},
		{
			name: "mysql",
			input: GeneralDbConnectConfig{
				driver:        "mysql",
				host:          "localhost",
				port:          ptr(int32(3309)),
				database:      ptr("mydb"),
				user:          "specialuser!*-",
				pass:          "46!ZfMv3@Uh8*-<",
				mysqlProtocol: ptr("tcp"),
				queryParams:   url.Values{"foo": []string{"bar"}},
			},
			expected: "specialuser!*-:46!ZfMv3@Uh8*-<@tcp(localhost:3309)/mydb?foo=bar",
		},
		{
			name: "mssql",
			input: GeneralDbConnectConfig{
				driver:      "sqlserver",
				host:        "localhost",
				port:        ptr(int32(1433)),
				database:    ptr("myinstance"),
				user:        "sa",
				pass:        "myStr0ngP@assword",
				queryParams: url.Values{"database": []string{"master"}},
			},
			expected: "sqlserver://sa:myStr0ngP%40assword@localhost:1433/myinstance?database=master",
		},
		{
			name: "mssql-noinstance",
			input: GeneralDbConnectConfig{
				driver:      "sqlserver",
				host:        "localhost",
				port:        ptr(int32(1433)),
				database:    nil,
				user:        "sa",
				pass:        "myStr0ngP@assword",
				queryParams: url.Values{"database": []string{"master"}},
			},
			expected: "sqlserver://sa:myStr0ngP%40assword@localhost:1433?database=master",
		},
		{
			name: "mssql-noinstance-noport",
			input: GeneralDbConnectConfig{
				driver:      "sqlserver",
				host:        "localhost",
				port:        nil,
				database:    nil,
				user:        "sa",
				pass:        "myStr0ngP@assword",
				queryParams: url.Values{"database": []string{"master"}},
			},
			expected: "sqlserver://sa:myStr0ngP%40assword@localhost?database=master",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.input.String(), tc.expected)
		})
	}
}
