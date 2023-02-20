// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.com/akita/akita/v3/pipelining (interfaces: Pipeline)

package writeevict

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	pipelining "gitlab.com/akita/akita/v3/pipelining"
	sim "gitlab.com/akita/akita/v3/sim"
)

// MockPipeline is a mock of Pipeline interface.
type MockPipeline struct {
	ctrl     *gomock.Controller
	recorder *MockPipelineMockRecorder
}

// MockPipelineMockRecorder is the mock recorder for MockPipeline.
type MockPipelineMockRecorder struct {
	mock *MockPipeline
}

// NewMockPipeline creates a new mock instance.
func NewMockPipeline(ctrl *gomock.Controller) *MockPipeline {
	mock := &MockPipeline{ctrl: ctrl}
	mock.recorder = &MockPipelineMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPipeline) EXPECT() *MockPipelineMockRecorder {
	return m.recorder
}

// Accept mocks base method.
func (m *MockPipeline) Accept(arg0 sim.VTimeInSec, arg1 pipelining.PipelineItem) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Accept", arg0, arg1)
}

// Accept indicates an expected call of Accept.
func (mr *MockPipelineMockRecorder) Accept(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Accept", reflect.TypeOf((*MockPipeline)(nil).Accept), arg0, arg1)
}

// AcceptHook mocks base method.
func (m *MockPipeline) AcceptHook(arg0 sim.Hook) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AcceptHook", arg0)
}

// AcceptHook indicates an expected call of AcceptHook.
func (mr *MockPipelineMockRecorder) AcceptHook(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AcceptHook", reflect.TypeOf((*MockPipeline)(nil).AcceptHook), arg0)
}

// CanAccept mocks base method.
func (m *MockPipeline) CanAccept() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CanAccept")
	ret0, _ := ret[0].(bool)
	return ret0
}

// CanAccept indicates an expected call of CanAccept.
func (mr *MockPipelineMockRecorder) CanAccept() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CanAccept", reflect.TypeOf((*MockPipeline)(nil).CanAccept))
}

// Clear mocks base method.
func (m *MockPipeline) Clear() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Clear")
}

// Clear indicates an expected call of Clear.
func (mr *MockPipelineMockRecorder) Clear() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Clear", reflect.TypeOf((*MockPipeline)(nil).Clear))
}

// InvokeHook mocks base method.
func (m *MockPipeline) InvokeHook(arg0 sim.HookCtx) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "InvokeHook", arg0)
}

// InvokeHook indicates an expected call of InvokeHook.
func (mr *MockPipelineMockRecorder) InvokeHook(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InvokeHook", reflect.TypeOf((*MockPipeline)(nil).InvokeHook), arg0)
}

// Name mocks base method.
func (m *MockPipeline) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockPipelineMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockPipeline)(nil).Name))
}

// NumHooks mocks base method.
func (m *MockPipeline) NumHooks() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NumHooks")
	ret0, _ := ret[0].(int)
	return ret0
}

// NumHooks indicates an expected call of NumHooks.
func (mr *MockPipelineMockRecorder) NumHooks() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NumHooks", reflect.TypeOf((*MockPipeline)(nil).NumHooks))
}

// Tick mocks base method.
func (m *MockPipeline) Tick(arg0 sim.VTimeInSec) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Tick", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Tick indicates an expected call of Tick.
func (mr *MockPipelineMockRecorder) Tick(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tick", reflect.TypeOf((*MockPipeline)(nil).Tick), arg0)
}
