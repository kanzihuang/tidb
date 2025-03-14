// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/pingcap/tidb/disttask/framework/scheduler (interfaces: TaskTable,Pool,Scheduler,Extension)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	proto "github.com/pingcap/tidb/disttask/framework/proto"
	execute "github.com/pingcap/tidb/disttask/framework/scheduler/execute"
	gomock "go.uber.org/mock/gomock"
)

// MockTaskTable is a mock of TaskTable interface.
type MockTaskTable struct {
	ctrl     *gomock.Controller
	recorder *MockTaskTableMockRecorder
}

// MockTaskTableMockRecorder is the mock recorder for MockTaskTable.
type MockTaskTableMockRecorder struct {
	mock *MockTaskTable
}

// NewMockTaskTable creates a new mock instance.
func NewMockTaskTable(ctrl *gomock.Controller) *MockTaskTable {
	mock := &MockTaskTable{ctrl: ctrl}
	mock.recorder = &MockTaskTableMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTaskTable) EXPECT() *MockTaskTableMockRecorder {
	return m.recorder
}

// FinishSubtask mocks base method.
func (m *MockTaskTable) FinishSubtask(arg0 int64, arg1 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FinishSubtask", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// FinishSubtask indicates an expected call of FinishSubtask.
func (mr *MockTaskTableMockRecorder) FinishSubtask(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FinishSubtask", reflect.TypeOf((*MockTaskTable)(nil).FinishSubtask), arg0, arg1)
}

// GetFirstSubtaskInStates mocks base method.
func (m *MockTaskTable) GetFirstSubtaskInStates(arg0 string, arg1, arg2 int64, arg3 ...interface{}) (*proto.Subtask, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetFirstSubtaskInStates", varargs...)
	ret0, _ := ret[0].(*proto.Subtask)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFirstSubtaskInStates indicates an expected call of GetFirstSubtaskInStates.
func (mr *MockTaskTableMockRecorder) GetFirstSubtaskInStates(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFirstSubtaskInStates", reflect.TypeOf((*MockTaskTable)(nil).GetFirstSubtaskInStates), varargs...)
}

// GetGlobalTaskByID mocks base method.
func (m *MockTaskTable) GetGlobalTaskByID(arg0 int64) (*proto.Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGlobalTaskByID", arg0)
	ret0, _ := ret[0].(*proto.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGlobalTaskByID indicates an expected call of GetGlobalTaskByID.
func (mr *MockTaskTableMockRecorder) GetGlobalTaskByID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGlobalTaskByID", reflect.TypeOf((*MockTaskTable)(nil).GetGlobalTaskByID), arg0)
}

// GetGlobalTasksInStates mocks base method.
func (m *MockTaskTable) GetGlobalTasksInStates(arg0 ...interface{}) ([]*proto.Task, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetGlobalTasksInStates", varargs...)
	ret0, _ := ret[0].([]*proto.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGlobalTasksInStates indicates an expected call of GetGlobalTasksInStates.
func (mr *MockTaskTableMockRecorder) GetGlobalTasksInStates(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGlobalTasksInStates", reflect.TypeOf((*MockTaskTable)(nil).GetGlobalTasksInStates), arg0...)
}

// GetSubtasksInStates mocks base method.
func (m *MockTaskTable) GetSubtasksInStates(arg0 string, arg1, arg2 int64, arg3 ...interface{}) ([]*proto.Subtask, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetSubtasksInStates", varargs...)
	ret0, _ := ret[0].([]*proto.Subtask)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSubtasksInStates indicates an expected call of GetSubtasksInStates.
func (mr *MockTaskTableMockRecorder) GetSubtasksInStates(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSubtasksInStates", reflect.TypeOf((*MockTaskTable)(nil).GetSubtasksInStates), varargs...)
}

// HasSubtasksInStates mocks base method.
func (m *MockTaskTable) HasSubtasksInStates(arg0 string, arg1, arg2 int64, arg3 ...interface{}) (bool, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "HasSubtasksInStates", varargs...)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasSubtasksInStates indicates an expected call of HasSubtasksInStates.
func (mr *MockTaskTableMockRecorder) HasSubtasksInStates(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasSubtasksInStates", reflect.TypeOf((*MockTaskTable)(nil).HasSubtasksInStates), varargs...)
}

// IsSchedulerCanceled mocks base method.
func (m *MockTaskTable) IsSchedulerCanceled(arg0 string, arg1 int64) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsSchedulerCanceled", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsSchedulerCanceled indicates an expected call of IsSchedulerCanceled.
func (mr *MockTaskTableMockRecorder) IsSchedulerCanceled(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsSchedulerCanceled", reflect.TypeOf((*MockTaskTable)(nil).IsSchedulerCanceled), arg0, arg1)
}

// PauseSubtasks mocks base method.
func (m *MockTaskTable) PauseSubtasks(arg0 string, arg1 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PauseSubtasks", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PauseSubtasks indicates an expected call of PauseSubtasks.
func (mr *MockTaskTableMockRecorder) PauseSubtasks(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PauseSubtasks", reflect.TypeOf((*MockTaskTable)(nil).PauseSubtasks), arg0, arg1)
}

// StartManager mocks base method.
func (m *MockTaskTable) StartManager(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartManager", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// StartManager indicates an expected call of StartManager.
func (mr *MockTaskTableMockRecorder) StartManager(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartManager", reflect.TypeOf((*MockTaskTable)(nil).StartManager), arg0, arg1)
}

// StartSubtask mocks base method.
func (m *MockTaskTable) StartSubtask(arg0 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartSubtask", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// StartSubtask indicates an expected call of StartSubtask.
func (mr *MockTaskTableMockRecorder) StartSubtask(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartSubtask", reflect.TypeOf((*MockTaskTable)(nil).StartSubtask), arg0)
}

// UpdateErrorToSubtask mocks base method.
func (m *MockTaskTable) UpdateErrorToSubtask(arg0 string, arg1 int64, arg2 error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateErrorToSubtask", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateErrorToSubtask indicates an expected call of UpdateErrorToSubtask.
func (mr *MockTaskTableMockRecorder) UpdateErrorToSubtask(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateErrorToSubtask", reflect.TypeOf((*MockTaskTable)(nil).UpdateErrorToSubtask), arg0, arg1, arg2)
}

// UpdateSubtaskStateAndError mocks base method.
func (m *MockTaskTable) UpdateSubtaskStateAndError(arg0 int64, arg1 string, arg2 error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateSubtaskStateAndError", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateSubtaskStateAndError indicates an expected call of UpdateSubtaskStateAndError.
func (mr *MockTaskTableMockRecorder) UpdateSubtaskStateAndError(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSubtaskStateAndError", reflect.TypeOf((*MockTaskTable)(nil).UpdateSubtaskStateAndError), arg0, arg1, arg2)
}

// MockPool is a mock of Pool interface.
type MockPool struct {
	ctrl     *gomock.Controller
	recorder *MockPoolMockRecorder
}

// MockPoolMockRecorder is the mock recorder for MockPool.
type MockPoolMockRecorder struct {
	mock *MockPool
}

// NewMockPool creates a new mock instance.
func NewMockPool(ctrl *gomock.Controller) *MockPool {
	mock := &MockPool{ctrl: ctrl}
	mock.recorder = &MockPoolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPool) EXPECT() *MockPoolMockRecorder {
	return m.recorder
}

// ReleaseAndWait mocks base method.
func (m *MockPool) ReleaseAndWait() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ReleaseAndWait")
}

// ReleaseAndWait indicates an expected call of ReleaseAndWait.
func (mr *MockPoolMockRecorder) ReleaseAndWait() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReleaseAndWait", reflect.TypeOf((*MockPool)(nil).ReleaseAndWait))
}

// Run mocks base method.
func (m *MockPool) Run(arg0 func()) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockPoolMockRecorder) Run(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockPool)(nil).Run), arg0)
}

// RunWithConcurrency mocks base method.
func (m *MockPool) RunWithConcurrency(arg0 chan func(), arg1 uint32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RunWithConcurrency", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RunWithConcurrency indicates an expected call of RunWithConcurrency.
func (mr *MockPoolMockRecorder) RunWithConcurrency(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunWithConcurrency", reflect.TypeOf((*MockPool)(nil).RunWithConcurrency), arg0, arg1)
}

// MockScheduler is a mock of Scheduler interface.
type MockScheduler struct {
	ctrl     *gomock.Controller
	recorder *MockSchedulerMockRecorder
}

// MockSchedulerMockRecorder is the mock recorder for MockScheduler.
type MockSchedulerMockRecorder struct {
	mock *MockScheduler
}

// NewMockScheduler creates a new mock instance.
func NewMockScheduler(ctrl *gomock.Controller) *MockScheduler {
	mock := &MockScheduler{ctrl: ctrl}
	mock.recorder = &MockSchedulerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockScheduler) EXPECT() *MockSchedulerMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockScheduler) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockSchedulerMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockScheduler)(nil).Close))
}

// Init mocks base method.
func (m *MockScheduler) Init(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Init", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Init indicates an expected call of Init.
func (mr *MockSchedulerMockRecorder) Init(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockScheduler)(nil).Init), arg0)
}

