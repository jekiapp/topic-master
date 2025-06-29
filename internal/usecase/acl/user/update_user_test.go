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

func TestUpdateUserUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validGroups := []groupInput{{GroupID: "g1", Role: "member", Name: "dev"}}
	rootGroup := []groupInput{{GroupID: "g2", Role: "admin", Name: acl.GroupRoot}}
	multiGroups := []groupInput{
		{GroupID: "g1", Role: "member", Name: acl.GroupRoot},
		{GroupID: "g2", Role: "admin", Name: "dev"},
	}

	tests := []struct {
		name      string
		req       UpdateUserRequest
		setupMock func(m *user_mock.MockiUpdateUserRepo)
		wantErr   bool
	}{
		{
			name:      "missing username",
			req:       UpdateUserRequest{Name: "Alice", Groups: validGroups},
			setupMock: func(m *user_mock.MockiUpdateUserRepo) {},
			wantErr:   true,
		},
		{
			name:      "missing name",
			req:       UpdateUserRequest{Username: "alice", Groups: validGroups},
			setupMock: func(m *user_mock.MockiUpdateUserRepo) {},
			wantErr:   true,
		},
		{
			name:      "no groups",
			req:       UpdateUserRequest{Username: "alice", Name: "Alice", Groups: nil},
			setupMock: func(m *user_mock.MockiUpdateUserRepo) {},
			wantErr:   true,
		},
		{
			name:      "group missing group_id",
			req:       UpdateUserRequest{Username: "alice", Name: "Alice", Groups: []groupInput{{GroupID: "", Role: "member", Name: "dev"}}},
			setupMock: func(m *user_mock.MockiUpdateUserRepo) {},
			wantErr:   true,
		},
		{
			name:      "group missing role",
			req:       UpdateUserRequest{Username: "alice", Name: "Alice", Groups: []groupInput{{GroupID: "g1", Role: "", Name: "dev"}}},
			setupMock: func(m *user_mock.MockiUpdateUserRepo) {},
			wantErr:   true,
		},
		{
			name:      "multiple groups with root",
			req:       UpdateUserRequest{Username: "alice", Name: "Alice", Groups: multiGroups},
			setupMock: func(m *user_mock.MockiUpdateUserRepo) {},
			wantErr:   true,
		},
		{
			name: "user not found",
			req:  UpdateUserRequest{Username: "alice", Name: "Alice", Groups: validGroups},
			setupMock: func(m *user_mock.MockiUpdateUserRepo) {
				m.EXPECT().GetUserByUsername(
					"alice",
				).Return(acl.User{}, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "update user error",
			req:  UpdateUserRequest{Username: "bob", Name: "Bob", Groups: validGroups},
			setupMock: func(m *user_mock.MockiUpdateUserRepo) {
				m.EXPECT().GetUserByUsername(
					"bob",
				).Return(acl.User{ID: "u2", Username: "bob", Name: "Bob"}, nil)
				m.EXPECT().UpdateUser(
					gomock.Any(),
				).Return(errors.New("fail"))
			},
			wantErr: true,
		},
		{
			name: "delete user groups error",
			req:  UpdateUserRequest{Username: "carol", Name: "Carol", Groups: validGroups},
			setupMock: func(m *user_mock.MockiUpdateUserRepo) {
				m.EXPECT().GetUserByUsername(
					"carol",
				).Return(acl.User{ID: "u3", Username: "carol", Name: "Carol"}, nil)
				m.EXPECT().UpdateUser(
					gomock.Any(),
				).Return(nil)
				m.EXPECT().DeleteUserGroupsByUserID(
					"u3",
				).Return(errors.New("fail"))
			},
			wantErr: true,
		},
		{
			name: "create user group error",
			req:  UpdateUserRequest{Username: "dave", Name: "Dave", Groups: validGroups},
			setupMock: func(m *user_mock.MockiUpdateUserRepo) {
				m.EXPECT().GetUserByUsername(
					"dave",
				).Return(acl.User{ID: "u4", Username: "dave", Name: "Dave"}, nil)
				m.EXPECT().UpdateUser(
					gomock.Any(),
				).Return(nil)
				m.EXPECT().DeleteUserGroupsByUserID(
					"u4",
				).Return(nil)
				m.EXPECT().CreateUserGroup(
					gomock.Any(),
				).Return(errors.New("fail"))
			},
			wantErr: true,
		},
		{
			name: "success with reset password",
			req:  UpdateUserRequest{Username: "eve", Name: "Eve", ResetPassword: true, Groups: validGroups},
			setupMock: func(m *user_mock.MockiUpdateUserRepo) {
				m.EXPECT().GetUserByUsername(
					"eve",
				).Return(acl.User{ID: "u5", Username: "eve", Name: "Eve"}, nil)
				m.EXPECT().UpdateUser(
					gomock.Any(),
				).Return(nil)
				m.EXPECT().DeleteUserGroupsByUserID(
					"u5",
				).Return(nil)
				m.EXPECT().CreateUserGroup(
					gomock.Any(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success with only root group",
			req:  UpdateUserRequest{Username: "frank", Name: "Frank", Groups: rootGroup},
			setupMock: func(m *user_mock.MockiUpdateUserRepo) {
				m.EXPECT().GetUserByUsername(
					"frank",
				).Return(acl.User{ID: "u6", Username: "frank", Name: "Frank"}, nil)
				m.EXPECT().UpdateUser(
					gomock.Any(),
				).Return(nil)
				m.EXPECT().DeleteUserGroupsByUserID(
					"u6",
				).Return(nil)
				m.EXPECT().CreateUserGroup(
					gomock.Any(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success with multiple groups (no root)",
			req:  UpdateUserRequest{Username: "grace", Name: "Grace", Groups: []groupInput{{GroupID: "g1", Role: "member", Name: "dev"}, {GroupID: "g2", Role: "admin", Name: "ops"}}},
			setupMock: func(m *user_mock.MockiUpdateUserRepo) {
				m.EXPECT().GetUserByUsername(
					"grace",
				).Return(acl.User{ID: "u7", Username: "grace", Name: "Grace"}, nil)
				m.EXPECT().UpdateUser(
					gomock.Any(),
				).Return(nil)
				m.EXPECT().DeleteUserGroupsByUserID(
					"u7",
				).Return(nil)
				m.EXPECT().CreateUserGroup(
					gomock.Any(),
				).Return(nil).Times(2)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := user_mock.NewMockiUpdateUserRepo(ctrl)
			tt.setupMock(mockRepo)
			uc := UpdateUserUsecase{repo: mockRepo}
			_, err := uc.Handle(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
