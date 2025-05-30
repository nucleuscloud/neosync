// Code generated by mockery. DO NOT EDIT.

package mssql_queries

import (
	context "context"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	mock "github.com/stretchr/testify/mock"
)

// MockQuerier is an autogenerated mock type for the Querier type
type MockQuerier struct {
	mock.Mock
}

type MockQuerier_Expecter struct {
	mock *mock.Mock
}

func (_m *MockQuerier) EXPECT() *MockQuerier_Expecter {
	return &MockQuerier_Expecter{mock: &_m.Mock}
}

// GetAllSchemas provides a mock function with given fields: ctx, db
func (_m *MockQuerier) GetAllSchemas(ctx context.Context, db mysql_queries.DBTX) ([]string, error) {
	ret := _m.Called(ctx, db)

	if len(ret) == 0 {
		panic("no return value specified for GetAllSchemas")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX) ([]string, error)); ok {
		return rf(ctx, db)
	}
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX) []string); ok {
		r0 = rf(ctx, db)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, mysql_queries.DBTX) error); ok {
		r1 = rf(ctx, db)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockQuerier_GetAllSchemas_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAllSchemas'
type MockQuerier_GetAllSchemas_Call struct {
	*mock.Call
}

// GetAllSchemas is a helper method to define mock.On call
//   - ctx context.Context
//   - db mysql_queries.DBTX
func (_e *MockQuerier_Expecter) GetAllSchemas(ctx interface{}, db interface{}) *MockQuerier_GetAllSchemas_Call {
	return &MockQuerier_GetAllSchemas_Call{Call: _e.mock.On("GetAllSchemas", ctx, db)}
}

func (_c *MockQuerier_GetAllSchemas_Call) Run(run func(ctx context.Context, db mysql_queries.DBTX)) *MockQuerier_GetAllSchemas_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(mysql_queries.DBTX))
	})
	return _c
}

func (_c *MockQuerier_GetAllSchemas_Call) Return(_a0 []string, _a1 error) *MockQuerier_GetAllSchemas_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockQuerier_GetAllSchemas_Call) RunAndReturn(run func(context.Context, mysql_queries.DBTX) ([]string, error)) *MockQuerier_GetAllSchemas_Call {
	_c.Call.Return(run)
	return _c
}

// GetAllTables provides a mock function with given fields: ctx, db
func (_m *MockQuerier) GetAllTables(ctx context.Context, db mysql_queries.DBTX) ([]*GetAllTablesRow, error) {
	ret := _m.Called(ctx, db)

	if len(ret) == 0 {
		panic("no return value specified for GetAllTables")
	}

	var r0 []*GetAllTablesRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX) ([]*GetAllTablesRow, error)); ok {
		return rf(ctx, db)
	}
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX) []*GetAllTablesRow); ok {
		r0 = rf(ctx, db)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*GetAllTablesRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, mysql_queries.DBTX) error); ok {
		r1 = rf(ctx, db)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockQuerier_GetAllTables_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAllTables'
type MockQuerier_GetAllTables_Call struct {
	*mock.Call
}

// GetAllTables is a helper method to define mock.On call
//   - ctx context.Context
//   - db mysql_queries.DBTX
func (_e *MockQuerier_Expecter) GetAllTables(ctx interface{}, db interface{}) *MockQuerier_GetAllTables_Call {
	return &MockQuerier_GetAllTables_Call{Call: _e.mock.On("GetAllTables", ctx, db)}
}

func (_c *MockQuerier_GetAllTables_Call) Run(run func(ctx context.Context, db mysql_queries.DBTX)) *MockQuerier_GetAllTables_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(mysql_queries.DBTX))
	})
	return _c
}

func (_c *MockQuerier_GetAllTables_Call) Return(_a0 []*GetAllTablesRow, _a1 error) *MockQuerier_GetAllTables_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockQuerier_GetAllTables_Call) RunAndReturn(run func(context.Context, mysql_queries.DBTX) ([]*GetAllTablesRow, error)) *MockQuerier_GetAllTables_Call {
	_c.Call.Return(run)
	return _c
}

