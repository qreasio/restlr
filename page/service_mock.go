package page

import (
	"context"
	"reflect"

	"github.com/golang/mock/gomock"
	"github.com/qreasio/restlr/model"
)

// MockService is a mock of Service interface
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// GetPage mocks base method
func (m *MockService) GetPage(ctx context.Context, req model.GetItemRequest) (interface{}, error) {
	if *req.ID > 99998 {
		return nil, model.ErrInvalidPostID
	}

	b := model.Base{ID: *req.ID}
	cv := model.ContentView{Base: b}
	return model.Post{ContentView: cv}, nil
}

// GetPage indicates an expected call of GetPage
func (mr *MockServiceMockRecorder) GetPage(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPage", reflect.TypeOf((*MockService)(nil).GetPage), ctx, req)
}

// ListPages mocks base method
func (m *MockService) ListPages(ctx context.Context, params model.ListRequest) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPages", ctx, params)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPages indicates an expected call of ListPages
func (mr *MockServiceMockRecorder) ListPages(ctx, params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPages", reflect.TypeOf((*MockService)(nil).ListPages), ctx, params)
}
