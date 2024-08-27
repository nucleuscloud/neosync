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
				Driver:      "postgres",
				Host:        "localhost",
				Port:        ptr(int32(5432)),
				Database:    ptr("mydb"),
				User:        "test-user",
				Pass:        "test-pass",
				QueryParams: url.Values{"sslmode": []string{"verify"}},
			},
			expected: "postgres://test-user:test-pass@localhost:5432/mydb?sslmode=verify",
		},
		{
			name: "mysql",
			input: GeneralDbConnectConfig{
				Driver:      "mysql",
				Host:        "localhost",
				Port:        ptr(int32(3309)),
				Database:    ptr("mydb"),
				User:        "test-user",
				Pass:        "test-pass",
				Protocol:    ptr("tcp"),
				QueryParams: url.Values{"foo": []string{"bar"}},
			},
			expected: "test-user:test-pass@tcp(localhost:3309)/mydb?foo=bar",
		},
		{
			name: "mysql",
			input: GeneralDbConnectConfig{
				Driver:      "mysql",
				Host:        "localhost",
				Port:        ptr(int32(3309)),
				Database:    ptr("mydb"),
				User:        "specialuser!*-",
				Pass:        "46!ZfMv3@Uh8*-<",
				Protocol:    ptr("tcp"),
				QueryParams: url.Values{"foo": []string{"bar"}},
			},
			expected: "specialuser!*-:46!ZfMv3@Uh8*-<@tcp(localhost:3309)/mydb?foo=bar",
		},
		{
			name: "mssql",
			input: GeneralDbConnectConfig{
				Driver:      "sqlserver",
				Host:        "localhost",
				Port:        ptr(int32(1433)),
				Database:    ptr("myinstance"),
				User:        "sa",
				Pass:        "myStr0ngP@assword",
				QueryParams: url.Values{"database": []string{"master"}},
			},
			expected: "sqlserver://sa:myStr0ngP%40assword@localhost:1433/myinstance?database=master",
		},
		{
			name: "mssql-noinstance",
			input: GeneralDbConnectConfig{
				Driver:      "sqlserver",
				Host:        "localhost",
				Port:        ptr(int32(1433)),
				Database:    nil,
				User:        "sa",
				Pass:        "myStr0ngP@assword",
				QueryParams: url.Values{"database": []string{"master"}},
			},
			expected: "sqlserver://sa:myStr0ngP%40assword@localhost:1433?database=master",
		},
		{
			name: "mssql-noinstance-noport",
			input: GeneralDbConnectConfig{
				Driver:      "sqlserver",
				Host:        "localhost",
				Port:        nil,
				Database:    nil,
				User:        "sa",
				Pass:        "myStr0ngP@assword",
				QueryParams: url.Values{"database": []string{"master"}},
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
