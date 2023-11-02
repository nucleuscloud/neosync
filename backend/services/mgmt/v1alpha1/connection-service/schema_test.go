package v1alpha1_connectionservice

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"

	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_GetConnectionSchema_Postgres(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	m.SqlMock.ExpectQuery(`SELECT (.+) FROM (.+) WHERE.*c.table_schema NOT IN \('pg_catalog', 'information_schema'\).*`).WillReturnRows(getRowsMock())

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)

	resp, err := m.Service.GetConnectionSchema(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest]{
		Msg: &mgmtv1alpha1.GetConnectionSchemaRequest{
			Id: mockConnectionId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Msg.GetSchemas()))
	assert.ElementsMatch(t, getDatabaseColumnsMock(), resp.Msg.Schemas)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_GetConnectionSchema_Mysql(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	m.SqlMock.ExpectQuery(`SELECT (.+) FROM (.+) WHERE.*c.table_schema NOT IN \('sys', 'performance_schema', 'mysql'\).*`).WillReturnRows(getRowsMock())

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, MysqlMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)

	resp, err := m.Service.GetConnectionSchema(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest]{
		Msg: &mgmtv1alpha1.GetConnectionSchemaRequest{
			Id: mockConnectionId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Msg.GetSchemas()))
	assert.ElementsMatch(t, getDatabaseColumnsMock(), resp.Msg.Schemas)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_GetConnectionSchema_NoRows(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	m.SqlMock.ExpectQuery(".*").WillReturnError(sql.ErrNoRows)

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)

	resp, err := m.Service.GetConnectionSchema(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest]{
		Msg: &mgmtv1alpha1.GetConnectionSchemaRequest{
			Id: mockConnectionId,
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, len(resp.Msg.GetSchemas()))
	assert.ElementsMatch(t, []*mgmtv1alpha1.DatabaseColumn{}, resp.Msg.Schemas)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_GetConnectionSchema_Error(t *testing.T) {
	m := createServiceMock(t)
	defer m.SqlDbMock.Close()

	m.SqlMock.ExpectQuery(".*").WillReturnError(errors.New("oh no"))

	connectionUuid, _ := nucleusdb.ToUuid(mockConnectionId)
	connection := getConnectionMock(mockAccountId, mockConnectionName, connectionUuid, PostgresMock)
	mockIsUserInAccount(m.UserAccountServiceMock, true)
	m.QuerierMock.On("GetConnectionById", context.Background(), mock.Anything, connectionUuid).Return(connection, nil)

	resp, err := m.Service.GetConnectionSchema(context.Background(), &connect.Request[mgmtv1alpha1.GetConnectionSchemaRequest]{
		Msg: &mgmtv1alpha1.GetConnectionSchemaRequest{
			Id: mockConnectionId,
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	if err := m.SqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func getDatabaseColumnsMock() []*mgmtv1alpha1.DatabaseColumn {
	return []*mgmtv1alpha1.DatabaseColumn{
		{Schema: "schema1", Table: "table1", Column: "column1", DataType: "datatype1"},
		{Schema: "schema2", Table: "table2", Column: "column2", DataType: "datatype2"},
	}
}

func getRowsMock() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"table_schema", "table_name", "column_name", "data_type"}).
		AddRow("schema1", "table1", "column1", "datatype1").
		AddRow("schema2", "table2", "column2", "datatype2")
}
