package acl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/usecase/acl/auth/mock"
	"github.com/stretchr/testify/assert"
)

func TestResetPasswordUsecase_HandleGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPRepo := mock.NewMockIResetPasswordRepo(ctrl)
	uc := ResetPasswordUsecase{rpRepo: mockRPRepo, userRepo: nil}

	tests := []struct {
		name    string
		req     map[string]string
		setup   func()
		want    ResetPasswordGetResponse
		wantErr bool
	}{
		{
			name:    "missing token",
			req:     map[string]string{},
			setup:   func() {},
			want:    ResetPasswordGetResponse{Error: "Token is required"},
			wantErr: false,
		},
		{
			name: "invalid token",
			req:  map[string]string{"token": "badtoken"},
			setup: func() {
				mockRPRepo.EXPECT().GetResetPasswordByToken("badtoken").Return(acl.ResetPassword{}, errors.New("not found"))
			},
			want:    ResetPasswordGetResponse{Error: "Invalid or expired token"},
			wantErr: false,
		},
		{
			name: "success",
			req:  map[string]string{"token": "goodtoken"},
			setup: func() {
				mockRPRepo.EXPECT().GetResetPasswordByToken("goodtoken").Return(acl.ResetPassword{Username: "user1"}, nil)
			},
			want:    ResetPasswordGetResponse{Username: "user1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := uc.HandleGet(context.Background(), tt.req)
			assert.Equal(t, tt.want, resp)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResetPasswordUsecase_HandlePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRPRepo := mock.NewMockIResetPasswordRepo(ctrl)
	mockUserRepo := mock.NewMockIUserRepo(ctrl)
	uc := ResetPasswordUsecase{rpRepo: mockRPRepo, userRepo: mockUserRepo}

	now := time.Now().Unix()

	tests := []struct {
		name    string
		req     ResetPasswordRequest
		setup   func()
		want    ResetPasswordResponse
		wantErr bool
	}{
		{
			name:    "missing password fields",
			req:     ResetPasswordRequest{Token: "t1", NewPassword: "", ConfirmPassword: ""},
			setup:   func() {},
			want:    ResetPasswordResponse{Success: false, Error: "Password fields required"},
			wantErr: false,
		},
		{
			name:    "passwords do not match",
			req:     ResetPasswordRequest{Token: "t1", NewPassword: "abc12345", ConfirmPassword: "abc123456"},
			setup:   func() {},
			want:    ResetPasswordResponse{Success: false, Error: "Passwords do not match"},
			wantErr: false,
		},
		{
			name:    "password too short",
			req:     ResetPasswordRequest{Token: "t1", NewPassword: "short", ConfirmPassword: "short"},
			setup:   func() {},
			want:    ResetPasswordResponse{Success: false, Error: "Password too short (min 8)"},
			wantErr: false,
		},
		{
			name: "invalid token",
			req:  ResetPasswordRequest{Token: "badtoken", NewPassword: "password123", ConfirmPassword: "password123"},
			setup: func() {
				mockRPRepo.EXPECT().GetResetPasswordByToken("badtoken").Return(acl.ResetPassword{}, errors.New("not found"))
			},
			want:    ResetPasswordResponse{Success: false, Error: "Invalid or expired token"},
			wantErr: false,
		},
		{
			name: "token expired",
			req:  ResetPasswordRequest{Token: "expiredtoken", NewPassword: "password123", ConfirmPassword: "password123"},
			setup: func() {
				mockRPRepo.EXPECT().GetResetPasswordByToken("expiredtoken").Return(acl.ResetPassword{Username: "user1", ExpiresAt: now - 100}, nil)
			},
			want:    ResetPasswordResponse{Success: false, Error: "Token expired"},
			wantErr: false,
		},
		{
			name: "user not found",
			req:  ResetPasswordRequest{Token: "goodtoken", NewPassword: "password123", ConfirmPassword: "password123"},
			setup: func() {
				mockRPRepo.EXPECT().GetResetPasswordByToken("goodtoken").Return(acl.ResetPassword{Username: "user1", ExpiresAt: now + 100}, nil)
				mockUserRepo.EXPECT().GetUserByUsername("user1").Return(acl.User{}, errors.New("not found"))
			},
			want:    ResetPasswordResponse{Success: false, Error: "User not found"},
			wantErr: false,
		},
		{
			name: "failed to update password",
			req:  ResetPasswordRequest{Token: "goodtoken", NewPassword: "password123", ConfirmPassword: "password123"},
			setup: func() {
				mockRPRepo.EXPECT().GetResetPasswordByToken("goodtoken").Return(acl.ResetPassword{Username: "user1", ExpiresAt: now + 100}, nil)
				mockUserRepo.EXPECT().GetUserByUsername("user1").Return(acl.User{Username: "user1"}, nil)
				mockUserRepo.EXPECT().UpdateUser(gomock.Any()).DoAndReturn(func(user acl.User) error {
					assert.Equal(t, "user1", user.Username)
					hash := sha256.Sum256([]byte("password123"))
					expectedPassword := hex.EncodeToString(hash[:])
					assert.Equal(t, expectedPassword, user.Password)
					assert.Equal(t, acl.StatusUserActive, user.Status)
					assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Second)
					return errors.New("db error")
				})
			},
			want:    ResetPasswordResponse{Success: false, Error: "Failed to update password"},
			wantErr: false,
		},
		{
			name: "success",
			req:  ResetPasswordRequest{Token: "goodtoken", NewPassword: "password123", ConfirmPassword: "password123"},
			setup: func() {
				mockRPRepo.EXPECT().GetResetPasswordByToken("goodtoken").Return(acl.ResetPassword{Username: "user1", ExpiresAt: now + 100}, nil)
				mockUserRepo.EXPECT().GetUserByUsername("user1").Return(acl.User{Username: "user1"}, nil)
				mockUserRepo.EXPECT().UpdateUser(gomock.Any()).DoAndReturn(func(user acl.User) error {
					assert.Equal(t, "user1", user.Username)
					hash := sha256.Sum256([]byte("password123"))
					expectedPassword := hex.EncodeToString(hash[:])
					assert.Equal(t, expectedPassword, user.Password)
					assert.Equal(t, acl.StatusUserActive, user.Status)
					assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Second)
					return nil
				})
				mockRPRepo.EXPECT().DeleteResetPasswordByToken("goodtoken").Return(nil)
			},
			want:    ResetPasswordResponse{Success: true, Redirect: "/login"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := uc.HandlePost(context.Background(), tt.req)
			assert.Equal(t, tt.want, resp)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