// GetCustomSequencesBySchemas provides a mock function with given fields: ctx, db, schemas
func (_m *MockQuerier) GetCustomSequencesBySchemas(ctx context.Context, db mysql_queries.DBTX, schemas []string) ([]*GetCustomSequencesBySchemasRow, error) {
	ret := _m.Called(ctx, db, schemas)

	if len(ret) == 0 {
		panic("no return value specified for GetCustomSequencesBySchemas")
	}

	var r0 []*GetCustomSequencesBySchemasRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) ([]*GetCustomSequencesBySchemasRow, error)); ok {
		return rf(ctx, db, schemas)
	}
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) []*GetCustomSequencesBySchemasRow); ok {
		r0 = rf(ctx, db, schemas)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*GetCustomSequencesBySchemasRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, mysql_queries.DBTX, []string) error); ok {
		r1 = rf(ctx, db, schemas)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockQuerier_GetCustomSequencesBySchemas_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCustomSequencesBySchemas'
type MockQuerier_GetCustomSequencesBySchemas_Call struct {
	*mock.Call
}

// GetCustomSequencesBySchemas is a helper method to define mock.On call
//   - ctx context.Context
//   - db mysql_queries.DBTX
//   - schemas []string
func (_e *MockQuerier_Expecter) GetCustomSequencesBySchemas(ctx interface{}, db interface{}, schemas interface{}) *MockQuerier_GetCustomSequencesBySchemas_Call {
	return &MockQuerier_GetCustomSequencesBySchemas_Call{Call: _e.mock.On("GetCustomSequencesBySchemas", ctx, db, schemas)}
}

func (_c *MockQuerier_GetCustomSequencesBySchemas_Call) Run(run func(ctx context.Context, db mysql_queries.DBTX, schemas []string)) *MockQuerier_GetCustomSequencesBySchemas_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(mysql_queries.DBTX), args[2].([]string))
	})
	return _c
}

func (_c *MockQuerier_GetCustomSequencesBySchemas_Call) Return(_a0 []*GetCustomSequencesBySchemasRow, _a1 error) *MockQuerier_GetCustomSequencesBySchemas_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockQuerier_GetCustomSequencesBySchemas_Call) RunAndReturn(run func(context.Context, mysql_queries.DBTX, []string) ([]*GetCustomSequencesBySchemasRow, error)) *MockQuerier_GetCustomSequencesBySchemas_Call {
	_c.Call.Return(run)
	return _c
}

// GetCustomTriggersBySchemasAndTables provides a mock function with given fields: ctx, db, schematables
func (_m *MockQuerier) GetCustomTriggersBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetCustomTriggersBySchemasAndTablesRow, error) {
	ret := _m.Called(ctx, db, schematables)

	if len(ret) == 0 {
		panic("no return value specified for GetCustomTriggersBySchemasAndTables")
	}

	var r0 []*GetCustomTriggersBySchemasAndTablesRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) ([]*GetCustomTriggersBySchemasAndTablesRow, error)); ok {
		return rf(ctx, db, schematables)
	}
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) []*GetCustomTriggersBySchemasAndTablesRow); ok {
		r0 = rf(ctx, db, schematables)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*GetCustomTriggersBySchemasAndTablesRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, mysql_queries.DBTX, []string) error); ok {
		r1 = rf(ctx, db, schematables)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockQuerier_GetCustomTriggersBySchemasAndTables_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCustomTriggersBySchemasAndTables'
type MockQuerier_GetCustomTriggersBySchemasAndTables_Call struct {
	*mock.Call
}

// GetCustomTriggersBySchemasAndTables is a helper method to define mock.On call
//   - ctx context.Context
//   - db mysql_queries.DBTX
//   - schematables []string
func (_e *MockQuerier_Expecter) GetCustomTriggersBySchemasAndTables(ctx interface{}, db interface{}, schematables interface{}) *MockQuerier_GetCustomTriggersBySchemasAndTables_Call {
	return &MockQuerier_GetCustomTriggersBySchemasAndTables_Call{Call: _e.mock.On("GetCustomTriggersBySchemasAndTables", ctx, db, schematables)}
}

func (_c *MockQuerier_GetCustomTriggersBySchemasAndTables_Call) Run(run func(ctx context.Context, db mysql_queries.DBTX, schematables []string)) *MockQuerier_GetCustomTriggersBySchemasAndTables_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(mysql_queries.DBTX), args[2].([]string))
	})
	return _c
}

