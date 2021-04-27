// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Fantom-foundation/go-opera/gossip/emitter (interfaces: External,TxPool)

// Package mock is a generated GoMock package.
package mock

import (
	evmcore "github.com/Fantom-foundation/go-opera/evmcore"
	inter "github.com/Fantom-foundation/go-opera/inter"
	opera "github.com/Fantom-foundation/go-opera/opera"
	vecmt "github.com/Fantom-foundation/go-opera/vecmt"
	hash "github.com/Fantom-foundation/lachesis-base/hash"
	idx "github.com/Fantom-foundation/lachesis-base/inter/idx"
	pos "github.com/Fantom-foundation/lachesis-base/inter/pos"
	common "github.com/ethereum/go-ethereum/common"
	types "github.com/ethereum/go-ethereum/core/types"
	event "github.com/ethereum/go-ethereum/event"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockExternal is a mock of External interface
type MockExternal struct {
	ctrl     *gomock.Controller
	recorder *MockExternalMockRecorder
}

// MockExternalMockRecorder is the mock recorder for MockExternal
type MockExternalMockRecorder struct {
	mock *MockExternal
}

// NewMockExternal creates a new mock instance
func NewMockExternal(ctrl *gomock.Controller) *MockExternal {
	mock := &MockExternal{ctrl: ctrl}
	mock.recorder = &MockExternalMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockExternal) EXPECT() *MockExternalMockRecorder {
	return m.recorder
}

// Build mocks base method
func (m *MockExternal) Build(arg0 *inter.MutableEventPayload, arg1 func()) error {
	ret := m.ctrl.Call(m, "Build", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Build indicates an expected call of Build
func (mr *MockExternalMockRecorder) Build(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Build", reflect.TypeOf((*MockExternal)(nil).Build), arg0, arg1)
}

// Check mocks base method
func (m *MockExternal) Check(arg0 *inter.EventPayload, arg1 inter.Events) error {
	ret := m.ctrl.Call(m, "Check", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Check indicates an expected call of Check
func (mr *MockExternalMockRecorder) Check(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Check", reflect.TypeOf((*MockExternal)(nil).Check), arg0, arg1)
}

// DagIndex mocks base method
func (m *MockExternal) DagIndex() *vecmt.Index {
	ret := m.ctrl.Call(m, "DagIndex")
	ret0, _ := ret[0].(*vecmt.Index)
	return ret0
}

// DagIndex indicates an expected call of DagIndex
func (mr *MockExternalMockRecorder) DagIndex() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DagIndex", reflect.TypeOf((*MockExternal)(nil).DagIndex))
}

// GetEpochValidators mocks base method
func (m *MockExternal) GetEpochValidators() (*pos.Validators, idx.Epoch) {
	ret := m.ctrl.Call(m, "GetEpochValidators")
	ret0, _ := ret[0].(*pos.Validators)
	ret1, _ := ret[1].(idx.Epoch)
	return ret0, ret1
}

// GetEpochValidators indicates an expected call of GetEpochValidators
func (mr *MockExternalMockRecorder) GetEpochValidators() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEpochValidators", reflect.TypeOf((*MockExternal)(nil).GetEpochValidators))
}

// GetEvent mocks base method
func (m *MockExternal) GetEvent(arg0 hash.Event) *inter.Event {
	ret := m.ctrl.Call(m, "GetEvent", arg0)
	ret0, _ := ret[0].(*inter.Event)
	return ret0
}

// GetEvent indicates an expected call of GetEvent
func (mr *MockExternalMockRecorder) GetEvent(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEvent", reflect.TypeOf((*MockExternal)(nil).GetEvent), arg0)
}

// GetEventPayload mocks base method
func (m *MockExternal) GetEventPayload(arg0 hash.Event) *inter.EventPayload {
	ret := m.ctrl.Call(m, "GetEventPayload", arg0)
	ret0, _ := ret[0].(*inter.EventPayload)
	return ret0
}

// GetEventPayload indicates an expected call of GetEventPayload
func (mr *MockExternalMockRecorder) GetEventPayload(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEventPayload", reflect.TypeOf((*MockExternal)(nil).GetEventPayload), arg0)
}

// GetGenesisTime mocks base method
func (m *MockExternal) GetGenesisTime() inter.Timestamp {
	ret := m.ctrl.Call(m, "GetGenesisTime")
	ret0, _ := ret[0].(inter.Timestamp)
	return ret0
}

// GetGenesisTime indicates an expected call of GetGenesisTime
func (mr *MockExternalMockRecorder) GetGenesisTime() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGenesisTime", reflect.TypeOf((*MockExternal)(nil).GetGenesisTime))
}

// GetHeads mocks base method
func (m *MockExternal) GetHeads(arg0 idx.Epoch) hash.Events {
	ret := m.ctrl.Call(m, "GetHeads", arg0)
	ret0, _ := ret[0].(hash.Events)
	return ret0
}

// GetHeads indicates an expected call of GetHeads
func (mr *MockExternalMockRecorder) GetHeads(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeads", reflect.TypeOf((*MockExternal)(nil).GetHeads), arg0)
}

// GetLastEvent mocks base method
func (m *MockExternal) GetLastEvent(arg0 idx.Epoch, arg1 idx.ValidatorID) *hash.Event {
	ret := m.ctrl.Call(m, "GetLastEvent", arg0, arg1)
	ret0, _ := ret[0].(*hash.Event)
	return ret0
}

// GetLastEvent indicates an expected call of GetLastEvent
func (mr *MockExternalMockRecorder) GetLastEvent(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLastEvent", reflect.TypeOf((*MockExternal)(nil).GetLastEvent), arg0, arg1)
}

// GetLatestBlockIndex mocks base method
func (m *MockExternal) GetLatestBlockIndex() idx.Block {
	ret := m.ctrl.Call(m, "GetLatestBlockIndex")
	ret0, _ := ret[0].(idx.Block)
	return ret0
}

// GetLatestBlockIndex indicates an expected call of GetLatestBlockIndex
func (mr *MockExternalMockRecorder) GetLatestBlockIndex() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLatestBlockIndex", reflect.TypeOf((*MockExternal)(nil).GetLatestBlockIndex))
}

// GetRules mocks base method
func (m *MockExternal) GetRules() opera.Rules {
	ret := m.ctrl.Call(m, "GetRules")
	ret0, _ := ret[0].(opera.Rules)
	return ret0
}

// GetRules indicates an expected call of GetRules
func (mr *MockExternalMockRecorder) GetRules() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRules", reflect.TypeOf((*MockExternal)(nil).GetRules))
}

