// Code generated by MockGen. DO NOT EDIT.
// Source: internal/usecase/acl/group/update_group_by_id.go
//
// Generated by this command:
//
//	Cursor-0.48.8-x86_64.AppImage -source=internal/usecase/acl/group/update_group_by_id.go -destination=internal/usecase/acl/group/mock/mock_update_group_repo.go -package=group_mock
//

// Package group_mock is a generated GoMock package.
package group_mock

import (
	reflect "reflect"

	acl "github.com/jekiapp/topic-master/internal/model/acl"
	gomock "go.uber.org/mock/gomock"
)

// MockiUpdateGroupRepo is a mock of iUpdateGroupRepo interface.
type MockiUpdateGroupRepo struct {
	ctrl     *gomock.Controller
	recorder *MockiUpdateGroupRepoMockRecorder
}

// MockiUpdateGroupRepoMockRecorder is the mock recorder for MockiUpdateGroupRepo.
type MockiUpdateGroupRepoMockRecorder struct {
	mock *MockiUpdateGroupRepo
}

// NewMockiUpdateGroupRepo creates a new mock instance.
func NewMockiUpdateGroupRepo(ctrl *gomock.Controller) *MockiUpdateGroupRepo {
	mock := &MockiUpdateGroupRepo{ctrl: ctrl}
	mock.recorder = &MockiUpdateGroupRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockiUpdateGroupRepo) EXPECT() *MockiUpdateGroupRepoMockRecorder {
	return m.recorder
}

// GetGroupByID mocks base method.
func (m *MockiUpdateGroupRepo) GetGroupByID(id string) (acl.Group, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroupByID", id)
	ret0, _ := ret[0].(acl.Group)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGroupByID indicates an expected call of GetGroupByID.
func (mr *MockiUpdateGroupRepoMockRecorder) GetGroupByID(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroupByID", reflect.TypeOf((*MockiUpdateGroupRepo)(nil).GetGroupByID), id)
}

// UpdateGroup mocks base method.
func (m *MockiUpdateGroupRepo) UpdateGroup(group acl.Group) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateGroup", group)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateGroup indicates an expected call of UpdateGroup.
func (mr *MockiUpdateGroupRepoMockRecorder) UpdateGroup(group any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateGroup", reflect.TypeOf((*MockiUpdateGroupRepo)(nil).UpdateGroup), group)
}
