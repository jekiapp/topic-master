package user

import (
	"context"
	"errors"
	"testing"

	acl "github.com/jekiapp/topic-master/internal/model/acl"
	user_mock "github.com/jekiapp/topic-master/internal/usecase/acl/user/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetUserListUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		setupMock func(m *user_mock.MockiUserDataRepo)
		wantErr   bool
		wantLen   int
	}{
		{
			name: "repo error",
			setupMock: func(m *user_mock.MockiUserDataRepo) {
				m.EXPECT().GetAllUsers().Return(nil, errors.New("fail"))
			},
			wantErr: true,
		},
		{
			name: "success single user, single group",
			setupMock: func(m *user_mock.MockiUserDataRepo) {
				m.EXPECT().GetAllUsers().Return([]acl.User{{ID: "u1", Username: "alice", Name: "Alice"}}, nil)
				m.EXPECT().ListUserGroupsByUserID(
					"u1",
				).Return([]acl.UserGroup{{GroupID: "g1", Role: "admin"}}, nil)
				m.EXPECT().GetGroupByID(
					"g1",
				).Return(acl.Group{Name: "dev"}, nil)
			},
			wantLen: 1,
		},
		{
			name: "user group error",
			setupMock: func(m *user_mock.MockiUserDataRepo) {
				m.EXPECT().GetAllUsers().Return([]acl.User{{ID: "u2", Username: "bob", Name: "Bob"}}, nil)
				m.EXPECT().ListUserGroupsByUserID(
					"u2",
				).Return(nil, errors.New("fail"))
			},
			wantLen: 1,
		},
		{
			name: "user with no groups",
			setupMock: func(m *user_mock.MockiUserDataRepo) {
				m.EXPECT().GetAllUsers().Return([]acl.User{{ID: "u3", Username: "carol", Name: "Carol"}}, nil)
				m.EXPECT().ListUserGroupsByUserID(
					"u3",
				).Return([]acl.UserGroup{}, nil)
			},
			wantLen: 1,
		},
		{
			name: "user with multiple groups, one group lookup error",
			setupMock: func(m *user_mock.MockiUserDataRepo) {
				m.EXPECT().GetAllUsers().Return([]acl.User{{ID: "u4", Username: "dave", Name: "Dave"}}, nil)
				m.EXPECT().ListUserGroupsByUserID(
					"u4",
				).Return([]acl.UserGroup{{GroupID: "g1", Role: "admin"}, {GroupID: "g2", Role: "member"}}, nil)
				m.EXPECT().GetGroupByID(
					"g1",
				).Return(acl.Group{Name: "dev"}, nil)
				m.EXPECT().GetGroupByID(
					"g2",
				).Return(acl.Group{}, errors.New("fail"))
			},
			wantLen: 1,
		},
		{
			name: "multiple users, mixed groups",
			setupMock: func(m *user_mock.MockiUserDataRepo) {
				m.EXPECT().GetAllUsers().Return([]acl.User{
					{ID: "u5", Username: "eve", Name: "Eve"},
					{ID: "u6", Username: "frank", Name: "Frank"},
				}, nil)
				m.EXPECT().ListUserGroupsByUserID(
					"u5",
				).Return([]acl.UserGroup{{GroupID: "g1", Role: "admin"}}, nil)
				m.EXPECT().GetGroupByID(
					"g1",
				).Return(acl.Group{Name: "dev"}, nil)
				m.EXPECT().ListUserGroupsByUserID(
					"u6",
				).Return([]acl.UserGroup{}, nil)
			},
			wantLen: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := user_mock.NewMockiUserDataRepo(ctrl)
			tt.setupMock(mockRepo)
			uc := GetUserListUsecase{dataRepo: mockRepo}
			resp, err := uc.Handle(context.Background(), GetUserListRequest{})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLen, len(resp.Users))
			}
		})
	}
}