// IsBusy mocks base method
func (m *MockExternal) IsBusy() bool {
	ret := m.ctrl.Call(m, "IsBusy")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsBusy indicates an expected call of IsBusy
func (mr *MockExternalMockRecorder) IsBusy() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsBusy", reflect.TypeOf((*MockExternal)(nil).IsBusy))
}

// IsSynced mocks base method
func (m *MockExternal) IsSynced() bool {
	ret := m.ctrl.Call(m, "IsSynced")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsSynced indicates an expected call of IsSynced
func (mr *MockExternalMockRecorder) IsSynced() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsSynced", reflect.TypeOf((*MockExternal)(nil).IsSynced))
}

// Lock mocks base method
func (m *MockExternal) Lock() {
	m.ctrl.Call(m, "Lock")
}

// Lock indicates an expected call of Lock
func (mr *MockExternalMockRecorder) Lock() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Lock", reflect.TypeOf((*MockExternal)(nil).Lock))
}

// PeersNum mocks base method
func (m *MockExternal) PeersNum() int {
	ret := m.ctrl.Call(m, "PeersNum")
	ret0, _ := ret[0].(int)
	return ret0
}

// PeersNum indicates an expected call of PeersNum
func (mr *MockExternalMockRecorder) PeersNum() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PeersNum", reflect.TypeOf((*MockExternal)(nil).PeersNum))
}

// Process mocks base method
func (m *MockExternal) Process(arg0 *inter.EventPayload) error {
	ret := m.ctrl.Call(m, "Process", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Process indicates an expected call of Process
func (mr *MockExternalMockRecorder) Process(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Process", reflect.TypeOf((*MockExternal)(nil).Process), arg0)
}

// Unlock mocks base method
func (m *MockExternal) Unlock() {
	m.ctrl.Call(m, "Unlock")
}

// Unlock indicates an expected call of Unlock
func (mr *MockExternalMockRecorder) Unlock() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unlock", reflect.TypeOf((*MockExternal)(nil).Unlock))
}

// MockTxPool is a mock of TxPool interface
type MockTxPool struct {
	ctrl     *gomock.Controller
	recorder *MockTxPoolMockRecorder
}

// MockTxPoolMockRecorder is the mock recorder for MockTxPool
type MockTxPoolMockRecorder struct {
	mock *MockTxPool
}

// NewMockTxPool creates a new mock instance
func NewMockTxPool(ctrl *gomock.Controller) *MockTxPool {
	mock := &MockTxPool{ctrl: ctrl}
	mock.recorder = &MockTxPoolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTxPool) EXPECT() *MockTxPoolMockRecorder {
	return m.recorder
}

// Has mocks base method
func (m *MockTxPool) Has(arg0 common.Hash) bool {
	ret := m.ctrl.Call(m, "Has", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Has indicates an expected call of Has
func (mr *MockTxPoolMockRecorder) Has(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Has", reflect.TypeOf((*MockTxPool)(nil).Has), arg0)
}

// Pending mocks base method
func (m *MockTxPool) Pending() (map[common.Address]types.Transactions, error) {
	ret := m.ctrl.Call(m, "Pending")
	ret0, _ := ret[0].(map[common.Address]types.Transactions)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Pending indicates an expected call of Pending
func (mr *MockTxPoolMockRecorder) Pending() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pending", reflect.TypeOf((*MockTxPool)(nil).Pending))
}

// SubscribeNewTxsNotify mocks base method
func (m *MockTxPool) SubscribeNewTxsNotify(arg0 chan<- evmcore.NewTxsNotify) event.Subscription {
	ret := m.ctrl.Call(m, "SubscribeNewTxsNotify", arg0)
	ret0, _ := ret[0].(event.Subscription)
	return ret0
}

// SubscribeNewTxsNotify indicates an expected call of SubscribeNewTxsNotify
func (mr *MockTxPoolMockRecorder) SubscribeNewTxsNotify(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribeNewTxsNotify", reflect.TypeOf((*MockTxPool)(nil).SubscribeNewTxsNotify), arg0)
}
