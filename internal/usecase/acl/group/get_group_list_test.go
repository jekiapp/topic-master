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
			name: "repo error",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return(nil, context.DeadlineExceeded)
			},
			wantErr: true,
		},
		{
			name: "success with members",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return([]acl.Group{{ID: "dev-group", Name: "Dev Group", Description: "Development group"}}, nil)
				mu.EXPECT().ListUserGroupsByGroupID("dev-group", 3).Return([]acl.UserGroup{{UserID: "john"}}, nil)
				mu.EXPECT().GetUserByID("john").Return(acl.User{Username: "john", Status: acl.StatusUserActive}, nil)
			},
			wantLen:  1,
			wantUser: "john",
		},
		{
			name: "empty group list",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return(nil, nil)
			},
			wantLen: 0,
		},
		{
			name: "user repo returns error",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return([]acl.Group{{ID: "qa-team", Name: "QA Team", Description: "QA group"}}, nil)
				mu.EXPECT().ListUserGroupsByGroupID("qa-team", 3).Return(nil, context.DeadlineExceeded)
			},
			wantLen: 1,
		},
		{
			name: "group with no active users",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return([]acl.Group{{ID: "ops-team", Name: "Ops Team", Description: "Operations group"}}, nil)
				mu.EXPECT().ListUserGroupsByGroupID("ops-team", 3).Return([]acl.UserGroup{{UserID: "alice"}}, nil)
				mu.EXPECT().GetUserByID("alice").Return(acl.User{Username: "alice", Status: acl.StatusUserInactive}, nil)
			},
			wantLen: 1,
		},
		{
			name: "group with multiple users, some inactive",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return([]acl.Group{{ID: "hr-team", Name: "HR Team", Description: "HR group"}}, nil)
				mu.EXPECT().ListUserGroupsByGroupID("hr-team", 3).Return([]acl.UserGroup{{UserID: "bob"}, {UserID: "carol"}}, nil)
				mu.EXPECT().GetUserByID("bob").Return(acl.User{Username: "bob", Status: acl.StatusUserActive}, nil)
				mu.EXPECT().GetUserByID("carol").Return(acl.User{Username: "carol", Status: acl.StatusUserInactive}, nil)
			},
			wantLen: 1,
		},
		{
			name: "multiple groups",
			setupMock: func(mg *group_mock.MockiGroupListRepo, mu *group_mock.MockiUserGroupRepo) {
				mg.EXPECT().GetAllGroups().Return([]acl.Group{{ID: "dev-group", Name: "Dev Group", Description: "Development group"}, {ID: "qa-team", Name: "QA Team", Description: "QA group"}}, nil)
				mu.EXPECT().ListUserGroupsByGroupID("dev-group", 3).Return([]acl.UserGroup{{UserID: "john"}}, nil)
				mu.EXPECT().GetUserByID("john").Return(acl.User{Username: "john", Status: acl.StatusUserActive}, nil)
				mu.EXPECT().ListUserGroupsByGroupID("qa-team", 3).Return([]acl.UserGroup{{UserID: "john"}}, nil)
				mu.EXPECT().GetUserByID("john").Return(acl.User{Username: "john", Status: acl.StatusUserActive}, nil)
			},
			wantLen:  2,
			wantUser: "john",
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