func (_c *MockQuerier_GetCustomTriggersBySchemasAndTables_Call) Return(_a0 []*GetCustomTriggersBySchemasAndTablesRow, _a1 error) *MockQuerier_GetCustomTriggersBySchemasAndTables_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockQuerier_GetCustomTriggersBySchemasAndTables_Call) RunAndReturn(run func(context.Context, mysql_queries.DBTX, []string) ([]*GetCustomTriggersBySchemasAndTablesRow, error)) *MockQuerier_GetCustomTriggersBySchemasAndTables_Call {
	_c.Call.Return(run)
	return _c
}

// GetDataTypesBySchemas provides a mock function with given fields: ctx, db, schematables
func (_m *MockQuerier) GetDataTypesBySchemas(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetDataTypesBySchemasRow, error) {
	ret := _m.Called(ctx, db, schematables)

	if len(ret) == 0 {
		panic("no return value specified for GetDataTypesBySchemas")
	}

	var r0 []*GetDataTypesBySchemasRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) ([]*GetDataTypesBySchemasRow, error)); ok {
		return rf(ctx, db, schematables)
	}
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) []*GetDataTypesBySchemasRow); ok {
		r0 = rf(ctx, db, schematables)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*GetDataTypesBySchemasRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, mysql_queries.DBTX, []string) error); ok {
		r1 = rf(ctx, db, schematables)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockQuerier_GetDataTypesBySchemas_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDataTypesBySchemas'
type MockQuerier_GetDataTypesBySchemas_Call struct {
	*mock.Call
}

// GetDataTypesBySchemas is a helper method to define mock.On call
//   - ctx context.Context
//   - db mysql_queries.DBTX
//   - schematables []string
func (_e *MockQuerier_Expecter) GetDataTypesBySchemas(ctx interface{}, db interface{}, schematables interface{}) *MockQuerier_GetDataTypesBySchemas_Call {
	return &MockQuerier_GetDataTypesBySchemas_Call{Call: _e.mock.On("GetDataTypesBySchemas", ctx, db, schematables)}
}

func (_c *MockQuerier_GetDataTypesBySchemas_Call) Run(run func(ctx context.Context, db mysql_queries.DBTX, schematables []string)) *MockQuerier_GetDataTypesBySchemas_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(mysql_queries.DBTX), args[2].([]string))
	})
	return _c
}

func (_c *MockQuerier_GetDataTypesBySchemas_Call) Return(_a0 []*GetDataTypesBySchemasRow, _a1 error) *MockQuerier_GetDataTypesBySchemas_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockQuerier_GetDataTypesBySchemas_Call) RunAndReturn(run func(context.Context, mysql_queries.DBTX, []string) ([]*GetDataTypesBySchemasRow, error)) *MockQuerier_GetDataTypesBySchemas_Call {
	_c.Call.Return(run)
	return _c
}

