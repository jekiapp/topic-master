package acl

import (
	"context"
	"errors"
	"testing"

	aclmodel "github.com/jekiapp/topic-master/internal/model/acl"
	usergroup_mock "github.com/jekiapp/topic-master/internal/usecase/acl/usergroup/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAssignUserToGroupUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		req       AssignUserToGroupRequest
		setupMock func(m *usergroup_mock.MockiUserGroupRepo)
		wantErr   bool
	}{
		{
			name:      "missing user_id",
			req:       AssignUserToGroupRequest{GroupID: "g1"},
			setupMock: func(m *usergroup_mock.MockiUserGroupRepo) {},
			wantErr:   true,
		},
		{
			name:      "missing group_id",
			req:       AssignUserToGroupRequest{UserID: "alice"},
			setupMock: func(m *usergroup_mock.MockiUserGroupRepo) {},
			wantErr:   true,
		},
		{
			name: "already assigned",
			req:  AssignUserToGroupRequest{UserID: "alice", GroupID: "g1"},
			setupMock: func(m *usergroup_mock.MockiUserGroupRepo) {
				m.EXPECT().GetUserGroup(
					"alice",
					"g1",
				).Return(aclmodel.UserGroup{}, nil)
			},
			wantErr: true,
		},
		{
			name: "repo create user group error",
			req:  AssignUserToGroupRequest{UserID: "bob", GroupID: "g2"},
			setupMock: func(m *usergroup_mock.MockiUserGroupRepo) {
				m.EXPECT().GetUserGroup(
					"bob",
					"g2",
				).Return(aclmodel.UserGroup{}, errors.New("not found"))
				m.EXPECT().CreateUserGroup(
					gomock.Any(),
				).Return(errors.New("fail"))
			},
			wantErr: true,
		},
		{
			name: "success",
			req:  AssignUserToGroupRequest{UserID: "carol", GroupID: "g3"},
			setupMock: func(m *usergroup_mock.MockiUserGroupRepo) {
				m.EXPECT().GetUserGroup(
					"carol",
					"g3",
				).Return(aclmodel.UserGroup{}, errors.New("not found"))
				m.EXPECT().CreateUserGroup(
					gomock.Any(),
				).Return(nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := usergroup_mock.NewMockiUserGroupRepo(ctrl)
			tt.setupMock(mockRepo)
			uc := AssignUserToGroupUsecase{repo: mockRepo}
			_, err := uc.Handle(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
