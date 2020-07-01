// Code generated by mockery v2.0.3. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Cacher is an autogenerated mock type for the Cacher type
type Cacher struct {
	mock.Mock
}

// Get provides a mock function with given fields: key, data
func (_m *Cacher) Get(key string, data interface{}) error {
	ret := _m.Called(key, data)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, interface{}) error); ok {
		r0 = rf(key, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Set provides a mock function with given fields: key, data
func (_m *Cacher) Set(key string, data interface{}) error {
	ret := _m.Called(key, data)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, interface{}) error); ok {
		r0 = rf(key, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