// GetDatabaseSchema provides a mock function with given fields: ctx, db
func (_m *MockQuerier) GetDatabaseSchema(ctx context.Context, db mysql_queries.DBTX) ([]*GetDatabaseSchemaRow, error) {
	ret := _m.Called(ctx, db)

	if len(ret) == 0 {
		panic("no return value specified for GetDatabaseSchema")
	}

	var r0 []*GetDatabaseSchemaRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX) ([]*GetDatabaseSchemaRow, error)); ok {
		return rf(ctx, db)
	}
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX) []*GetDatabaseSchemaRow); ok {
		r0 = rf(ctx, db)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*GetDatabaseSchemaRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, mysql_queries.DBTX) error); ok {
		r1 = rf(ctx, db)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockQuerier_GetDatabaseSchema_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDatabaseSchema'
type MockQuerier_GetDatabaseSchema_Call struct {
	*mock.Call
}

// GetDatabaseSchema is a helper method to define mock.On call
//   - ctx context.Context
//   - db mysql_queries.DBTX
func (_e *MockQuerier_Expecter) GetDatabaseSchema(ctx interface{}, db interface{}) *MockQuerier_GetDatabaseSchema_Call {
	return &MockQuerier_GetDatabaseSchema_Call{Call: _e.mock.On("GetDatabaseSchema", ctx, db)}
}

func (_c *MockQuerier_GetDatabaseSchema_Call) Run(run func(ctx context.Context, db mysql_queries.DBTX)) *MockQuerier_GetDatabaseSchema_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(mysql_queries.DBTX))
	})
	return _c
}

func (_c *MockQuerier_GetDatabaseSchema_Call) Return(_a0 []*GetDatabaseSchemaRow, _a1 error) *MockQuerier_GetDatabaseSchema_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockQuerier_GetDatabaseSchema_Call) RunAndReturn(run func(context.Context, mysql_queries.DBTX) ([]*GetDatabaseSchemaRow, error)) *MockQuerier_GetDatabaseSchema_Call {
	_c.Call.Return(run)
	return _c
}

// GetDatabaseTableSchemasBySchemasAndTables provides a mock function with given fields: ctx, db, schematables
func (_m *MockQuerier) GetDatabaseTableSchemasBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetDatabaseSchemaRow, error) {
	ret := _m.Called(ctx, db, schematables)

	if len(ret) == 0 {
		panic("no return value specified for GetDatabaseTableSchemasBySchemasAndTables")
	}

	var r0 []*GetDatabaseSchemaRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) ([]*GetDatabaseSchemaRow, error)); ok {
		return rf(ctx, db, schematables)
	}
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) []*GetDatabaseSchemaRow); ok {
		r0 = rf(ctx, db, schematables)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*GetDatabaseSchemaRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, mysql_queries.DBTX, []string) error); ok {
		r1 = rf(ctx, db, schematables)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockQuerier_GetDatabaseTableSchemasBySchemasAndTables_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDatabaseTableSchemasBySchemasAndTables'
type MockQuerier_GetDatabaseTableSchemasBySchemasAndTables_Call struct {
	*mock.Call
}

// GetDatabaseTableSchemasBySchemasAndTables is a helper method to define mock.On call
//   - ctx context.Context
//   - db mysql_queries.DBTX
//   - schematables []string
func (_e *MockQuerier_Expecter) GetDatabaseTableSchemasBySchemasAndTables(ctx interface{}, db interface{}, schematables interface{}) *MockQuerier_GetDatabaseTableSchemasBySchemasAndTables_Call {
	return &MockQuerier_GetDatabaseTableSchemasBySchemasAndTables_Call{Call: _e.mock.On("GetDatabaseTableSchemasBySchemasAndTables", ctx, db, schematables)}
}

func (_c *MockQuerier_GetDatabaseTableSchemasBySchemasAndTables_Call) Run(run func(ctx context.Context, db mysql_queries.DBTX, schematables []string)) *MockQuerier_GetDatabaseTableSchemasBySchemasAndTables_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(mysql_queries.DBTX), args[2].([]string))
	})
	return _c
}

func (_c *MockQuerier_GetDatabaseTableSchemasBySchemasAndTables_Call) Return(_a0 []*GetDatabaseSchemaRow, _a1 error) *MockQuerier_GetDatabaseTableSchemasBySchemasAndTables_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockQuerier_GetDatabaseTableSchemasBySchemasAndTables_Call) RunAndReturn(run func(context.Context, mysql_queries.DBTX, []string) ([]*GetDatabaseSchemaRow, error)) *MockQuerier_GetDatabaseTableSchemasBySchemasAndTables_Call {
	_c.Call.Return(run)
	return _c
}

