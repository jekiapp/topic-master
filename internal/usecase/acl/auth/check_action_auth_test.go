package acl

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jekiapp/topic-master/internal/model"
	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/jekiapp/topic-master/internal/usecase/acl/auth/mock"
	"github.com/stretchr/testify/assert"
)

func TestCheckActionAuthUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockICheckUserActionPermission(ctrl)
	uc := CheckActionAuthUsecase{repo: mockRepo}

	tests := []struct {
		name    string
		ctx     context.Context
		req     CheckActionAuthRequest
		setup   func()
		want    CheckActionAuthResponse
		wantErr bool
	}{
		{
			name:    "user not found",
			ctx:     context.Background(),
			req:     CheckActionAuthRequest{EntityID: "e1", Action: "read"},
			setup:   func() {},
			want:    CheckActionAuthResponse{Allowed: false, Error: "user not found"},
			wantErr: false,
		},
		{
			name: "permission denied",
			ctx:  context.WithValue(context.Background(), model.UserInfoKey, &acl.JWTClaims{UserID: "u1"}),
			req:  CheckActionAuthRequest{EntityID: "e1", Action: "write"},
			setup: func() {
				mockRepo.EXPECT().GetEntityByID("e1").Return(&entity.Entity{}, nil)
				mockRepo.EXPECT().GetGroupsByUserID("u1").Return([]acl.GroupRole{}, nil)
				mockRepo.EXPECT().GetPermissionByActionEntity("u1", "e1", "write").Return(acl.PermissionMap{}, errors.New("denied"))
			},
			want:    CheckActionAuthResponse{Allowed: false, Error: "permission denied"},
			wantErr: false,
		},
		{
			name: "permission allowed",
			ctx:  context.WithValue(context.Background(), model.UserInfoKey, &acl.JWTClaims{UserID: "u2"}),
			req:  CheckActionAuthRequest{EntityID: "e2", Action: "read"},
			setup: func() {
				mockRepo.EXPECT().GetEntityByID("e2").Return(&entity.Entity{GroupOwner: ""}, nil)
			},
			want:    CheckActionAuthResponse{Allowed: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := uc.Handle(tt.ctx, tt.req)
			assert.Equal(t, tt.want, resp)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
