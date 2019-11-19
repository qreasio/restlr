// Code generated by MockGen. DO NOT EDIT.
// Source: shared/repository.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model "github.com/qreasio/restlr/model"
)

// MockRepository is a mock of Repository interface
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// LoadOption mocks base method
func (m *MockRepository) LoadOption(ctx context.Context, optionName string) (*model.Option, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadOption", ctx, optionName)
	ret0, _ := ret[0].(*model.Option)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoadOption indicates an expected call of LoadOption
func (mr *MockRepositoryMockRecorder) LoadOption(ctx, optionName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadOption", reflect.TypeOf((*MockRepository)(nil).LoadOption), ctx, optionName)
}

// PostMetasByPostIDs mocks base method
func (m *MockRepository) PostMetasByPostIDs(ctx context.Context, idList []uint64) (map[uint64]map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostMetasByPostIDs", ctx, idList)
	ret0, _ := ret[0].(map[uint64]map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PostMetasByPostIDs indicates an expected call of PostMetasByPostIDs
func (mr *MockRepositoryMockRecorder) PostMetasByPostIDs(ctx, idList interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostMetasByPostIDs", reflect.TypeOf((*MockRepository)(nil).PostMetasByPostIDs), ctx, idList)
}