// GetIndicesBySchemasAndTables provides a mock function with given fields: ctx, db, schematables
func (_m *MockQuerier) GetIndicesBySchemasAndTables(ctx context.Context, db mysql_queries.DBTX, schematables []string) ([]*GetIndicesBySchemasAndTablesRow, error) {
	ret := _m.Called(ctx, db, schematables)

	if len(ret) == 0 {
		panic("no return value specified for GetIndicesBySchemasAndTables")
	}

	var r0 []*GetIndicesBySchemasAndTablesRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) ([]*GetIndicesBySchemasAndTablesRow, error)); ok {
		return rf(ctx, db, schematables)
	}
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) []*GetIndicesBySchemasAndTablesRow); ok {
		r0 = rf(ctx, db, schematables)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*GetIndicesBySchemasAndTablesRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, mysql_queries.DBTX, []string) error); ok {
		r1 = rf(ctx, db, schematables)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockQuerier_GetIndicesBySchemasAndTables_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetIndicesBySchemasAndTables'
type MockQuerier_GetIndicesBySchemasAndTables_Call struct {
	*mock.Call
}

// GetIndicesBySchemasAndTables is a helper method to define mock.On call
//   - ctx context.Context
//   - db mysql_queries.DBTX
//   - schematables []string
func (_e *MockQuerier_Expecter) GetIndicesBySchemasAndTables(ctx interface{}, db interface{}, schematables interface{}) *MockQuerier_GetIndicesBySchemasAndTables_Call {
	return &MockQuerier_GetIndicesBySchemasAndTables_Call{Call: _e.mock.On("GetIndicesBySchemasAndTables", ctx, db, schematables)}
}

func (_c *MockQuerier_GetIndicesBySchemasAndTables_Call) Run(run func(ctx context.Context, db mysql_queries.DBTX, schematables []string)) *MockQuerier_GetIndicesBySchemasAndTables_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(mysql_queries.DBTX), args[2].([]string))
	})
	return _c
}

func (_c *MockQuerier_GetIndicesBySchemasAndTables_Call) Return(_a0 []*GetIndicesBySchemasAndTablesRow, _a1 error) *MockQuerier_GetIndicesBySchemasAndTables_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockQuerier_GetIndicesBySchemasAndTables_Call) RunAndReturn(run func(context.Context, mysql_queries.DBTX, []string) ([]*GetIndicesBySchemasAndTablesRow, error)) *MockQuerier_GetIndicesBySchemasAndTables_Call {
	_c.Call.Return(run)
	return _c
}

// GetRolePermissions provides a mock function with given fields: ctx, db
func (_m *MockQuerier) GetRolePermissions(ctx context.Context, db mysql_queries.DBTX) ([]*GetRolePermissionsRow, error) {
	ret := _m.Called(ctx, db)

	if len(ret) == 0 {
		panic("no return value specified for GetRolePermissions")
	}

	var r0 []*GetRolePermissionsRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX) ([]*GetRolePermissionsRow, error)); ok {
		return rf(ctx, db)
	}
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX) []*GetRolePermissionsRow); ok {
		r0 = rf(ctx, db)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*GetRolePermissionsRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, mysql_queries.DBTX) error); ok {
		r1 = rf(ctx, db)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockQuerier_GetRolePermissions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRolePermissions'
type MockQuerier_GetRolePermissions_Call struct {
	*mock.Call
}

// GetRolePermissions is a helper method to define mock.On call
//   - ctx context.Context
//   - db mysql_queries.DBTX
func (_e *MockQuerier_Expecter) GetRolePermissions(ctx interface{}, db interface{}) *MockQuerier_GetRolePermissions_Call {
	return &MockQuerier_GetRolePermissions_Call{Call: _e.mock.On("GetRolePermissions", ctx, db)}
}

func (_c *MockQuerier_GetRolePermissions_Call) Run(run func(ctx context.Context, db mysql_queries.DBTX)) *MockQuerier_GetRolePermissions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(mysql_queries.DBTX))
	})
	return _c
}