// Pause mocks base method.
func (m *MockScheduler) Pause(arg0 context.Context, arg1 *proto.Task) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Pause", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Pause indicates an expected call of Pause.
func (mr *MockSchedulerMockRecorder) Pause(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pause", reflect.TypeOf((*MockScheduler)(nil).Pause), arg0, arg1)
}

// Rollback mocks base method.
func (m *MockScheduler) Rollback(arg0 context.Context, arg1 *proto.Task) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rollback", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Rollback indicates an expected call of Rollback.
func (mr *MockSchedulerMockRecorder) Rollback(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rollback", reflect.TypeOf((*MockScheduler)(nil).Rollback), arg0, arg1)
}

// Run mocks base method.
func (m *MockScheduler) Run(arg0 context.Context, arg1 *proto.Task) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run.
func (mr *MockSchedulerMockRecorder) Run(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockScheduler)(nil).Run), arg0, arg1)
}

// MockExtension is a mock of Extension interface.
type MockExtension struct {
	ctrl     *gomock.Controller
	recorder *MockExtensionMockRecorder
}

// MockExtensionMockRecorder is the mock recorder for MockExtension.
type MockExtensionMockRecorder struct {
	mock *MockExtension
}

// NewMockExtension creates a new mock instance.
func NewMockExtension(ctrl *gomock.Controller) *MockExtension {
	mock := &MockExtension{ctrl: ctrl}
	mock.recorder = &MockExtensionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExtension) EXPECT() *MockExtensionMockRecorder {
	return m.recorder
}

// GetSubtaskExecutor mocks base method.
func (m *MockExtension) GetSubtaskExecutor(arg0 context.Context, arg1 *proto.Task, arg2 *execute.Summary) (execute.SubtaskExecutor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSubtaskExecutor", arg0, arg1, arg2)
	ret0, _ := ret[0].(execute.SubtaskExecutor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSubtaskExecutor indicates an expected call of GetSubtaskExecutor.
func (mr *MockExtensionMockRecorder) GetSubtaskExecutor(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSubtaskExecutor", reflect.TypeOf((*MockExtension)(nil).GetSubtaskExecutor), arg0, arg1, arg2)
}
