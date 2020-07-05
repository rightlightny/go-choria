// Code generated by MockGen. DO NOT EDIT.
// Source: entity.go

// Package scout is a generated GoMock package.
package scout

import (
	gomock "github.com/golang/mock/gomock"
	nats_go "github.com/nats-io/nats.go"
	logrus "github.com/sirupsen/logrus"
	reflect "reflect"
)

// MockFramework is a mock of Framework interface
type MockFramework struct {
	ctrl     *gomock.Controller
	recorder *MockFrameworkMockRecorder
}

// MockFrameworkMockRecorder is the mock recorder for MockFramework
type MockFrameworkMockRecorder struct {
	mock *MockFramework
}

// NewMockFramework creates a new mock instance
func NewMockFramework(ctrl *gomock.Controller) *MockFramework {
	mock := &MockFramework{ctrl: ctrl}
	mock.recorder = &MockFrameworkMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFramework) EXPECT() *MockFrameworkMockRecorder {
	return m.recorder
}

// Identity mocks base method
func (m *MockFramework) Identity() string {
	ret := m.ctrl.Call(m, "Identity")
	ret0, _ := ret[0].(string)
	return ret0
}

// Identity indicates an expected call of Identity
func (mr *MockFrameworkMockRecorder) Identity() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Identity", reflect.TypeOf((*MockFramework)(nil).Identity))
}

// NATSConn mocks base method
func (m *MockFramework) NATSConn() *nats_go.Conn {
	ret := m.ctrl.Call(m, "NATSConn")
	ret0, _ := ret[0].(*nats_go.Conn)
	return ret0
}

// NATSConn indicates an expected call of NATSConn
func (mr *MockFrameworkMockRecorder) NATSConn() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NATSConn", reflect.TypeOf((*MockFramework)(nil).NATSConn))
}

// Logger mocks base method
func (m *MockFramework) Logger(arg0 string) *logrus.Entry {
	ret := m.ctrl.Call(m, "Logger", arg0)
	ret0, _ := ret[0].(*logrus.Entry)
	return ret0
}

// Logger indicates an expected call of Logger
func (mr *MockFrameworkMockRecorder) Logger(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Logger", reflect.TypeOf((*MockFramework)(nil).Logger), arg0)
}

// OverridesFile mocks base method
func (m *MockFramework) OverridesFile() string {
	ret := m.ctrl.Call(m, "OverridesFile")
	ret0, _ := ret[0].(string)
	return ret0
}

// OverridesFile indicates an expected call of OverridesFile
func (mr *MockFrameworkMockRecorder) OverridesFile() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OverridesFile", reflect.TypeOf((*MockFramework)(nil).OverridesFile))
}

// Tags mocks base method
func (m *MockFramework) Tags() ([]string, error) {
	ret := m.ctrl.Call(m, "Tags")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Tags indicates an expected call of Tags
func (mr *MockFrameworkMockRecorder) Tags() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tags", reflect.TypeOf((*MockFramework)(nil).Tags))
}

// MachineSourceDir mocks base method
func (m *MockFramework) MachineSourceDir() string {
	ret := m.ctrl.Call(m, "MachineSourceDir")
	ret0, _ := ret[0].(string)
	return ret0
}

// MachineSourceDir indicates an expected call of MachineSourceDir
func (mr *MockFrameworkMockRecorder) MachineSourceDir() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MachineSourceDir", reflect.TypeOf((*MockFramework)(nil).MachineSourceDir))
}