func (_c *MockQuerier_GetRolePermissions_Call) Return(_a0 []*GetRolePermissionsRow, _a1 error) *MockQuerier_GetRolePermissions_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockQuerier_GetRolePermissions_Call) RunAndReturn(run func(context.Context, mysql_queries.DBTX) ([]*GetRolePermissionsRow, error)) *MockQuerier_GetRolePermissions_Call {
	_c.Call.Return(run)
	return _c
}

// GetTableConstraintsBySchemas provides a mock function with given fields: ctx, db, schemas
func (_m *MockQuerier) GetTableConstraintsBySchemas(ctx context.Context, db mysql_queries.DBTX, schemas []string) ([]*GetTableConstraintsBySchemasRow, error) {
	ret := _m.Called(ctx, db, schemas)

	if len(ret) == 0 {
		panic("no return value specified for GetTableConstraintsBySchemas")
	}

	var r0 []*GetTableConstraintsBySchemasRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) ([]*GetTableConstraintsBySchemasRow, error)); ok {
		return rf(ctx, db, schemas)
	}
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) []*GetTableConstraintsBySchemasRow); ok {
		r0 = rf(ctx, db, schemas)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*GetTableConstraintsBySchemasRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, mysql_queries.DBTX, []string) error); ok {
		r1 = rf(ctx, db, schemas)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockQuerier_GetTableConstraintsBySchemas_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTableConstraintsBySchemas'
type MockQuerier_GetTableConstraintsBySchemas_Call struct {
	*mock.Call
}

// GetTableConstraintsBySchemas is a helper method to define mock.On call
//   - ctx context.Context
//   - db mysql_queries.DBTX
//   - schemas []string
func (_e *MockQuerier_Expecter) GetTableConstraintsBySchemas(ctx interface{}, db interface{}, schemas interface{}) *MockQuerier_GetTableConstraintsBySchemas_Call {
	return &MockQuerier_GetTableConstraintsBySchemas_Call{Call: _e.mock.On("GetTableConstraintsBySchemas", ctx, db, schemas)}
}

func (_c *MockQuerier_GetTableConstraintsBySchemas_Call) Run(run func(ctx context.Context, db mysql_queries.DBTX, schemas []string)) *MockQuerier_GetTableConstraintsBySchemas_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(mysql_queries.DBTX), args[2].([]string))
	})
	return _c
}

func (_c *MockQuerier_GetTableConstraintsBySchemas_Call) Return(_a0 []*GetTableConstraintsBySchemasRow, _a1 error) *MockQuerier_GetTableConstraintsBySchemas_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockQuerier_GetTableConstraintsBySchemas_Call) RunAndReturn(run func(context.Context, mysql_queries.DBTX, []string) ([]*GetTableConstraintsBySchemasRow, error)) *MockQuerier_GetTableConstraintsBySchemas_Call {
	_c.Call.Return(run)
	return _c
}

// GetUniqueIndexesBySchema provides a mock function with given fields: ctx, db, schemas
func (_m *MockQuerier) GetUniqueIndexesBySchema(ctx context.Context, db mysql_queries.DBTX, schemas []string) ([]*GetUniqueIndexesBySchemaRow, error) {
	ret := _m.Called(ctx, db, schemas)

	if len(ret) == 0 {
		panic("no return value specified for GetUniqueIndexesBySchema")
	}

	var r0 []*GetUniqueIndexesBySchemaRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) ([]*GetUniqueIndexesBySchemaRow, error)); ok {
		return rf(ctx, db, schemas)
	}
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) []*GetUniqueIndexesBySchemaRow); ok {
		r0 = rf(ctx, db, schemas)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*GetUniqueIndexesBySchemaRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, mysql_queries.DBTX, []string) error); ok {
		r1 = rf(ctx, db, schemas)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockQuerier_GetUniqueIndexesBySchema_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUniqueIndexesBySchema'
type MockQuerier_GetUniqueIndexesBySchema_Call struct {
	*mock.Call
}

// GetUniqueIndexesBySchema is a helper method to define mock.On call
//   - ctx context.Context
//   - db mysql_queries.DBTX
//   - schemas []string
func (_e *MockQuerier_Expecter) GetUniqueIndexesBySchema(ctx interface{}, db interface{}, schemas interface{}) *MockQuerier_GetUniqueIndexesBySchema_Call {
	return &MockQuerier_GetUniqueIndexesBySchema_Call{Call: _e.mock.On("GetUniqueIndexesBySchema", ctx, db, schemas)}
}

func (_c *MockQuerier_GetUniqueIndexesBySchema_Call) Run(run func(ctx context.Context, db mysql_queries.DBTX, schemas []string)) *MockQuerier_GetUniqueIndexesBySchema_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(mysql_queries.DBTX), args[2].([]string))
	})
	return _c
}

