package dbconnectconfig

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
