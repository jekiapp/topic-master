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

type groupInput = struct {
	GroupID string `json:"group_id"`
	Role    string `json:"role"`
	Name    string `json:"name"`
}

func TestCreateUserUsecase_Handle(t *testing.T) {
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
		req       CreateUserRequest
		setupMock func(m *user_mock.MockiUserRepo)
		patchRand bool
		wantErr   bool
		wantResp  CreateUserResponse
	}{
		{
			name:      "missing username",
			req:       CreateUserRequest{Name: "Alice", Password: "pass", Groups: validGroups},
			setupMock: func(m *user_mock.MockiUserRepo) {},
			wantErr:   true,
		},
		{
			name:      "missing name",
			req:       CreateUserRequest{Username: "alice", Password: "pass", Groups: validGroups},
			setupMock: func(m *user_mock.MockiUserRepo) {},
			wantErr:   true,
		},
		{
			name:      "no groups",
			req:       CreateUserRequest{Username: "alice", Name: "Alice", Password: "pass", Groups: nil},
			setupMock: func(m *user_mock.MockiUserRepo) {},
			wantErr:   true,
		},
		{
			name:      "group missing group_id",
			req:       CreateUserRequest{Username: "alice", Name: "Alice", Password: "pass", Groups: []groupInput{{GroupID: "", Role: "member", Name: "dev"}}},
			setupMock: func(m *user_mock.MockiUserRepo) {},
			wantErr:   true,
		},
		{
			name:      "group missing role",
			req:       CreateUserRequest{Username: "alice", Name: "Alice", Password: "pass", Groups: []groupInput{{GroupID: "g1", Role: "", Name: "dev"}}},
			setupMock: func(m *user_mock.MockiUserRepo) {},
			wantErr:   true,
		},
		{
			name:      "multiple groups with root",
			req:       CreateUserRequest{Username: "alice", Name: "Alice", Password: "pass", Groups: multiGroups},
			setupMock: func(m *user_mock.MockiUserRepo) {},
			wantErr:   true,
		},
		{
			name: "user already exists",
			req:  CreateUserRequest{Username: "alice", Name: "Alice", Password: "pass", Groups: validGroups},
			setupMock: func(m *user_mock.MockiUserRepo) {
				m.EXPECT().GetUserByUsername(
					"alice",
				).Return(acl.User{}, nil)
			},
			wantErr: true,
		},
		{
			name: "repo create user error",
			req:  CreateUserRequest{Username: "bob", Name: "Bob", Password: "pass", Groups: validGroups},
			setupMock: func(m *user_mock.MockiUserRepo) {
				m.EXPECT().GetUserByUsername(
					"bob",
				).Return(acl.User{}, errors.New("not found"))
				m.EXPECT().CreateUser(
					gomock.Any(),
				).Return(errors.New("fail"))
			},
			wantErr: true,
		},
		{
			name: "repo create user group error",
			req:  CreateUserRequest{Username: "carol", Name: "Carol", Password: "pass", Groups: validGroups},
			setupMock: func(m *user_mock.MockiUserRepo) {
				m.EXPECT().GetUserByUsername(
					"carol",
				).Return(acl.User{}, errors.New("not found"))
				m.EXPECT().CreateUser(
					gomock.Any(),
				).Return(nil)
				m.EXPECT().CreateUserGroup(
					gomock.Any(),
				).Return(errors.New("fail"))
			},
			wantErr: true,
		},
		{
			name: "success with password provided",
			req:  CreateUserRequest{Username: "bob", Name: "Bob", Password: "pass", Groups: validGroups},
			setupMock: func(m *user_mock.MockiUserRepo) {
				m.EXPECT().GetUserByUsername(
					"bob",
				).Return(acl.User{}, errors.New("not found"))
				m.EXPECT().CreateUser(
					gomock.Any(),
				).Return(nil)
				m.EXPECT().CreateUserGroup(
					gomock.Any(),
				).Return(nil)
			},
			wantErr:  false,
			wantResp: CreateUserResponse{Username: "bob"},
		},
		{
			name: "success with random password",
			req:  CreateUserRequest{Username: "dave", Name: "Dave", Password: "", Groups: validGroups},
			setupMock: func(m *user_mock.MockiUserRepo) {
				m.EXPECT().GetUserByUsername(
					"dave",
				).Return(acl.User{}, errors.New("not found"))
				m.EXPECT().CreateUser(
					gomock.Any(),
				).Return(nil)
				m.EXPECT().CreateUserGroup(
					gomock.Any(),
				).Return(nil)
			},
			wantErr:  false,
			wantResp: CreateUserResponse{Username: "dave"},
		},
		{
			name: "success with only root group",
			req:  CreateUserRequest{Username: "eve", Name: "Eve", Password: "pass", Groups: rootGroup},
			setupMock: func(m *user_mock.MockiUserRepo) {
				m.EXPECT().GetUserByUsername(
					"eve",
				).Return(acl.User{}, errors.New("not found"))
				m.EXPECT().CreateUser(
					gomock.Any(),
				).Return(nil)
				m.EXPECT().CreateUserGroup(
					gomock.Any(),
				).Return(nil)
			},
			wantErr:  false,
			wantResp: CreateUserResponse{Username: "eve"},
		},
		{
			name: "success with multiple groups (no root)",
			req:  CreateUserRequest{Username: "frank", Name: "Frank", Password: "pass", Groups: []groupInput{{GroupID: "g1", Role: "member", Name: "dev"}, {GroupID: "g2", Role: "admin", Name: "ops"}}},
			setupMock: func(m *user_mock.MockiUserRepo) {
				m.EXPECT().GetUserByUsername(
					"frank",
				).Return(acl.User{}, errors.New("not found"))
				m.EXPECT().CreateUser(
					gomock.Any(),
				).Return(nil)
				m.EXPECT().CreateUserGroup(
					gomock.Any(),
				).Return(nil).Times(2)
			},
			wantErr:  false,
			wantResp: CreateUserResponse{Username: "frank"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := user_mock.NewMockiUserRepo(ctrl)
			tt.setupMock(mockRepo)
			uc := CreateUserUsecase{repo: mockRepo}
			resp, err := uc.Handle(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp.Username, resp.Username)
			}
		})
	}
}
