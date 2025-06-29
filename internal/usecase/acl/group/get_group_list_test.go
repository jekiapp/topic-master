package group

import (
	"context"
	"testing"

	"github.com/jekiapp/topic-master/internal/model/acl"
	group_mock "github.com/jekiapp/topic-master/internal/usecase/acl/group/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetGroupListUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		setupMock func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo)
		wantErr   bool
		wantLen   int
		wantUser  string
	}{
		{
			name: "repo returns error for group list",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return(
					nil,
					context.DeadlineExceeded,
				)
			},
			wantErr: true,
		},
		{
			name: "success with alice as member",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return(
					[]acl.Group{{ID: "alice-group", Name: "Alice Group", Description: "Engineering group"}},
					nil,
				)
				mu.EXPECT().ListUserGroupsByGroupID(
					"alice-group",
					3,
				).Return(
					[]acl.UserGroup{{UserID: "alice"}},
					nil,
				)
				mu.EXPECT().GetUserByID(
					"alice",
				).Return(
					acl.User{Username: "alice", Status: acl.StatusUserActive},
					nil,
				)
			},
			wantLen:  1,
			wantUser: "alice",
		},
		{
			name: "empty group list should return zero",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return(
					nil,
					nil,
				)
			},
			wantLen: 0,
		},
		{
			name: "user repo returns error for bob's group",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return(
					[]acl.Group{{ID: "bob-group", Name: "Bob Group", Description: "QA group"}},
					nil,
				)
				mu.EXPECT().ListUserGroupsByGroupID(
					"bob-group",
					3,
				).Return(
					nil,
					context.DeadlineExceeded,
				)
			},
			wantLen: 1,
		},
		{
			name: "group with no active users for carol's group",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return(
					[]acl.Group{{ID: "carol-group", Name: "Carol Group", Description: "Operations group"}},
					nil,
				)
				mu.EXPECT().ListUserGroupsByGroupID(
					"carol-group",
					3,
				).Return(
					[]acl.UserGroup{{UserID: "carol"}},
					nil,
				)
				mu.EXPECT().GetUserByID(
					"carol",
				).Return(
					acl.User{Username: "carol", Status: acl.StatusUserInactive},
					nil,
				)
			},
			wantLen: 1,
		},
		{
			name: "group with multiple users, some inactive (dave, eve)",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return(
					[]acl.Group{{ID: "dave-group", Name: "Dave Group", Description: "HR group"}},
					nil,
				)
				mu.EXPECT().ListUserGroupsByGroupID(
					"dave-group",
					3,
				).Return(
					[]acl.UserGroup{{UserID: "dave"}, {UserID: "eve"}},
					nil,
				)
				mu.EXPECT().GetUserByID(
					"dave",
				).Return(
					acl.User{Username: "dave", Status: acl.StatusUserActive},
					nil,
				)
				mu.EXPECT().GetUserByID(
					"eve",
				).Return(
					acl.User{Username: "eve", Status: acl.StatusUserInactive},
					nil,
				)
			},
			wantLen: 1,
		},
		{
			name: "multiple groups (alice, bob)",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return([]acl.Group{{ID: "alice-group", Name: "Alice Group", Description: "Engineering group"}, {ID: "bob-group", Name: "Bob Group", Description: "QA group"}}, nil)
				mu.EXPECT().ListUserGroupsByGroupID("alice-group", 3).Return([]acl.UserGroup{{UserID: "alice"}}, nil)
				mu.EXPECT().GetUserByID("alice").Return(acl.User{Username: "alice", Status: acl.StatusUserActive}, nil)
				mu.EXPECT().ListUserGroupsByGroupID("bob-group", 3).Return([]acl.UserGroup{{UserID: "bob"}}, nil)
				mu.EXPECT().GetUserByID("bob").Return(acl.User{Username: "bob", Status: acl.StatusUserActive}, nil)
			},
			wantLen:  2,
			wantUser: "alice",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroup := group_mock.NewMockiGroupListRepo(ctrl)
			mockUser := group_mock.NewMockiUserGroupRepo(ctrl)
			tt.setupMock(mockGroup, mockUser)
			uc := GetGroupListUsecase{groupRepo: mockGroup, userGroupRepo: mockUser}
			resp, err := uc.Handle(context.Background(), GetGroupListRequest{})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantLen, len(resp.Groups))
			if tt.wantLen > 0 && tt.wantUser != "" {
				assert.Equal(t, tt.wantUser, resp.Groups[0].Members)
			}
		})
	}
}