func (_c *MockQuerier_GetUniqueIndexesBySchema_Call) Return(_a0 []*GetUniqueIndexesBySchemaRow, _a1 error) *MockQuerier_GetUniqueIndexesBySchema_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockQuerier_GetUniqueIndexesBySchema_Call) RunAndReturn(run func(context.Context, mysql_queries.DBTX, []string) ([]*GetUniqueIndexesBySchemaRow, error)) *MockQuerier_GetUniqueIndexesBySchema_Call {
	_c.Call.Return(run)
	return _c
}

// GetViewsAndFunctionsBySchemas provides a mock function with given fields: ctx, db, schemas
func (_m *MockQuerier) GetViewsAndFunctionsBySchemas(ctx context.Context, db mysql_queries.DBTX, schemas []string) ([]*GetViewsAndFunctionsBySchemasRow, error) {
	ret := _m.Called(ctx, db, schemas)

	if len(ret) == 0 {
		panic("no return value specified for GetViewsAndFunctionsBySchemas")
	}

	var r0 []*GetViewsAndFunctionsBySchemasRow
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) ([]*GetViewsAndFunctionsBySchemasRow, error)); ok {
		return rf(ctx, db, schemas)
	}
	if rf, ok := ret.Get(0).(func(context.Context, mysql_queries.DBTX, []string) []*GetViewsAndFunctionsBySchemasRow); ok {
		r0 = rf(ctx, db, schemas)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*GetViewsAndFunctionsBySchemasRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, mysql_queries.DBTX, []string) error); ok {
		r1 = rf(ctx, db, schemas)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockQuerier_GetViewsAndFunctionsBySchemas_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetViewsAndFunctionsBySchemas'
type MockQuerier_GetViewsAndFunctionsBySchemas_Call struct {
	*mock.Call
}

// GetViewsAndFunctionsBySchemas is a helper method to define mock.On call
//   - ctx context.Context
//   - db mysql_queries.DBTX
//   - schemas []string
func (_e *MockQuerier_Expecter) GetViewsAndFunctionsBySchemas(ctx interface{}, db interface{}, schemas interface{}) *MockQuerier_GetViewsAndFunctionsBySchemas_Call {
	return &MockQuerier_GetViewsAndFunctionsBySchemas_Call{Call: _e.mock.On("GetViewsAndFunctionsBySchemas", ctx, db, schemas)}
}

func (_c *MockQuerier_GetViewsAndFunctionsBySchemas_Call) Run(run func(ctx context.Context, db mysql_queries.DBTX, schemas []string)) *MockQuerier_GetViewsAndFunctionsBySchemas_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(mysql_queries.DBTX), args[2].([]string))
	})
	return _c
}

func (_c *MockQuerier_GetViewsAndFunctionsBySchemas_Call) Return(_a0 []*GetViewsAndFunctionsBySchemasRow, _a1 error) *MockQuerier_GetViewsAndFunctionsBySchemas_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockQuerier_GetViewsAndFunctionsBySchemas_Call) RunAndReturn(run func(context.Context, mysql_queries.DBTX, []string) ([]*GetViewsAndFunctionsBySchemasRow, error)) *MockQuerier_GetViewsAndFunctionsBySchemas_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockQuerier creates a new instance of MockQuerier. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockQuerier(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockQuerier {
	mock := &MockQuerier{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
