// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.com/akita/mem/dram/internal/org (interfaces: Channel)

package cmdq

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	"gitlab.com/akita/akita/v2/sim"
	signal "gitlab.com/akita/mem/v2/dram/internal/signal"
)

// MockChannel is a mock of Channel interface
type MockChannel struct {
	ctrl     *gomock.Controller
	recorder *MockChannelMockRecorder
}

// MockChannelMockRecorder is the mock recorder for MockChannel
type MockChannelMockRecorder struct {
	mock *MockChannel
}

// NewMockChannel creates a new mock instance
func NewMockChannel(ctrl *gomock.Controller) *MockChannel {
	mock := &MockChannel{ctrl: ctrl}
	mock.recorder = &MockChannelMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockChannel) EXPECT() *MockChannelMockRecorder {
	return m.recorder
}

// GetReadyCommand mocks base method
func (m *MockChannel) GetReadyCommand(arg0 sim.VTimeInSec, arg1 *signal.Command) *signal.Command {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetReadyCommand", arg0, arg1)
	ret0, _ := ret[0].(*signal.Command)
	return ret0
}

// GetReadyCommand indicates an expected call of GetReadyCommand
func (mr *MockChannelMockRecorder) GetReadyCommand(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetReadyCommand", reflect.TypeOf((*MockChannel)(nil).GetReadyCommand), arg0, arg1)
}

// StartCommand mocks base method
func (m *MockChannel) StartCommand(arg0 sim.VTimeInSec, arg1 *signal.Command) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StartCommand", arg0, arg1)
}

// StartCommand indicates an expected call of StartCommand
func (mr *MockChannelMockRecorder) StartCommand(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartCommand", reflect.TypeOf((*MockChannel)(nil).StartCommand), arg0, arg1)
}

// Tick mocks base method
func (m *MockChannel) Tick(arg0 sim.VTimeInSec) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Tick", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Tick indicates an expected call of Tick
func (mr *MockChannelMockRecorder) Tick(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tick", reflect.TypeOf((*MockChannel)(nil).Tick), arg0)
}

// UpdateTiming mocks base method
func (m *MockChannel) UpdateTiming(arg0 sim.VTimeInSec, arg1 *signal.Command) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UpdateTiming", arg0, arg1)
}

// UpdateTiming indicates an expected call of UpdateTiming
func (mr *MockChannelMockRecorder) UpdateTiming(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateTiming", reflect.TypeOf((*MockChannel)(nil).UpdateTiming), arg0, arg1)
}
