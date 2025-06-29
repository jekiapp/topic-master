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

func TestDeleteUserUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		req       DeleteUserRequest
		setupMock func(m *user_mock.MockiUserDeleteRepo)
		wantErr   bool
		wantOK    bool
	}{
		{
			name: "delete user error",
			req:  DeleteUserRequest{UserID: "alice"},
			setupMock: func(m *user_mock.MockiUserDeleteRepo) {
				m.EXPECT().DeleteUser("alice").Return(errors.New("fail"))
			},
			wantErr: true,
		},
		{
			name: "list user groups error",
			req:  DeleteUserRequest{UserID: "bob"},
			setupMock: func(m *user_mock.MockiUserDeleteRepo) {
				m.EXPECT().DeleteUser("bob").Return(nil)
				m.EXPECT().ListUserGroupsByUserID("bob").Return(nil, errors.New("fail"))
			},
			wantErr: true,
		},
		{
			name: "delete user group error",
			req:  DeleteUserRequest{UserID: "carol"},
			setupMock: func(m *user_mock.MockiUserDeleteRepo) {
				m.EXPECT().DeleteUser("carol").Return(nil)
				m.EXPECT().ListUserGroupsByUserID("carol").Return([]acl.UserGroup{{ID: "ug1"}}, nil)
				m.EXPECT().DeleteUserGroup("ug1").Return(errors.New("fail"))
			},
			wantErr: true,
		},
		{
			name: "success",
			req:  DeleteUserRequest{UserID: "dave"},
			setupMock: func(m *user_mock.MockiUserDeleteRepo) {
				m.EXPECT().DeleteUser("dave").Return(nil)
				m.EXPECT().ListUserGroupsByUserID("dave").Return([]acl.UserGroup{{ID: "ug2"}}, nil)
				m.EXPECT().DeleteUserGroup("ug2").Return(nil)
			},
			wantOK: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := user_mock.NewMockiUserDeleteRepo(ctrl)
			tt.setupMock(mockRepo)
			uc := DeleteUserUsecase{repo: mockRepo}
			resp, err := uc.Handle(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, resp.Success)
			}
		})
	}
}
