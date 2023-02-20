// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.com/akita/util/akitaext (interfaces: BufferedSender)

package mmu

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	"gitlab.com/akita/akita/v2/sim"
)

// MockBufferedSender is a mock of BufferedSender interface
type MockBufferedSender struct {
	ctrl     *gomock.Controller
	recorder *MockBufferedSenderMockRecorder
}

// MockBufferedSenderMockRecorder is the mock recorder for MockBufferedSender
type MockBufferedSenderMockRecorder struct {
	mock *MockBufferedSender
}

// NewMockBufferedSender creates a new mock instance
func NewMockBufferedSender(ctrl *gomock.Controller) *MockBufferedSender {
	mock := &MockBufferedSender{ctrl: ctrl}
	mock.recorder = &MockBufferedSenderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBufferedSender) EXPECT() *MockBufferedSenderMockRecorder {
	return m.recorder
}

// CanSend mocks base method
func (m *MockBufferedSender) CanSend(arg0 int) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CanSend", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// CanSend indicates an expected call of CanSend
func (mr *MockBufferedSenderMockRecorder) CanSend(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CanSend", reflect.TypeOf((*MockBufferedSender)(nil).CanSend), arg0)
}

// Clear mocks base method
func (m *MockBufferedSender) Clear() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Clear")
}

// Clear indicates an expected call of Clear
func (mr *MockBufferedSenderMockRecorder) Clear() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Clear", reflect.TypeOf((*MockBufferedSender)(nil).Clear))
}

// Send mocks base method
func (m *MockBufferedSender) Send(arg0 sim.Msg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Send", arg0)
}

// Send indicates an expected call of Send
func (mr *MockBufferedSenderMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockBufferedSender)(nil).Send), arg0)
}

// Tick mocks base method
func (m *MockBufferedSender) Tick(arg0 sim.VTimeInSec) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Tick", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Tick indicates an expected call of Tick
func (mr *MockBufferedSenderMockRecorder) Tick(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tick", reflect.TypeOf((*MockBufferedSender)(nil).Tick), arg0)
}
