// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/xescugc/maze-wars/server (interfaces: WSConnector)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	websocket "nhooyr.io/websocket"
)

// MockWSConnector is a mock of WSConnector interface.
type MockWSConnector struct {
	ctrl     *gomock.Controller
	recorder *MockWSConnectorMockRecorder
}

// MockWSConnectorMockRecorder is the mock recorder for MockWSConnector.
type MockWSConnectorMockRecorder struct {
	mock *MockWSConnector
}

// NewMockWSConnector creates a new mock instance.
func NewMockWSConnector(ctrl *gomock.Controller) *MockWSConnector {
	mock := &MockWSConnector{ctrl: ctrl}
	mock.recorder = &MockWSConnectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWSConnector) EXPECT() *MockWSConnectorMockRecorder {
	return m.recorder
}

// Read mocks base method.
func (m *MockWSConnector) Read(arg0 context.Context, arg1 *websocket.Conn, arg2 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Read indicates an expected call of Read.
func (mr *MockWSConnectorMockRecorder) Read(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockWSConnector)(nil).Read), arg0, arg1, arg2)
}

// Write mocks base method.
func (m *MockWSConnector) Write(arg0 context.Context, arg1 *websocket.Conn, arg2 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Write indicates an expected call of Write.
func (mr *MockWSConnectorMockRecorder) Write(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockWSConnector)(nil).Write), arg0, arg1, arg2)
}