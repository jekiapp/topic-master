package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	acl "github.com/jekiapp/topic-master/internal/model/acl"
	user_mock "github.com/jekiapp/topic-master/internal/usecase/acl/user/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func hash(pw string) string {
	h := sha256.Sum256([]byte(pw))
	return hex.EncodeToString(h[:])
}

func TestChangePasswordUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		req       ChangePasswordRequest
		setupMock func(m *user_mock.MockiUserPasswordRepo)
		wantErr   bool
		wantOK    bool
	}{
		{
			name:      "missing fields",
			req:       ChangePasswordRequest{UserID: "", OldPassword: "", NewPassword: ""},
			setupMock: func(m *user_mock.MockiUserPasswordRepo) {},
			wantErr:   true,
		},
		{
			name: "user not found",
			req:  ChangePasswordRequest{UserID: "alice", OldPassword: "old", NewPassword: "new"},
			setupMock: func(m *user_mock.MockiUserPasswordRepo) {
				m.EXPECT().GetUserByID(
					"alice",
				).Return(acl.User{}, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "invalid old password",
			req:  ChangePasswordRequest{UserID: "bob", OldPassword: "wrong", NewPassword: "new"},
			setupMock: func(m *user_mock.MockiUserPasswordRepo) {
				m.EXPECT().GetUserByID(
					"bob",
				).Return(acl.User{Password: hash("right")}, nil)
			},
			wantErr: true,
		},
		{
			name: "update user error",
			req:  ChangePasswordRequest{UserID: "carol", OldPassword: "old", NewPassword: "new"},
			setupMock: func(m *user_mock.MockiUserPasswordRepo) {
				m.EXPECT().GetUserByID(
					"carol",
				).Return(acl.User{Password: hash("old")}, nil)
				m.EXPECT().UpdateUser(
					gomock.Any(),
				).Return(errors.New("fail"))
			},
			wantErr: true,
		},
		{
			name: "success",
			req:  ChangePasswordRequest{UserID: "dave", OldPassword: "old", NewPassword: "new"},
			setupMock: func(m *user_mock.MockiUserPasswordRepo) {
				m.EXPECT().GetUserByID(
					"dave",
				).Return(acl.User{Password: hash("old")}, nil)
				m.EXPECT().UpdateUser(
					gomock.Any(),
				).Return(nil)
			},
			wantOK: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := user_mock.NewMockiUserPasswordRepo(ctrl)
			tt.setupMock(mockRepo)
			uc := ChangePasswordUsecase{repo: mockRepo}
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
