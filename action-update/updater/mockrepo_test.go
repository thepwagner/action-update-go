// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package updater_test

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	updater2 "github.com/thepwagner/action-update/updater"
)

// mockRepo is an autogenerated mock type for the Repo type
type mockRepo struct {
	mock.Mock
}

// Branch provides a mock function with given fields:
func (_m *mockRepo) Branch() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Fetch provides a mock function with given fields: ctx, branch
func (_m *mockRepo) Fetch(ctx context.Context, branch string) error {
	ret := _m.Called(ctx, branch)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, branch)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewBranch provides a mock function with given fields: base, branch
func (_m *mockRepo) NewBranch(base string, branch string) error {
	ret := _m.Called(base, branch)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(base, branch)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Push provides a mock function with given fields: _a0, _a1
func (_m *mockRepo) Push(_a0 context.Context, _a1 ...updater2.Update) error {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...updater2.Update) error); ok {
		r0 = rf(_a0, _a1...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Root provides a mock function with given fields:
func (_m *mockRepo) Root() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// SetBranch provides a mock function with given fields: branch
func (_m *mockRepo) SetBranch(branch string) error {
	ret := _m.Called(branch)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(branch)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}