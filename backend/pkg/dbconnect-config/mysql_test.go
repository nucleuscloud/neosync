package dbconnectconfig

import (
	"net/url"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/assert"
)

var (
	mysqlconnectionFixture = &mgmtv1alpha1.MysqlConnection{
		Host:     "localhost",
		Port:     3309,
		Name:     "mydb",
		User:     "test-user",
		Pass:     "test-pass",
		Protocol: "tcp",
	}
)

func Test_NewFromMysqlConnection_Connection(t *testing.T) {
	out, err := NewFromMysqlConnection(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
				Connection: mysqlconnectionFixture,
			},
		},
	}, ptr(uint32(5)))

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, out, &GeneralDbConnectConfig{
		driver:        "mysql",
		host:          "localhost",
		port:          ptr(int32(3309)),
		Database:      ptr("mydb"),
		user:          "test-user",
		pass:          "test-pass",
		mysqlProtocol: ptr("tcp"),
		queryParams:   url.Values{"timeout": []string{"5s"}, "multiStatements": []string{"true"}},
	})
}

func Test_NewFromMysqlConnection_Url_mysql(t *testing.T) {
	out, err := NewFromMysqlConnection(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
				Url: "mysql://myuser:mypassword@localhost:3306/mydatabase?ssl=true",
			},
		},
	}, ptr(uint32(5)))

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, out, &GeneralDbConnectConfig{
		driver:        "mysql",
		host:          "localhost",
		port:          ptr(int32(3306)),
		Database:      ptr("mydatabase"),
		user:          "myuser",
		pass:          "mypassword",
		mysqlProtocol: nil,
		queryParams:   url.Values{"ssl": []string{"true"}, "multiStatements": []string{"true"}, "timeout": []string{"5s"}},
	})
}
func Test_NewFromMysqlConnection_Url_mysqlx(t *testing.T) {
	out, err := NewFromMysqlConnection(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
				Url: "mysqlx://myuser:mypassword@localhost:3306/mydatabase?ssl=true",
			},
		},
	}, ptr(uint32(5)))

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, out, &GeneralDbConnectConfig{
		driver:        "mysqlx",
		host:          "localhost",
		port:          ptr(int32(3306)),
		Database:      ptr("mydatabase"),
		user:          "myuser",
		pass:          "mypassword",
		mysqlProtocol: nil,
		queryParams:   url.Values{"ssl": []string{"true"}, "multiStatements": []string{"true"}, "timeout": []string{"5s"}},
	})
}

func Test_NewFromMysqlConnection_Url_Error(t *testing.T) {
	_, err := NewFromMysqlConnection(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
				Url: "mysql://myuser:mypassword/mydatabase?ssl=true",
			},
		},
	}, ptr(uint32(5)))

	assert.Error(t, err)
}

func Test_NewFromMysqlConnection_Url_NoScheme(t *testing.T) {
	_, err := NewFromMysqlConnection(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
				Url: "mysqlxxx://myuser:mypassword@localhost:3306/mydatabase?ssl=true",
			},
		},
	}, ptr(uint32(5)))

	assert.Error(t, err)
}

func Test_NewFromMysqlConnection_Url_NoPort(t *testing.T) {
	_, err := NewFromMysqlConnection(&mgmtv1alpha1.ConnectionConfig_MysqlConfig{
		MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
			ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
				Url: "mysqlxxx://myuser:mypassword@localhost/mydatabase?ssl=true",
			},
		},
	}, ptr(uint32(5)))

	assert.Error(t, err)
}
