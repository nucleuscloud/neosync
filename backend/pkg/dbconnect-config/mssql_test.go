package dbconnectconfig

import (
	"net/url"
	"reflect"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/zeebo/assert"
)

func Test_NewFromMssqlConnection(t *testing.T) {
	t.Run("standard string url", func(t *testing.T) {
		out, err := NewFromMssqlConnection(&mgmtv1alpha1.ConnectionConfig_MssqlConfig{
			MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
				ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
					Url: "sqlserver://test-user:test-pass@localhost:1433/myinstance?database=master",
				},
			},
		}, ptr(uint32(5)))

		assert.NoError(t, err)
		assert.NotNil(t, out)
		expected := &GeneralDbConnectConfig{
			driver:      "sqlserver",
			host:        "localhost",
			port:        ptr(int32(1433)),
			Database:    ptr("myinstance"),
			User:        "test-user",
			Pass:        "test-pass",
			queryParams: url.Values{"database": []string{"master"}, "connection timeout": []string{"5"}},
		}
		if !reflect.DeepEqual(out, expected) {
			t.Errorf("Expected %v, got %v", expected, out)
		}
	})
}

func ptr[T any](val T) *T {
	return &val
}
