package acl

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jekiapp/topic-master/internal/config"
	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/usecase/acl/auth/mock"
	"github.com/stretchr/testify/assert"
)

func TestLoginUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIUserLoginRepo(ctrl)
	uc := LoginUsecase{repo: mockRepo, config: &config.Config{SecretKey: []byte("c2VjcmV0a2V5MTIzNDU2")}} // base64 for 'secretkey123456'

	hash := func(pw string) string {
		h := sha256.Sum256([]byte(pw))
		return hex.EncodeToString(h[:])
	}

	tests := []struct {
		name     string
		method   string
		body     LoginRequest
		setup    func()
		wantCode int
		wantKey  string
		wantVal  interface{}
	}{
		{
			name:     "method not allowed",
			method:   http.MethodGet,
			body:     LoginRequest{},
			setup:    func() {},
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:   "user not found",
			method: http.MethodPost,
			body:   LoginRequest{Username: "nouser", Password: "pw"},
			setup: func() {
				mockRepo.EXPECT().GetUserByUsername("nouser").Return(acl.User{}, errors.New("not found"))
			},
			wantCode: http.StatusUnauthorized,
			wantKey:  "error",
			wantVal:  "user not found",
		},
		{
			name:   "invalid password",
			method: http.MethodPost,
			body:   LoginRequest{Username: "user1", Password: "wrong"},
			setup: func() {
				mockRepo.EXPECT().GetUserByUsername("user1").Return(acl.User{Username: "user1", Password: hash("pw")}, nil)
			},
			wantCode: http.StatusUnauthorized,
			wantKey:  "error",
			wantVal:  "invalid password",
		},
		{
			name:   "pending user triggers reset",
			method: http.MethodPost,
			body:   LoginRequest{Username: "pending", Password: "pw"},
			setup: func() {
				mockRepo.EXPECT().GetUserByUsername("pending").Return(acl.User{Username: "pending", Password: hash("pw"), Status: acl.StatusUserPending}, nil)
				mockRepo.EXPECT().InsertResetPassword(gomock.Any()).Return(nil)
			},
			wantCode: http.StatusOK,
			wantKey:  "redirect",
		},
		{
			name:   "success login",
			method: http.MethodPost,
			body:   LoginRequest{Username: "user2", Password: "pw2"},
			setup: func() {
				mockRepo.EXPECT().GetUserByUsername("user2").Return(acl.User{ID: "id2", Username: "user2", Password: hash("pw2")}, nil)
				mockRepo.EXPECT().ListGroupsForUser("id2").Return([]acl.GroupRole{}, nil)
			},
			wantCode: http.StatusOK,
			wantKey:  "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, "/login", bytes.NewReader(b))
			rw := httptest.NewRecorder()
			uc.Handle(rw, req)
			resp := rw.Result()
			assert.Equal(t, tt.wantCode, resp.StatusCode)
			if tt.wantKey != "" {
				var m map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&m)
				assert.Contains(t, m, tt.wantKey)
				if tt.wantVal != nil {
					assert.Equal(t, tt.wantVal, m[tt.wantKey])
				}
			}
		})
	}
}
