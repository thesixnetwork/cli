// Code generated by mockery v2.46.0. DO NOT EDIT.

package mocks

import (
	context "context"

	cosmosclient "github.com/ignite/cli/v29/ignite/pkg/cosmosclient"
	mock "github.com/stretchr/testify/mock"
)

// Saver is an autogenerated mock type for the Saver type
type Saver struct {
	mock.Mock
}

type Saver_Expecter struct {
	mock *mock.Mock
}

func (_m *Saver) EXPECT() *Saver_Expecter {
	return &Saver_Expecter{mock: &_m.Mock}
}

// Save provides a mock function with given fields: _a0, _a1
func (_m *Saver) Save(_a0 context.Context, _a1 []cosmosclient.TX) error {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Save")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []cosmosclient.TX) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Saver_Save_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Save'
type Saver_Save_Call struct {
	*mock.Call
}

// Save is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 []cosmosclient.TX
func (_e *Saver_Expecter) Save(_a0 interface{}, _a1 interface{}) *Saver_Save_Call {
	return &Saver_Save_Call{Call: _e.mock.On("Save", _a0, _a1)}
}

func (_c *Saver_Save_Call) Run(run func(_a0 context.Context, _a1 []cosmosclient.TX)) *Saver_Save_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]cosmosclient.TX))
	})
	return _c
}

func (_c *Saver_Save_Call) Return(_a0 error) *Saver_Save_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Saver_Save_Call) RunAndReturn(run func(context.Context, []cosmosclient.TX) error) *Saver_Save_Call {
	_c.Call.Return(run)
	return _c
}

// NewSaver creates a new instance of Saver. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSaver(t interface {
	mock.TestingT
	Cleanup(func())
}) *Saver {
	mock := &Saver{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
