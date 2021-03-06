// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/controllers/trial/dbclient/dbclient.go

// Package mock_dbclient is a generated GoMock package.
package mock_dbclient

import (
	v1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockDBClient is a mock of DBClient interface
type MockDBClient struct {
	ctrl     *gomock.Controller
	recorder *MockDBClientMockRecorder
}

// MockDBClientMockRecorder is the mock recorder for MockDBClient
type MockDBClientMockRecorder struct {
	mock *MockDBClient
}

// NewMockDBClient creates a new mock instance
func NewMockDBClient(ctrl *gomock.Controller) *MockDBClient {
	mock := &MockDBClient{ctrl: ctrl}
	mock.recorder = &MockDBClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDBClient) EXPECT() *MockDBClientMockRecorder {
	return m.recorder
}

// GetTrialResult mocks base method
func (m *MockDBClient) GetTrialResult(trial *v1alpha1.Trial) (*v1alpha1.TrialResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTrialResult", trial)
	ret0, _ := ret[0].(*v1alpha1.TrialResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTrialResult indicates an expected call of GetTrialResult
func (mr *MockDBClientMockRecorder) GetTrialResult(trial interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTrialResult", reflect.TypeOf((*MockDBClient)(nil).GetTrialResult), trial)
}
