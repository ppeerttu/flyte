// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// Reservation is an autogenerated mock type for the Reservation type
type Reservation struct {
	mock.Mock
}

type Reservation_Expecter struct {
	mock *mock.Mock
}

func (_m *Reservation) EXPECT() *Reservation_Expecter {
	return &Reservation_Expecter{mock: &_m.Mock}
}

// Cancel provides a mock function with given fields:
func (_m *Reservation) Cancel() {
	_m.Called()
}

// Reservation_Cancel_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Cancel'
type Reservation_Cancel_Call struct {
	*mock.Call
}

// Cancel is a helper method to define mock.On call
func (_e *Reservation_Expecter) Cancel() *Reservation_Cancel_Call {
	return &Reservation_Cancel_Call{Call: _e.mock.On("Cancel")}
}

func (_c *Reservation_Cancel_Call) Run(run func()) *Reservation_Cancel_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Reservation_Cancel_Call) Return() *Reservation_Cancel_Call {
	_c.Call.Return()
	return _c
}

func (_c *Reservation_Cancel_Call) RunAndReturn(run func()) *Reservation_Cancel_Call {
	_c.Call.Return(run)
	return _c
}

// CancelAt provides a mock function with given fields: t
func (_m *Reservation) CancelAt(t time.Time) {
	_m.Called(t)
}

// Reservation_CancelAt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CancelAt'
type Reservation_CancelAt_Call struct {
	*mock.Call
}

// CancelAt is a helper method to define mock.On call
//   - t time.Time
func (_e *Reservation_Expecter) CancelAt(t interface{}) *Reservation_CancelAt_Call {
	return &Reservation_CancelAt_Call{Call: _e.mock.On("CancelAt", t)}
}

func (_c *Reservation_CancelAt_Call) Run(run func(t time.Time)) *Reservation_CancelAt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(time.Time))
	})
	return _c
}

func (_c *Reservation_CancelAt_Call) Return() *Reservation_CancelAt_Call {
	_c.Call.Return()
	return _c
}

func (_c *Reservation_CancelAt_Call) RunAndReturn(run func(time.Time)) *Reservation_CancelAt_Call {
	_c.Call.Return(run)
	return _c
}

// Delay provides a mock function with given fields:
func (_m *Reservation) Delay() time.Duration {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Delay")
	}

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// Reservation_Delay_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delay'
type Reservation_Delay_Call struct {
	*mock.Call
}

// Delay is a helper method to define mock.On call
func (_e *Reservation_Expecter) Delay() *Reservation_Delay_Call {
	return &Reservation_Delay_Call{Call: _e.mock.On("Delay")}
}

func (_c *Reservation_Delay_Call) Run(run func()) *Reservation_Delay_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Reservation_Delay_Call) Return(_a0 time.Duration) *Reservation_Delay_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Reservation_Delay_Call) RunAndReturn(run func() time.Duration) *Reservation_Delay_Call {
	_c.Call.Return(run)
	return _c
}

// DelayFrom provides a mock function with given fields: t
func (_m *Reservation) DelayFrom(t time.Time) time.Duration {
	ret := _m.Called(t)

	if len(ret) == 0 {
		panic("no return value specified for DelayFrom")
	}

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func(time.Time) time.Duration); ok {
		r0 = rf(t)
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// Reservation_DelayFrom_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DelayFrom'
type Reservation_DelayFrom_Call struct {
	*mock.Call
}

// DelayFrom is a helper method to define mock.On call
//   - t time.Time
func (_e *Reservation_Expecter) DelayFrom(t interface{}) *Reservation_DelayFrom_Call {
	return &Reservation_DelayFrom_Call{Call: _e.mock.On("DelayFrom", t)}
}

func (_c *Reservation_DelayFrom_Call) Run(run func(t time.Time)) *Reservation_DelayFrom_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(time.Time))
	})
	return _c
}

func (_c *Reservation_DelayFrom_Call) Return(_a0 time.Duration) *Reservation_DelayFrom_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Reservation_DelayFrom_Call) RunAndReturn(run func(time.Time) time.Duration) *Reservation_DelayFrom_Call {
	_c.Call.Return(run)
	return _c
}

// OK provides a mock function with given fields:
func (_m *Reservation) OK() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for OK")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Reservation_OK_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OK'
type Reservation_OK_Call struct {
	*mock.Call
}

// OK is a helper method to define mock.On call
func (_e *Reservation_Expecter) OK() *Reservation_OK_Call {
	return &Reservation_OK_Call{Call: _e.mock.On("OK")}
}

func (_c *Reservation_OK_Call) Run(run func()) *Reservation_OK_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Reservation_OK_Call) Return(_a0 bool) *Reservation_OK_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Reservation_OK_Call) RunAndReturn(run func() bool) *Reservation_OK_Call {
	_c.Call.Return(run)
	return _c
}

// NewReservation creates a new instance of Reservation. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewReservation(t interface {
	mock.TestingT
	Cleanup(func())
}) *Reservation {
	mock := &Reservation{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
