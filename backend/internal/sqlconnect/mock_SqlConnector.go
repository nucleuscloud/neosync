// Code generated by mockery. DO NOT EDIT.

package sqlconnect

import (
	context "context"

	pgxpool "github.com/jackc/pgx/v5/pgxpool"
	mock "github.com/stretchr/testify/mock"

	sql "database/sql"
)

// MockSqlConnector is an autogenerated mock type for the SqlConnector type
type MockSqlConnector struct {
	mock.Mock
}

type MockSqlConnector_Expecter struct {
	mock *mock.Mock
}

func (_m *MockSqlConnector) EXPECT() *MockSqlConnector_Expecter {
	return &MockSqlConnector_Expecter{mock: &_m.Mock}
}

// MysqlOpen provides a mock function with given fields: connectionStr
func (_m *MockSqlConnector) MysqlOpen(connectionStr string) (*sql.DB, error) {
	ret := _m.Called(connectionStr)

	if len(ret) == 0 {
		panic("no return value specified for MysqlOpen")
	}

	var r0 *sql.DB
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*sql.DB, error)); ok {
		return rf(connectionStr)
	}
	if rf, ok := ret.Get(0).(func(string) *sql.DB); ok {
		r0 = rf(connectionStr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sql.DB)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(connectionStr)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockSqlConnector_MysqlOpen_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MysqlOpen'
type MockSqlConnector_MysqlOpen_Call struct {
	*mock.Call
}

// MysqlOpen is a helper method to define mock.On call
//   - connectionStr string
func (_e *MockSqlConnector_Expecter) MysqlOpen(connectionStr interface{}) *MockSqlConnector_MysqlOpen_Call {
	return &MockSqlConnector_MysqlOpen_Call{Call: _e.mock.On("MysqlOpen", connectionStr)}
}

func (_c *MockSqlConnector_MysqlOpen_Call) Run(run func(connectionStr string)) *MockSqlConnector_MysqlOpen_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockSqlConnector_MysqlOpen_Call) Return(_a0 *sql.DB, _a1 error) *MockSqlConnector_MysqlOpen_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockSqlConnector_MysqlOpen_Call) RunAndReturn(run func(string) (*sql.DB, error)) *MockSqlConnector_MysqlOpen_Call {
	_c.Call.Return(run)
	return _c
}

// Open provides a mock function with given fields: driverName, connectionStr
func (_m *MockSqlConnector) Open(driverName string, connectionStr string) (*sql.DB, error) {
	ret := _m.Called(driverName, connectionStr)

	if len(ret) == 0 {
		panic("no return value specified for Open")
	}

	var r0 *sql.DB
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (*sql.DB, error)); ok {
		return rf(driverName, connectionStr)
	}
	if rf, ok := ret.Get(0).(func(string, string) *sql.DB); ok {
		r0 = rf(driverName, connectionStr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sql.DB)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(driverName, connectionStr)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockSqlConnector_Open_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Open'
type MockSqlConnector_Open_Call struct {
	*mock.Call
}

// Open is a helper method to define mock.On call
//   - driverName string
//   - connectionStr string
func (_e *MockSqlConnector_Expecter) Open(driverName interface{}, connectionStr interface{}) *MockSqlConnector_Open_Call {
	return &MockSqlConnector_Open_Call{Call: _e.mock.On("Open", driverName, connectionStr)}
}

func (_c *MockSqlConnector_Open_Call) Run(run func(driverName string, connectionStr string)) *MockSqlConnector_Open_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *MockSqlConnector_Open_Call) Return(_a0 *sql.DB, _a1 error) *MockSqlConnector_Open_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockSqlConnector_Open_Call) RunAndReturn(run func(string, string) (*sql.DB, error)) *MockSqlConnector_Open_Call {
	_c.Call.Return(run)
	return _c
}

// PgPoolOpen provides a mock function with given fields: ctx, connectionStr
func (_m *MockSqlConnector) PgPoolOpen(ctx context.Context, connectionStr string) (*pgxpool.Pool, error) {
	ret := _m.Called(ctx, connectionStr)

	if len(ret) == 0 {
		panic("no return value specified for PgPoolOpen")
	}

	var r0 *pgxpool.Pool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*pgxpool.Pool, error)); ok {
		return rf(ctx, connectionStr)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *pgxpool.Pool); ok {
		r0 = rf(ctx, connectionStr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pgxpool.Pool)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, connectionStr)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockSqlConnector_PgPoolOpen_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PgPoolOpen'
type MockSqlConnector_PgPoolOpen_Call struct {
	*mock.Call
}

// PgPoolOpen is a helper method to define mock.On call
//   - ctx context.Context
//   - connectionStr string
func (_e *MockSqlConnector_Expecter) PgPoolOpen(ctx interface{}, connectionStr interface{}) *MockSqlConnector_PgPoolOpen_Call {
	return &MockSqlConnector_PgPoolOpen_Call{Call: _e.mock.On("PgPoolOpen", ctx, connectionStr)}
}

func (_c *MockSqlConnector_PgPoolOpen_Call) Run(run func(ctx context.Context, connectionStr string)) *MockSqlConnector_PgPoolOpen_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockSqlConnector_PgPoolOpen_Call) Return(_a0 *pgxpool.Pool, _a1 error) *MockSqlConnector_PgPoolOpen_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockSqlConnector_PgPoolOpen_Call) RunAndReturn(run func(context.Context, string) (*pgxpool.Pool, error)) *MockSqlConnector_PgPoolOpen_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockSqlConnector creates a new instance of MockSqlConnector. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockSqlConnector(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockSqlConnector {
	mock := &MockSqlConnector{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
